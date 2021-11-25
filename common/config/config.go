package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bes-agent/collector"
	"bes-agent/common/log"
	"bes-agent/common/plugin"
	"bes-agent/common/util"
	"github.com/BurntSushi/toml"
)

// VERSION sets the agent version here.
const VERSION = "0.5.0"

// NewConfig creates a new instance of Config.
func NewConfig(confPath string, pluginFilters []string) (*Config, error) {
	c := &Config{}
	c.pluginFilters = pluginFilters

	err := c.setProjectPath()
	if err != nil {
		return nil, fmt.Errorf("Failed to set the project path: %s", err)
	}

	err = c.LoadConfig(confPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load the config file: %s", err)
	}

	//if c.GlobalConfig.LicenseKey == "" {
	//	return nil, fmt.Errorf("LicenseKey must be specified in the config file.")
	//}

	return c, nil
}

// Config represents bes-agent's configuration file.
type Config struct {
	GlobalConfig  GlobalConfig  `toml:"global"`
	LoggingConfig LoggingConfig `toml:"logging"`
	Plugins       []*plugin.RunningPlugin
	PythonPlugins []*plugin.RunningPythonPlugin
	pluginFilters []string
	ProjectPath   string
}

// GlobalConfig XXX
type GlobalConfig struct {
	CiURL           string `toml:"ci_url"`
	LicenseKey      string `toml:"license_key"`
	Hostname        string `toml:"hostname"`
	Tags            string `toml:"tags"`
	Proxy           string `toml:"proxy"`
	BindHost        string `toml:"bind_host"`
	ListenPort      int    `toml:"listen_port"`
	StatsdPort      int    `toml:"statsd_port"`
	NonLocalTraffic bool   `toml:"non_local_traffic"`
}

// LoggingConfig XXX
type LoggingConfig struct {
	LogLevel string `toml:"log_level"`
	LogFile  string `toml:"log_file"`
}

// Try to find a default config file at these locations (in order):
//   1. $CWD/bes-agent.conf
//   2. /etc/bes-agent/bes-agent.conf
//
func getDefaultConfigPath() (string, error) {
	file := "bes-agent.conf"
	etcfile := "/etc/bes-agent/bes-agent.conf"
	return getPath(file, etcfile)
}

// Try to find plugins path at these locations (in order):
//   1. $CONFPATH/collector/conf.d
//   2. $CONFPATH/../../../collector/conf.d  **This is just for test case.**
//   3. /etc/bes-agent/conf.d
//
func getPluginsPath(confPath string) (string, error) {
	path := filepath.Join(filepath.Dir(confPath), "collector/conf.d")
	testpath := filepath.Join(filepath.Dir(confPath), "../../../collector/conf.d")
	etcpath := "/etc/bes-agent/conf.d"
	return getPath(path, testpath, etcpath)
}

func getPath(paths ...string) (string, error) {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// if we got here, we didn't find a file in a default location
	return "", fmt.Errorf("Could not find path in %s", paths)
}

// LoadConfig XXX
func (c *Config) LoadConfig(confPath string) error {
	var err error
	if confPath == "" {
		if confPath, err = getDefaultConfigPath(); err != nil {
			return err
		}
		log.Infof("Using config file: %s", confPath)
	}

	if _, err = toml.DecodeFile(confPath, c); err != nil {
		return err
	}

	pluginsPath, err := getPluginsPath(confPath)
	if err != nil {
		return err
	}
	patterns := [2]string{"*.yaml", "*.yaml.default"}

	// 获取config files
	var files []string
	for _, pattern := range patterns {
		m, _ := filepath.Glob(filepath.Join(pluginsPath, pattern))
		files = append(files, m...)
	}
	// config yaml files 循环加载每种插件config
	for _, file := range files {
		log.Infof("files %s", file)

		pluginConfig, err := plugin.LoadConfig(file)
		if err != nil {
			log.Errorf("Failed to parse Plugin Config %s: %s", file, err)
			continue
		}

		// windows 改斜杠
		file = filepath.ToSlash(file)
		filename := path.Base(file)
		pluginName := strings.Split(filename, ".")[0]

		// 通过插件名和配置文件 新增插件
		//py.LoadPy()
		err = c.addPlugin(pluginName, pluginConfig)
		fmt.Println(err)

		if err != nil {
			fmt.Printf("Failed to load Plugin %s: %s \n", pluginName, err)
			log.Errorf("Failed to load Plugin %s: %s", pluginName, err)
			continue
		}
	}

	return nil
}

