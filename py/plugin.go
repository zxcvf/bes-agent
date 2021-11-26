package py

import (
	"bes-agent/common/plugin"
	"errors"
	"fmt"
	"github.com/sbinet/go-python"
	"runtime"
	"sync"
	"time"
)

type ConfigRawMap map[interface{}]interface{}
type ConfigData []byte

// PythonCheck represents a Python check, implements `Check` interface
type PythonCheck struct {
	id           ID
	instance     *python.PyObject
	class        *python.PyObject
	ModuleName   string
	config       *python.PyObject
	interval     time.Duration
	lastWarnings []error
}

// NewPythonCheck conveniently creates a PythonPlugin instance
func NewPythonCheck(name string, class *python.PyObject) *PythonCheck {
	glock := NewStickyLock()
	class.IncRef() // own the ref
	glock.unlock()
	pyCheck := &PythonCheck{
		ModuleName:   name,
		class:        class,
		interval:     5 * time.Second,
		lastWarnings: []error{},
	}
	runtime.SetFinalizer(pyCheck, PythonPluginFinalizer)
	return pyCheck
}

// 配置check
func (c *PythonCheck) Configure(data plugin.Instance, initConfig plugin.InitConfig) error {
	c.id = Identify(c, data, initConfig)

	// See if a collection interval was specified
	x, ok := data["min_collection_interval"]
	if ok {
		// we should receive an int from the unmarshaller
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!", ok)
		if intl, ok := x.(int); ok {
			// all good, convert to the right type, assuming YAML contains seconds
			c.interval = time.Duration(intl) * time.Second
		}
	}

	conf := make(ConfigRawMap)
	conf["name"] = c.ModuleName
	conf["init_config"] = initConfig
	conf["instances"] = []interface{}{data}

	//kwargs, err := ToPython(&conf) // don't `DecRef` kwargs since we keep it around in c.config

	return nil
}

// Run 运行python插件
func (c *PythonCheck) Run() error {
	return nil
}

// RunSimple 运行python插件 不适用聚合器上报数据
func (c *PythonCheck) RunSimple() error {
	gstate := NewStickyLock()
	defer gstate.unlock()

	fmt.Printf("Running python check %s %s \n", c.ModuleName, c.id)
	emptyTuple := python.PyTuple_New(0)
	defer emptyTuple.DecRef()

	result := c.instance.CallMethod("run", emptyTuple)
	fmt.Printf("Run returned for %s %s \n", c.ModuleName, c.id)
	if result == nil {
		pyErr, err := gstate.getPythonError()
		if err != nil {
			return fmt.Errorf("An error occurred while running python check and couldn't be formatted: %v \n", err)
		}
		return errors.New(pyErr)
	}
	defer result.DecRef()

	//c.lastWarnings = c.getPythonWarnings(gstate)
	var resultStr = python.PyString_AsString(result)
	if resultStr == "" {
		return nil
	}
	return errors.New(resultStr)
}

// RUN PYTHON FUNC 运行简单的python函数
func RunFunc(pluginModule string, wg *sync.WaitGroup) error {
	defer wg.Done()
	// 在goroutine 需要加解释锁
	_gstate := python.PyGILState_Ensure()
	pythonRunningModule := python.PyImport_ImportModule(pluginModule)
	pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
	python.PyGILState_Release(_gstate)
	return nil
}

// PythonPluginFinalizer is a finalizer that decreases the reference count on the PyObject refs owned
// by the PythonPlugin.
// 减少PyObject的引用计数
func PythonPluginFinalizer(c *PythonCheck) {
	// Run in a separate goroutine because acquiring the python lock might take some time,
	// and we're in a finalizer
	go func(c *PythonCheck) {
		glock := NewStickyLock() // acquire lock to call DecRef
		defer glock.unlock()
		c.class.DecRef()
		if c.instance != nil {
			c.instance.DecRef()
		}
		if c.config != nil {
			c.config.DecRef()
		}
	}(c)
}
