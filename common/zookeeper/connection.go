package zookeeper

import (
	"bes-agent/common/log"
	"fmt"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

const (
	ZkCollectorInfoPath = "/webgate/cluster/collectors"
	ZkIpNodeConf        = "zookeeper_info"
	ZkConfigDbPath      = "/webgate/cluster/db-config"
	BootstrapServers    = "bootstrap.servers"
	TestPath            = "/testplatform/test"
)

type ZooKeeper struct {
	conn  *zk.Conn
	state zk.State
}

func NewZooKeeper(hosts []string) *ZooKeeper {
	conn, session, err := zk.Connect(hosts, time.Second*10)
	if err != nil {
		fmt.Println(err)
		log.Errorf("connect zk failed %s \n", err)
		return nil
	}
	var state zk.State

	for event := range session {
		if event.State == zk.StateHasSession {
			fmt.Printf("zookeeper State: %s \n", event.State)
			state = zk.StateHasSession
			break
		}
	}
	return &ZooKeeper{
		conn,
		state,
	}
}

func (z *ZooKeeper) get(path string) (string, error) {
	data, _, err := z.conn.Get(path)
	if err != nil {
		log.Errorf("查询%s失败, err: %v\n", "", err)
		fmt.Printf("查询%s失败, err: %v\n", "", err)
		return "", err
	}
	fmt.Printf("%s 的值为 %s\n", path, string(data))
	return string(data), nil
}

func (z ZooKeeper) close() {
	z.conn.Close()
}

func (z *ZooKeeper) add(path string) error {
	var data = []byte("test value")
	// flags有4种取值：
	// 0:永久，除非手动删除
	// zk.FlagEphemeral = 1:短暂，session断开则该节点也被删除
	// zk.FlagSequence  = 2:会自动在节点后面添加序号
	// 3:Ephemeral和Sequence，即，短暂且自动添加序号
	var flags int32 = zk.FlagEphemeral
	// 获取访问控制权限
	acls := zk.WorldACL(zk.PermAll)
	s, err := z.conn.Create(TestPath, data, flags, acls)

	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		return err
	}
	fmt.Printf("创建: %s 成功", s)
	return nil
}
