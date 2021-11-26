package main

import (
	"bes-agent/common/filebeat"
	"bes-agent/py"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"bes-agent/agent"
	"bes-agent/collector"
	_ "bes-agent/collector/plugins"
	"bes-agent/common/config"
	"bes-agent/common/log"
	"bes-agent/forwarder"
	"bes-agent/statsd"
)

var fConfig = flag.String("config", "", "configuration file to load")
var fTest = flag.Bool("test", false, "collect metrics, print them out, and exit")
var fPluginFilters = flag.String("plugin-filter", "",
	"filter the plugins to enable, separator is :")

const usage = `BES Agent, a system tool that monitors system processes and services.

Usage:

  bes-agent [commands|flags]

The commands & flags are:

  --config <file>     configuration file to load
  --test              collect metrics once, print them to stdout, and exit
  --plugin-filter     filter the plugins to enable, separator is :

Examples:

  # run bes-agent with all plugins defined in config file
  bes-agent --config bes-agent.conf

  # run a single collection, outputing metrics to stdout
  bes-agent --config bes-agent.conf -test

  # run bes-agent, enabling the system & disk plugins
  bes-agent --config bes-agent.conf --plugin-filter system:disk
`

func startAgent(shutdown chan struct{}, conf *config.Config, test bool) {
	//if conf
	// 通过插件名和配置文件 新增插件
	if conf.GlobalConfig.PythonPlugin {
		err := py.LoadPy()
		fmt.Println("@@@@LoadPy err:", err)
	}

	ag := agent.NewAgent(conf)
	if test {
		log.SetLevel("error")
		log.SetOutput(os.Stderr)
		if err := ag.Test(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if err := ag.Run(shutdown); err != nil {
		log.Fatal(err)
	}
}

func startForwarder(shutdown chan struct{}, conf *config.Config) {
	f := forwarder.NewForwarder(conf)
	if err := f.Run(shutdown); err != nil {
		log.Fatal(err)
	}
}

func startStatsd(shutdown chan struct{}, conf *config.Config) {
	s := statsd.NewStatsd(conf)
	if err := s.Run(shutdown); err != nil {
		log.Fatal(err)
	}
}

func startFilebeat(shutdown chan struct{}, conf *config.Config) {
	s := filebeat.NewFilebeat(conf)
	if err := s.Run(shutdown); err != nil {
		log.Fatal(err)
	}
}

func usageExit(rc int) {
	fmt.Println(usage)
	os.Exit(rc)
}

func main() {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false
		flag.Usage = func() { usageExit(0) }
		flag.Parse()

		shutdown := make(chan struct{})
		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP)
		go func() {
			select {
			case sig := <-signals:
				if sig == os.Interrupt {
					close(shutdown)
				}
				if sig == syscall.SIGHUP {
					log.Infof("Reloading config...")
					<-reload
					reload <- true
					close(shutdown)
				}
			}
		}()

		pluginFilters := []string{}
		if *fPluginFilters != "" {
			pluginFilters = strings.Split(":"+strings.TrimSpace(*fPluginFilters)+":", ":")
		}

		conf, err := config.NewConfig(*fConfig, pluginFilters)
		if err != nil {
			log.Fatalf("failed to load config: %s", err)
		}

		err = conf.InitializeLogging()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Available Plugins:")
		for k := range collector.Plugins {
			fmt.Printf("  %s\n", k)
		}

		log.Infof("Loaded plugins: %s", strings.Join(conf.PluginNames(), " "))
		fmt.Printf("Loaded plugins: %s \n", strings.Join(conf.PluginNames(), " "))

		var wg sync.WaitGroup
		//wg.Add(4)
		wg.Add(3)
		go func() {
			defer wg.Done()

			startAgent(shutdown, conf, *fTest)
		}()

		go func() {
			defer wg.Done()

			startForwarder(shutdown, conf)
		}()

		go func() {
			defer wg.Done()

			startStatsd(shutdown, conf)
		}()

		//go func() {
		//	defer wg.Done()
		//	startFilebeat(shutdown, conf)
		//}()

		wg.Wait()
	}
}
