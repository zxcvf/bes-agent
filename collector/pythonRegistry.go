package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var BasePythonPath string

// GetAllFiles 获取指定目录下的所有文件,包含子目录下的文件.  suffix识别文件后缀
func GetAllFiles(dirPth string, suffix string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			GetAllFiles(dirPth+PthSep+fi.Name(), suffix)
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), suffix)
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table, suffix)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

func GetPythonPaths() (files []string, err error) {
	x := os.Getenv("PROJECTPATH") + "/collector/plugins"
	files, err = GetAllFiles(x, ".py")
	return files, err
}

func setPythonEnvPath() {
	//if dev use pwdPath
	pwdPath, _ := os.Getwd()
	fmt.Println("setPythonEnvPath ", pwdPath)
	err := os.Setenv("PYTHONPATH", pwdPath)
	err = os.Setenv("PROJECTPATH", pwdPath)
	BasePythonPath = pwdPath
	if err != nil {
		fmt.Println(err)
	}

}

//var PythonPlugins = map[string]Checker{}
//
//// Hello_plugin XXX
//type pythonPlugin struct {
//	plugin.Plugin
//
//	fileName string
//	file string
//}
//
//
//// PythonChecker init 空的python插件
//func PythonChecker(conf plugin.InitConfig) plugin.Plugin {
//	return &pythonPlugin{}
//}
//
//// PythonAdd XXX
//func PythonAdd(name string, checker Checker, file string) {
//	Plugins[name] = checker
//}

var PythonPlugins = map[string]string{} // k: pluginName v: modulePath

func PythonAdd(name string, module string) {
	PythonPlugins[name] = module
}

// 注册所有python插件
func init() {
	setPythonEnvPath()
	files, _ := GetPythonPaths()

	// Add python plugin  应该去掉非插件的python文件
	for _, file := range files {
		file = filepath.ToSlash(file)

		filename := path.Base(file)
		pluginName := strings.Split(filename, ".")[0]
		fmt.Printf("register python plugin %v in %v \n", pluginName, file)

		relativePath, _ := filepath.Rel(BasePythonPath, file) // collector\plugins\hello_python2.py

		moduleName := strings.Split(relativePath, ".")[0]
		moduleName = strings.Replace(moduleName, "/", ".", -1)  // linux
		moduleName = strings.Replace(moduleName, "\\", ".", -1) // windows

		PythonAdd(pluginName, moduleName)
	}

}
