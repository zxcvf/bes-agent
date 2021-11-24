package plugin

import (
	"io/ioutil"

	"bes-agent/common/metric"

	yaml "gopkg.in/yaml.v2"
)

// Plugin ..
type Plugin interface {
	// Check takes in an aggregator and adds the metrics that the Plugin
	// gathers. This is called every "interval"
	Check(agg metric.Aggregator) error
}

// RunningPlugin XXX
type RunningPlugin struct {
	Name    string
	Plugins []Plugin
}

// RunningPythonPlugin XXX
type RunningPythonPlugin struct {
	Name   string
	Plugin string // 先跳过instances (目前是python模块)
}

// InitConfig XXX
type InitConfig map[string]interface{}

// Instance XXX
type Instance map[string]interface{}

// Config XXX
type Config struct {
	InitConfig InitConfig `yaml:"init_config"`
	Instances  []Instance `yaml:"instances"`
}

// LoadConfig parses the YAML file into a Config.
func LoadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal([]byte(string(content)), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
