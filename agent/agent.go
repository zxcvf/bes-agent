package agent

import (
	"bes-agent/py"
	"fmt"
	"github.com/sbinet/go-python"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"bes-agent/common/config"
	"bes-agent/common/log"
	"bes-agent/common/metric"
	"bes-agent/common/plugin"
)

// Agent runs agent and collects data based on the given config
type Agent struct {
	conf      *config.Config
	collector *Collector
}

// NewAgent returns an Agent struct based off the given Config
func NewAgent(conf *config.Config) *Agent {
	collector := NewCollector(conf)

	a := &Agent{
		conf:      conf,
		collector: collector,
	}

	return a
}

func panicRecover(plugin *plugin.RunningPlugin) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Infof("FATAL: Plugin [%s] panicked: %s, Stack:\n%s",
			plugin.Name, err, trace)
	}
}

func (a *Agent) collectPython(shutdown chan struct{}, rpp *plugin.RunningPythonPlugin, interval time.Duration, metricC chan metric.Metric) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	//agg := NewAggregator(metricC, a.conf)
	fmt.Println(metricC, "metricChannel?", reflect.TypeOf(metricC))

	for {
		collectPythonWithTimeout(shutdown, rpp, interval)

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func collectPythonWithTimeout(shutdown chan struct{}, rpp *plugin.RunningPythonPlugin, timeout time.Duration) {
	fmt.Println("=============RunningPythonPlugin")
	fmt.Println(rpp)
}

// collect runs the Plugins that have been configured with their own
// reporting interval.
func (a *Agent) collect(shutdown chan struct{}, rp *plugin.RunningPlugin, interval time.Duration, metricC chan metric.Metric) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	agg := NewAggregator(metricC, a.conf)

	for {
		// plugins loop
		collectWithTimeout(shutdown, rp, agg, interval)

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}

// collectWithTimeout collects from the given Plugin, with the given timeout.
//   when the given timeout is reached, and logs an error message
//   but continues waiting for it to return. This is to avoid leaving behind
//   hung processes, and to prevent re-calling the same hung process over and
//   over.
func collectWithTimeout(
	shutdown chan struct{},
	rp *plugin.RunningPlugin,
	agg metric.Aggregator,
	timeout time.Duration,
) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	done := make(chan error)

	var wg sync.WaitGroup
	for i, plug := range rp.Plugins {
		wg.Add(1)
		go func() {
			defer panicRecover(rp)
			defer wg.Done()

			done <- plug.Check(agg)
			//fmt.Println("Aggregator 冲洗 插件:", plug, reflect.TypeOf(plug))
			agg.Flush()
		}()

		select {
		case err := <-done:
			if err != nil {
				log.Errorf("ERROR to check plugin instance [%s#%d]: %s", rp.Name, i, err)
			}
		case <-ticker.C:
			log.Infof("ERROR: plugin instance [%s#%d] took longer to collect than "+
				"collection interval (%s)",
				rp.Name, i, timeout)
		case <-shutdown:
			return
		}
	}

	wg.Wait()
}

// Test verifies that we can 'collect' from all Plugins with their configured
// Config struct
func (a *Agent) Test() error {
	shutdown := make(chan struct{})
	metricC := make(chan metric.Metric)
	var metrics []string
	var checkRate bool

	// dummy receiver for the metric channel
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case m := <-metricC:
				if checkRate {
					metrics = append(metrics, m.String())
				}
			case <-shutdown:
				return
			}
		}
	}()

	agg := NewAggregator(metricC, a.conf)
	for _, rp := range a.conf.Plugins {
		fmt.Println("------------------------------------")
		for i, plug := range rp.Plugins {
			fmt.Printf("* Plugin: %s\n", rp.Name)
			if err := plug.Check(agg); err != nil {
				return err
			}
			agg.Flush()

			// Wait a second for collecting rate metrics.
			time.Sleep(time.Second)
			fmt.Println("* Running 2nd iteration to capture rate metrics")
			if err := plug.Check(agg); err != nil {
				return err
			}
			checkRate = true
			agg.Flush()

			// Waiting for the metrics filled up
			time.Sleep(time.Millisecond)

			fmt.Printf("* Instance #%d, Collected %d metrics\n", i, len(metrics))
			sort.Strings(metrics)
			for _, m := range metrics {
				fmt.Println("> " + m)
			}
			metrics = []string{}
		}
	}

	close(shutdown)
	wg.Wait()

	fmt.Println("Done!")
	return nil
}

