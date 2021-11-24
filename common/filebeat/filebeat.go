package filebeat

import (
	"bes-agent/common/config"
	"bes-agent/common/log"
	"bes-agent/common/plugin"
	"bes-agent/common/zookeeper"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type Filebeat struct {
	conf           *config.Config
	filebeatPath   string //common.filebeat文件夹路径 /agent/common/filebeat
	filebeatSource string //filebeat文件夹路径        /agent/filebeat
	state          *os.ProcessState
}

func NewFilebeat(conf *config.Config) *Filebeat {
	// getConfigDPath()  get
	// auto_conf 支持

	// get filebeat path
	// linux 改为ENV WEBGATE_NODE_HOME
	// todo node_path + /filebeat E:\\webgate\\webagte-infrastructue\\webgate-inrastructure-probe\\src/../filebeat

	return &Filebeat{
		conf:           conf,
		filebeatPath:   path.Join(conf.ProjectPath, "/common/filebeat"),
		filebeatSource: path.Join(conf.ProjectPath, "/filebeat"),
	}
}

func (f Filebeat) getInstances() ([]plugin.Instance, error) {
	// 获取filebeat.yml 中的instance
	filebeatConf, err := plugin.LoadConfig(f.conf.ProjectPath + "/conf/filebeat.yaml")
	if err != nil {
		return []plugin.Instance{}, err
	}

	return filebeatConf.Instances, err
}

func (f Filebeat) waitInstances(instances []plugin.Instance) ([]plugin.Instance, error) {
	// 等连接
	if len(instances) != 0 {
		time.Sleep(30 * time.Second)
	}

	return f.getInstances()
}

func (f Filebeat) writeFilebeatYaml() error {

	kafkaServers, filebeatTopic := zookeeper.GetKafkaServerInfo(f.conf)

	var kafkaServersSliceString string
	kafkaServersSlice := strings.Split(kafkaServers, ",")
	for _, v := range kafkaServersSlice {
		kafkaServersSliceString += fmt.Sprintf("  - %s\n", v)
	}

	temp, err := ioutil.ReadFile(path.Join(f.filebeatPath, "/filebeat.format.yaml"))
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
		return err
	}
	tempString := string(temp)

	// 写入./filebeat/filebeat.yml
	filebeatYamlString := fmt.Sprintf(
		tempString,
		f.filebeatPath,
		f.filebeatSource,
		kafkaServersSliceString,
		filebeatTopic,
	)
	filebeatYamlPath := path.Join(f.filebeatPath, "filebeat.yml")
	err = writeFile(filebeatYamlPath, filebeatYamlString)
	return err
}

func (f Filebeat) writeFilebeatInput(instance plugin.Instance) error {
	// valid yml required fileds
	required := []string{"log_type", "object_type", "paths", "multiline.negate", "multiline.match", "multiline.pattern"}
	var leftArgs []string
	for _, v := range required {
		value, has := instance[v]
		if has != true || value == nil {
			log.Errorf("plugins filebeat.yml: %s is null!", v)
			leftArgs = append(leftArgs, v)
		}
	}
	if len(leftArgs) > 0 {
		return errors.New(fmt.Sprintf("plugins filebeat.yml need %+q params", leftArgs))
	}

	var tagString string
	tags := instance["tags"].([]interface{})
	for _, v := range tags {
		tag := v.(string)
		tagSlice := strings.Split(tag, ":")
		if len(tagSlice) == 2 && tagSlice[0] != "" {
			tagString += fmt.Sprintf("    %s: %s\n", tagSlice[0], tagSlice[1])
			//tagMap[tagSlice[0]] = tagSlice[1]
		}
	}
	tagString += fmt.Sprintf("    %s: %s\n", "log_type", instance["log_type"].(string))
	tagString += fmt.Sprintf("    %s: %s\n", "object_type", instance["object_type"].(string))

	pattern := instance["object_type"]
	negate := "false"
	if strings.EqualFold(instance["multiline.negate"].(string), "true") {
		negate = "true"
	}
	match := instance["multiline.match"]

	var paths string
	for _, v := range strings.Split(instance["paths"].(string), ",") {
		paths += fmt.Sprintf("  - %s\n", v)
	}

	fmt.Println(pattern, negate, match, paths)

	temp, err := ioutil.ReadFile(path.Join(f.filebeatPath, "/input.format.yaml"))
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
		return err
	}
	tempString := string(temp)

	filebeatInputPath := path.Join(f.filebeatPath, "input.yml")
	filebeatInputString := fmt.Sprintf(
		tempString,
		pattern,
		negate,
		match,
		paths,
		tagString,
	)
	err = writeFile(filebeatInputPath, filebeatInputString)

	return err
}

func (f Filebeat) execute(err error) error {
	// todo reload

	// getSubprocessArgs 获取 执行文件路径 -e -c filebeat.yml路径
	execPath := path.Join(f.filebeatSource, "/filebeat") // linux execPath
	if runtime.GOOS == "windows" {
		execPath = path.Join(f.filebeatSource, "/filebeat.exe") // windows execPath
	}

	procAttr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	var process *os.Process
	if err == nil {
		process, err = os.StartProcess(execPath, []string{"-e", "-c", f.filebeatPath + "/filebeat.yml"}, procAttr)
	} else {
		process, err = os.StartProcess(execPath, []string{"-e"}, procAttr)
	}

	if err != nil {
		fmt.Println("filebeat start process error:", err)
		log.Errorf("filebeat start process error: %s\n", err)
		return err
	}
	processState, err := process.Wait()
	if err != nil {
		fmt.Println("filebeat wait error:", err)
		log.Errorf("filebeat wait error: %s\n", err)
		return err
	}
	f.state = processState
	fmt.Println("start filebeat")
	return nil
}

func (f Filebeat) Run(shutdown chan struct{}) error {

	instances, err := f.getInstances() // []
	if err != nil {
		log.Fatalf("failed to getInstances %s", err)
	}

	//instances, err = f.waitInstances(instances)
	fmt.Println(instances, "instances")
	if err != nil {
		log.Fatalf("failed to WaitInstances: %s", err)
	}

	// 没有instance 直接跳过以下
	if len(instances) == 0 {
		goto executeSubprocess
	}

	err = f.writeFilebeatYaml()
	if err != nil {
		log.Errorf("run writeFilebeatYaml in err %s", err)
		goto executeSubprocess
	}
	err = f.writeFilebeatInput(instances[0])
	if err != nil {
		log.Errorf("run writeFilebeatYaml in err %s", err)
		goto executeSubprocess
	}

executeSubprocess:
	err = f.execute(err)
	return err
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func writeFile(path string, content string) error {
	var err error
	var file *os.File
	if checkFileIsExist(path) {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	} else {
		file, _ = os.Create(path)
	}
	_, err = io.WriteString(file, content)
	return err
}
