package zookeeper

import (
	"bes-agent/common/config"
	"bes-agent/common/log"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

func GetKafkaServerInfo(conf *config.Config) (kafkaServers, filebeatTopic string) {
	myConfig := config.NewNodeConfig(conf.ProjectPath + "/conf/node.conf")
	zookeeperInfo := myConfig.Read("Main", "zookeeper_info", "")
	zookeeperInfo = strings.Replace(zookeeperInfo, " ", "", -1)

	zkAddrs := strings.Split(zookeeperInfo, ",")

	var hosts []string

	//var result string
	if !strings.Contains(zkAddrs[0], ":") {
		// ip,ip:port
		IpAndPort := strings.Split(zookeeperInfo, ":") // ip,ip,ip port
		for _, ip := range strings.Split(IpAndPort[0], ",") {
			port := IpAndPort[1]
			if port == "" {
				port = "2181"
			}
			addr := ip + ":" + port
			hosts = append(hosts, addr)
		}
	} else {
		// ip:port, ip:port  和 ip:port
		hosts = zkAddrs
	}

	// 连接zk
	zooKeeper := NewZooKeeper(hosts)
	//defer zooKeeper.close()

	if zooKeeper.state == zk.StateHasSession {
		result, _ := zooKeeper.get(ZkConfigDbPath)
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			bootstrapServersEquals := BootstrapServers + "="
			if strings.Contains(line, bootstrapServersEquals) {
				kafkaServers = strings.Replace(line, bootstrapServersEquals, "", -1)
				filebeatTopic = myConfig.Read("Main", "filebeat_topic", "webgate-infrastructure-metric-log")
				return kafkaServers, filebeatTopic
			}
		}
	}

	log.Warnln("Connection zk failed!")
	bootstrapServer := "localhost:9092"
	return bootstrapServer, "webgate-infrastructure-metric-log"
}
