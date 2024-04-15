package etcdleader

import (
	"context"
	"log"

	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/clientv3"
)

func ElectLeader(cli *clientv3.Client, leaderKey string) {
	// 创建一个新的 session
	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个分布式锁
	m := concurrency.NewMutex(s, leaderKey)

	// 尝试获取锁
	if err := m.Lock(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Printf("Node became the leader.")

	// 在这里，你的节点已经成为了领导者，你可以执行领导者的任务

	// 释放锁并结束领导者的任务
	if err := m.Unlock(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Printf("Node resigned from the leadership.")
}
