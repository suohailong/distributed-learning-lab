package etcdleader

import (
	"flag"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"distributed-learning-lab/util/log"
)

var nodeId = flag.String("id", "001", "-id=001")

func TestCampaign(t *testing.T) {
	// 启动一个嵌入式的etcd服务
	flag.Parse()
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:12379", "localhost:32379", "localhost:22379"}})
	if err != nil {
		log.Errorf("create etcd client failed, err: %v", err)
		return
	}
	defer cli.Close()
	s1, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		log.Errorf("create etcd session failed, err: %v", err)
		return
	}
	defer s1.Close()
	err = Campaign(s1, *nodeId)
	if err != nil {
		log.Errorf("campaign failed: %s", *nodeId)
		return
	}
	tick := time.NewTicker(5 * time.Second)
	for range tick.C {
		log.Debugf("node: %s start worker", *nodeId)
	}
}
