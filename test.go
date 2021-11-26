package main

import (
	"fmt"
	"github.com/sbinet/go-python"
	"os"
	"sync"
	//"time"
)

func setPythonEnvPath() {
	pwdPath, _ := os.Getwd()
	err := os.Setenv("PYTHONPATH", pwdPath)
	if err != nil {
		fmt.Println(err)
	}

}
func main() {
	setPythonEnvPath()
	python.Initialize()
	pluginName := "test2"

	var wg sync.WaitGroup
	state := python.PyEval_SaveThread()

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(pluginModule string, x int) {
			defer wg.Done()
			fmt.Println(x, "GIL Ensure")
			//runtime.LockOSThread()  // https://www.datadoghq.com/blog/engineering/cgo-and-python/
			_gstate := python.PyGILState_Ensure()
			pythonRunningModule := python.PyImport_ImportModule(pluginName)
			pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
			python.PyGILState_Release(_gstate)
			fmt.Println(x, "GIL release")
		}(pluginName, i)
	}
	wg.Wait()
	// 一定要在wg.Wait()后
	python.PyEval_RestoreThread(state)

}
