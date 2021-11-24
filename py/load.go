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
	if agentCheckModule == nil {
		fmt.Errorf("unable to initialize AgentCheck module")
		return errors.New("unable to initialize AgentCheck module")
	}
	defer agentCheckModule.DecRef()

	// 加载 python agent class
	const agentCheckClassName = "AgentCheck"
	agentCheckClass := agentCheckModule.GetAttrString(agentCheckClassName)
	if agentCheckClass == nil {
		fmt.Errorf("Unable to import %s class from Python module: %s \n", agentCheckClassName, agentCheckModuleName)
		return errors.New("unable to initialize AgentCheck class")
	}

	return nil
}

func LoadPlugin(name string, instance plugin.Instance) error {
	//plugins := []plugin.RunningPythonPlugin{}
	fmt.Println(instance, "!!!!!!!!!!!!")
	moduleName := name
	whlModuleName := fmt.Sprintf("bes_checks.%s", name)
	modules := []string{moduleName, whlModuleName}

	// Lock the GIL while working with go-python directly
	glock := newStickyLock()
	var err error
	var pyErr string
	var checkModule *python.PyObject
	for _, name := range modules {
		// import python module containing the check
		fmt.Println("!!!!!!!!!!!!!!!!!!!!", err, pyErr, checkModule, name)
		//checkModule = python.PyImport_ImportModule(name)
		//if checkModule != nil {
		//	break
		//}
		//
		//pyErr, err = glock.getPythonError()
		//if err != nil {
		//	err = fmt.Errorf("An error occurred while loading the python module and couldn't be formatted: %v \n", err)
		//} else {
		//	err = errors.New(pyErr)
		//}
		//fmt.Printf("Unable to load python module - %s: %v \n", name, err)
	}
	glock.unlock()

	return nil
}

// 没用
func PyRun() {
	python.Initialize()
	defer python.Finalize()
	//NewPythonCheckLoader
	//type PythonCheckLoader struct {
	//	agentCheckClass *python.PyObject
	//}
	const agentCheckClassName = "AgentCheck"
	const agentCheckModuleName = "collector.plugins"
	agentCheckModule := python.PyImport_ImportModule(agentCheckModuleName)
	if agentCheckModule == nil {
		fmt.Errorf("unable to initialize AgentCheck module")
		//return nil
	}
	defer agentCheckModule.DecRef()

	fmt.Println(agentCheckModule)

	//initAPI()          // `aggregator` module

	//agentCheckClass := agentCheckModule.GetAttrString(agentCheckClassName) // don't `DecRef` for now since we keep the ref around in the returned PythonCheckLoader
	//if agentCheckClass == nil {
	//	fmt.Printf("Unable to import %s class from Python module: %s\n", agentCheckClassName, agentCheckModuleName)
	//	//return nil
	//}
	//// return &PythonCheckLoader{agentCheckClass}, nil
	//
	//glock := newStickyLock()  // Lock the GIL while working with go-python directly
	//
	//var err error
	//var pyErr string
	//
	//// 导入包含插件的python模块 collector.plugins
	//var checkModule *python.PyObject
	//checkModule = python.PyImport_ImportModule("collector.plugins")
	//if checkModule == nil{
	//	pyErr, err = glock.getPythonError()
	//	if err != nil {
	//		err = fmt.Errorf("An error occurred while loading the python module and couldn't be formatted: %v", err)
	//	} else {
	//		err = errors.New(pyErr)
	//	}
	//}
	//
	//if checkModule == nil {
	//	defer glock.unlock()
	//	// GG
	//	return
	//}
	//
	//checkClass, err := findSubclassOf(agentCheckClass, checkModule, glock)
	//
	//fmt.Println(checkClass, "py demo")
	//fmt.Println(checkModule, "py demo")
	//fmt.Println(pyErr, "py demo")
	//fmt.Println(err, "py demo")

}
