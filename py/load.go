package py

import (
	"bes-agent/common/plugin"
	"errors"
	"fmt"
	"github.com/sbinet/go-python"
)

var AgentCheckClass *python.PyObject

func LoadPy() error {
	// 起到DataDog中python check loader作用
	err := python.Initialize()
	//defer python.Finalize()

	InitCAggregator()

	// 加载collector.plugins模块 再加载collector.plugins.test()等函数、类
	const agentCheckModuleName = "collector.plugins"
	agentCheckModule := python.PyImport_ImportModule(agentCheckModuleName)
	if agentCheckModule == nil {
		return errors.New("unable to initialize AgentCheck module")
	}
	defer agentCheckModule.DecRef()

	// 加载 python agent class
	const agentCheckClassName = "AgentCheck"
	AgentCheckClass = agentCheckModule.GetAttrString(agentCheckClassName)
	if AgentCheckClass == nil {
		return errors.New("unable to initialize AgentCheck class")
	}

	return err
}

// 将插件中的类 加载出instances对应obj
func LoadChecks(rpp *plugin.RunningPythonPlugin) ([]*PythonCheck, error) {
	var err error
	checks := []*PythonCheck{}
	glock := NewStickyLock() // GIL需要从从循环外部传入

	// windows 需要加  目前暂时不支持windows
	// Platform-specific preparation
	//err = platformLoaderPrep()
	//if err != nil {
	//	return nil, err
	//}
	//defer platformLoaderDone()

	var pyErr string
	var pluginModule *python.PyObject

	pluginModule = python.PyImport_ImportModule(rpp.Module)

	if pluginModule == nil {
		pyErr, err = glock.getPythonError()
		if err != nil {
			err = fmt.Errorf("An error occurred while loading the python module and couldn't be formatted: %v \n", err)
			fmt.Printf("Unable to load python module - %s: %v \n", rpp.Name, err)
		} else {
			err = errors.New(pyErr)
			fmt.Printf("Unable to load python module - %s: %v \n", rpp.Name, err)
		}
		defer glock.unlock()
		return nil, err
	}

	// 寻找AgentCheck的子类
	checkClass, err := findSubclassOf(AgentCheckClass, pluginModule, glock)

	pluginModule.DecRef()
	glock.unlock()
	// 没有子类 返回
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to find acheck class in the module: %v \n", err))
	}

	for _, instance := range rpp.Instances {
		check := NewPythonCheck(rpp.Module, checkClass)
		// The GIL should be unlocked at this point, `check.Configure` uses its own stickyLock and stickyLocks must not be nested

		if err := check.Configure(instance, rpp.InitConfig); err != nil {
			fmt.Errorf("py.loader: could not configure check '%s': %s", rpp.Name, err)
			continue
		}
		checks = append(checks, check)
	}

	glock = NewStickyLock()
	defer glock.unlock()
	checkClass.DecRef()
	//fmt.Printf("python loader: done loading check %s \n", rpp.Module)
	return checks, err
}
