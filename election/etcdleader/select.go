package etcdleader

import (
	"context"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func ElectLeader() {
	cli, err := clientv3.New(clientv3.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个新的 session
	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	e := concurrency.NewElection(s, "leader")

	leaderReady := make(chan struct{})
	go func() {
		err := e.Campaign(context.Background(), "value")
		// 创建一个分布式锁
		if err != nil {
			log.Fatal(err)
		}
		leaderReady <- e
	}()
}