// 在循环内， 插件级别
func (c *Config) addPlugin(name string, pluginConfig *plugin.Config) error {
	if len(c.pluginFilters) > 0 && !util.StringInSlice(name, c.pluginFilters) {
		return nil
	}

	// golang插件 (从plugins.registry注册
	checker, ok := collector.Plugins[name]
	// python插件 (从plugins.pythonRegistry注册
	pythonModule, ok2 := collector.PythonPlugins[name]
	// 是否存在该插件
	if !(ok || ok2) {
		return fmt.Errorf("Undefined plugin: %s \n", name)
	}

	// 新增是否是python插件
	//或者 if ok2 {}
	// load python 插件
	if ok2 {
		rpp := plugin.RunningPythonPlugin{
			Name:   name,
			Plugin: pythonModule,
		}
		// 先跳过多个instances循环
		c.PythonPlugins = append(c.PythonPlugins, &rpp)
		return nil
	}

	//if ok2 {
	//	for i, instance := range pluginConfig.Instances {
	//		err := py.LoadPlugin(name, instance)
	//		if err != nil {
	//			log.Errorf("ERROR to parse plugin instance [%s#%d]: %s", name, i, err)
	//			continue
	//		}
	//	}
	//	fmt.Println(pythonModule)
	//	return nil
	//}

	// golang 新增running plugin
	plug := checker(pluginConfig.InitConfig)

	// 一个插件的多个instance
	plugs := make([]plugin.Plugin, len(pluginConfig.Instances))
	for i, instance := range pluginConfig.Instances {
		err := util.FillStruct(instance, plug)
		if err != nil {
			log.Errorf("ERROR to parse plugin instance [%s#%d]: %s", name, i, err)
			continue
		}
		plugs[i] = plug
	}
	rp := &plugin.RunningPlugin{
		Name:    name,
		Plugins: plugs,
	}
	c.Plugins = append(c.Plugins, rp)
	return nil
}

// PluginNames returns a list of strings of the configured Plugins.
func (c *Config) PluginNames() []string {
	var name []string
	for _, plugin := range c.Plugins {
		name = append(name, plugin.Name)
	}
	return name
}

func (c *Config) getBindHost() string {
	host := c.GlobalConfig.BindHost
	if c.GlobalConfig.NonLocalTraffic {
		host = ""
	}
	return host
}

// GetForwarderAddr gets the address that Forwarder listening to.
func (c *Config) GetForwarderAddr() string {
	fmt.Println("ForwarderAddr -> ", c.getBindHost(), " | ", c.GlobalConfig.ListenPort)
	return fmt.Sprintf("%s:%d", c.getBindHost(), c.GlobalConfig.ListenPort)
}

// GetForwarderAddrWithScheme gets the address of Forwarder with scheme prefix.
func (c *Config) GetForwarderAddrWithScheme() string {
	return fmt.Sprintf("http://%s:%d", c.GlobalConfig.BindHost, c.GlobalConfig.ListenPort)
}

// GetStatsdAddr gets the address that Statsd listening to.
func (c *Config) GetStatsdAddr() string {
	fmt.Println("GetStatsdAddr -> ", c.getBindHost(), " | ", c.GlobalConfig.StatsdPort)
	return fmt.Sprintf("%s:%d", c.getBindHost(), c.GlobalConfig.StatsdPort)
}

// GetHostname gets the hostname from os itself if not set in the agent configuration.
func (c *Config) GetHostname() string {
	hostname := c.GlobalConfig.Hostname
	if hostname != "" {
		return hostname
	}

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Error(err)
	}
	return hostname
}

//InitializeLogging initializes logging level and output according to the agent configuration.
func (c *Config) InitializeLogging() error {
	log.Infoln("Initialize log...")
	err := log.SetLevel(c.LoggingConfig.LogLevel)
	if err != nil {
		return fmt.Errorf("Failed to parse log_level: %s", err)
	}

	logFile := c.LoggingConfig.LogFile

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(f)

	return nil
}

func (c *Config) setProjectPath() error {
	//if dev use pwdPath
	var err error
	pwdPath, _ := os.Getwd()
	err = os.Setenv("PROJECTPATH", pwdPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.ProjectPath = pwdPath
	// prd use other
	return err
}
