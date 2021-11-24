package py

import (
	"errors"
	"fmt"
	"github.com/sbinet/go-python"
	"runtime"
	"time"
)

// PythonPlugin represents a Python check, implements `Check` interface
type PythonPlugin struct {
	id           ID
	instance     *python.PyObject
	class        *python.PyObject
	ModuleName   string
	config       *python.PyObject
	interval     time.Duration
	lastWarnings []error
}

// NewPythonPlugin conveniently creates a PythonPlugin instance
func NewPythonPlugin(name string, class *python.PyObject) *PythonPlugin {
	glock := newStickyLock()
	class.IncRef() // own the ref
	glock.unlock()
	pyCheck := &PythonPlugin{
		ModuleName:   name,
		class:        class,
		interval:     5 * time.Second,
		lastWarnings: []error{},
	}
	runtime.SetFinalizer(pyCheck, PythonPluginFinalizer)
	return pyCheck
}

// RunSimple runs a Python check without sending data to the aggregator
func (c *PythonPlugin) RunSimple() error {
	gstate := newStickyLock()
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

// PythonPluginFinalizer is a finalizer that decreases the reference count on the PyObject refs owned
// by the PythonPlugin.
// 减少PyObject的引用计数
func PythonPluginFinalizer(c *PythonPlugin) {
	// Run in a separate goroutine because acquiring the python lock might take some time,
	// and we're in a finalizer
	go func(c *PythonPlugin) {
		glock := newStickyLock() // acquire lock to call DecRef
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
