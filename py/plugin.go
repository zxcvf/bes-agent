package py

import (
	"bes-agent/common/log"
	"bes-agent/common/metric"
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
		if intl, ok := x.(int); ok {
			// all good, convert to the right type, assuming YAML contains seconds
			c.interval = time.Duration(intl) * time.Second
		}
	}

	conf := make(ConfigRawMap)
	conf["name"] = c.ModuleName
	conf["init_config"] = initConfig
	conf["instances"] = []interface{}{data}

	kwargs, err := ToPython(&conf) // don't `DecRef` kwargs since we keep it around in c.config
	//fmt.Println("python kwargs >", kwargs, err)
	if err != nil {
		fmt.Println("Error parsing python check configuration: %v", err)
		return err
	}

	instance, err := c.getInstance(nil, kwargs) // don't `DecRef` instance since we keep it around in c.instance
	//fmt.Println("python instance >", instance, err)
	if err != nil {
		log.Warnf("could not get a check instance with the new api: %s", err)
		log.Warn("trying to instantiate the check with the old api, passing agentConfig to the constructor")
	}
	pyID := python.PyString_FromString(string(c.ID()))
	//fmt.Println(pyID)
	defer pyID.DecRef()
	instance.SetAttrString("check_id", pyID)
	c.instance = instance
	c.config = kwargs
	return nil
}

// Run 运行python插件
func (c *PythonCheck) Run(agg metric.Aggregator) error {
	PythonAggregatorPool[c.ID()] = &agg
	//fmt.Println("RUNheckID", c.ID())

	gstate := NewStickyLock()
	defer gstate.unlock()

	//fmt.Printf(" Running python check %s %s \n", c.ModuleName, c.id)
	emptyTuple := python.PyTuple_New(0)
	defer emptyTuple.DecRef()

	result := c.instance.CallMethod("run", emptyTuple)
	if result == nil {
		pyErr, err := gstate.getPythonError()
		if err != nil {
			return fmt.Errorf("An error occurred while running python check and couldn't be formatted: %v \n", err)
		}
		return errors.New(pyErr)
	}
	defer result.DecRef()

	// 在这里commit!!! 和RunSimple的区别是会使用聚合器上报
	//s, err := aggregator.GetSender(c.ID())
	//s.Commit()

	c.lastWarnings = c.getPythonWarnings(gstate)
	var resultStr = python.PyString_AsString(result)
	if resultStr == "" {
		return nil
	}
	return errors.New(resultStr)
}

// RunSimple 运行python插件 不使用聚合器上报数据
func (c *PythonCheck) RunSimple() error {
	gstate := NewStickyLock()
	defer gstate.unlock()

	//fmt.Printf("Simple Running python check %s %s \n", c.ModuleName, c.id)
	emptyTuple := python.PyTuple_New(0)
	defer emptyTuple.DecRef()

	result := c.instance.CallMethod("run", emptyTuple)
	//fmt.Printf("Simple Run returned for %s %s \n", c.ModuleName, c.id)
	if result == nil {
		pyErr, err := gstate.getPythonError()
		if err != nil {
			return fmt.Errorf("An error occurred while running python check and couldn't be formatted: %v \n", err)
		}
		return errors.New(pyErr)
	}
	defer result.DecRef()

	c.lastWarnings = c.getPythonWarnings(gstate)
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

// ID returns the ID of the check
func (c *PythonCheck) ID() ID {
	return c.id
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

// getInstance invokes the constructor on the Python class stored in
// `c.class` passing a tuple for args and a dictionary for keyword args.
//
// This function contains deferred calls to go-python: when you change
// this code, please ensure the Python thread unlock is always at the bottom
// of  the defer calls stack.
func (c *PythonCheck) getInstance(args, kwargs *python.PyObject) (*python.PyObject, error) {
	// Lock the GIL and release it at the end
	gstate := NewStickyLock()
	defer gstate.unlock()

	if args == nil {
		args = python.PyTuple_New(0)
		defer args.DecRef()
	}

	// invoke class constructor
	instance := c.class.Call(args, kwargs)
	if instance != nil {
		return instance, nil
	}

	// there was an error, retrieve it
	pyErr, err := gstate.getPythonError()
	if err != nil {
		return nil, fmt.Errorf("An error occurred while invoking the python check constructor, and couldn't be formatted: %v", err)
	}
	return nil, errors.New(pyErr)
}

// getPythonWarnings grabs the last warnings from the python check
func (c *PythonCheck) getPythonWarnings(gstate *stickyLock) []error {
	/**
	This function must be run before the GIL is unlocked, otherwise it will return nothing.
	**/
	warnings := []error{}
	emptyTuple := python.PyTuple_New(0)
	defer emptyTuple.DecRef()
	ws := c.instance.CallMethod("get_warnings", emptyTuple)
	if ws == nil {
		pyErr, err := gstate.getPythonError()
		if err != nil {
			log.Errorf("An error occurred while grabbing python check and couldn't be formatted: %v", err)
		}
		log.Infof("Python error: %v", pyErr)
		return warnings
	}
	defer ws.DecRef()
	numWarnings := python.PyList_Size(ws)
	idx := 0
	for idx < numWarnings {
		w := python.PyList_GetItem(ws, idx) // borrowed ref
		warnings = append(warnings, fmt.Errorf("%v", python.PyString_AsString(w)))
		idx++
	}
	return warnings
}
