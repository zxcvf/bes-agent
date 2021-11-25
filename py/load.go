package py

import (
	"bes-agent/common/plugin"
	"errors"
	"fmt"
	"github.com/sbinet/go-python"
)

func LoadPy() error {
	// 起到DataDog中python check loader作用
	python.Initialize()
	//defer python.Finalize()

	InitCAggregator()

	// 加载collector.plugins模块 再加载collector.plugins.test()等函数、类
	const agentCheckModuleName = "collector.plugins"
	agentCheckModule := python.PyImport_ImportModule(agentCheckModuleName)
	fmt.Println(agentCheckModule)
	if agentCheckModule == nil {
		return errors.New("unable to initialize AgentCheck module")
	}
	defer agentCheckModule.DecRef()

	// 加载 python agent class
	const agentCheckClassName = "AgentCheck"
	agentCheckClass := agentCheckModule.GetAttrString(agentCheckClassName)
	if agentCheckClass == nil {
		return errors.New("unable to initialize AgentCheck class")
	}

	return nil
}

func LoadPlugin(name string, instance plugin.Instance) error {
	return nil
}
