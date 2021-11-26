plugin 插件

check 插件根据intance实例化后的实体


```golang
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
func main()  {
	setPythonEnvPath()
	python.Initialize()
	pluginName := "test2"

	var wg sync.WaitGroup
    // Initialize() 已经锁定 GIL ，但这时我们并不需要它。
    // 我们保存当前状态和释放锁，从而让 goroutine 能获取它
	state := python.PyEval_SaveThread()

	for i:=0 ; i<10; i++{
		wg.Add(1)

		go func(pluginModule string, x int) {
			defer wg.Done()
			fmt.Println(x, "GIL Ensure")
			//runtime.LockOSThread()  // https://www.datadoghq.com/blog/engineering/cgo-and-python/

            // 加全局解释锁
			_gstate := python.PyGILState_Ensure()
			pythonRunningModule := python.PyImport_ImportModule(pluginName)
			pythonRunningModule.CallMethod("test", python.PyTuple_New(0))
			python.PyGILState_Release(_gstate)
			fmt.Println(x, "GIL release")
		}(pluginName, i)
	}

	wg.Wait()
	// 一定要在wg.Wait()后
    // 在这里我们知道程序不会再需要运行 Python 代码了，
    // 我们可以恢复状态和 GIL 锁，执行退出前的最后操作。
	python.PyEval_RestoreThread(state)
    python.Finalize()
}

```