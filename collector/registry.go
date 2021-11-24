package collector

import (
	"bes-agent/common/plugin"
)

// Checker XXX New插件
//    func NewNginx(conf plugin.InitConfig) plugin.Plugin {
//    	return &Nginx{}
//    }

type Checker func(conf plugin.InitConfig) plugin.Plugin

// Plugins XXX
var Plugins = map[string]Checker{}

// Add XXX  collector.Add
func Add(name string, checker Checker) {
	Plugins[name] = checker
}