// Agent Run runs the agent daemon, collecting every Interval
func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup
	interval := 15 * time.Second

	// channel shared between all Plugin threads for collecting metrics
	metricC := make(chan metric.Metric, 10000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.collector.Run(shutdown, metricC, interval); err != nil {
			log.Infof("Collector routine failed, exiting: %s", err.Error())
			close(shutdown)
		}
	}()

	python.Initialize()
	err := py.LoadPy() // 加载aggregator、python、相关依赖
	//defer python.Finalize()
	// 加载collector.plugins模块 再加载collector.plugins.test()等函数、类
	fmt.Println("@@@@LoadPy err:", err)

	//运行所有插件  暂时写到了支持普通函数
	//wg.Add(len(a.conf.PythonPlugins)) // config.Plugins []*plugin.RunningPlugin
	//for _, p := range a.conf.PythonPlugins{
	//	pythonRunningModule := python.PyImport_ImportModule(p.Plugin)
	//	go func(pluginModule string, pythonRunningModule *python.PyObject) {
	//		defer wg.Done()
	//	    // 在goroutine里会报错？
	//		//pythonRunningModule := python.PyImport_ImportModule(pluginModule)
	//		pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
	//	}(p.Plugin, pythonRunningModule)
	//	//fmt.Println(pythonRunningModule)
	//}

	// 简单调用脚本
	//wg.Add(len(a.conf.PythonPlugins)) // config.Plugins []*plugin.RunningPlugin
	//for _, p := range a.conf.PythonPlugins {
	//	fmt.Println("agent.go: run python plugin ", p, "")
	//	go func(pluginModule string) {
	//		defer wg.Done()
	//		pythonRunningModule := python.PyImport_ImportModule(pluginModule)
	//		if pythonRunningModule == nil {
	//			panic("pythonRunningModule is nil") //pluginModule
	//		}
	//		pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
	//		//checkFunc := pythonRunningModule.GetAttrString("check")
	//		//if checkFunc == nil {
	//		//	panic("Error importing function")
	//		//}
	//		//// The Python function takes no params but when using the C api
	//		//// we're required to send (empty) *args and **kwargs anyways.
	//		//checkFunc.Call(python.PyTuple_New(0), python.PyDict_New())
	//	}(p.Plugin)
	//}

	// load
	//wg.Add(len(a.conf.PythonPlugins))  // config.Plugins []*plugin.RunningPlugin
	//for _, p := range a.conf.PythonPlugins {
	//	fmt.Println("agent.go: run python plugin ", p, "")
	//	defer wg.Done()
	//	pythonRunningModule := python.PyImport_ImportModule(p.Plugin)
	//	if pythonRunningModule == nil {
	//		panic("pythonRunningModule is nil")  //pluginModule
	//	}
	//	pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
	//	//go func(rpp *plugin.RunningPythonPlugin, interval time.Duration){
	//	//	defer wg.Done()
	//	//	if err := a.collectPython(shutdown, rpp, interval, metricC); err != nil {
	//	//		log.Info(err.Error())
	//	//	}
	//	//}(p, interval)
	//}

	wg.Add(len(a.conf.Plugins)) // config.Plugins []*plugin.RunningPlugin
	for _, p := range a.conf.Plugins {
		fmt.Println("agent.go: run plugin ", p, "")
		go func(rp *plugin.RunningPlugin, interval time.Duration) {
			defer wg.Done()
			// aggregator collectWithTimeout
			if err := a.collect(shutdown, rp, interval, metricC); err != nil {
				log.Info(err.Error())
			}
		}(p, interval)
	}

	wg.Wait()
	return nil
}
