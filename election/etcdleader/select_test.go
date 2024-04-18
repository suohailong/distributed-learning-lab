package etcdleader

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/server/v3/embed"

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

func TestDistributeLock(t *testing.T) {
	// 启动一个嵌入式的etcd服务
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	defer func() {
		if err := os.RemoveAll(cfg.Dir); err != nil {
			log.Fatalf("clear path: %s err: %v", cfg.Dir, err)
		}
	}()
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("Failed to start etcd server: %v", err)
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2380"}})
		if err != nil {
			log.Fatalf(err.Error())
		}
		session, err := concurrency.NewSession(cli)
		if err != nil {
			log.Fatalf(err.Error())
		}
		lock := concurrency.NewMutex(session, "/my-lock")
		lock.Lock(context.Background())

	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Debug("Server took too long to start!")
	}
	log.Debug(<-e.Err())
}
