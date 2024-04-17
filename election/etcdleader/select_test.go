package etcdleader

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/server/v3/embed"
)

func TestCampaign(t *testing.T) {
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
		fmt.Println("Server is ready!")
		Campaign()
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Printf("Server took too long to start!")
	}
	log.Fatal(<-e.Err())
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
		log.Printf("Server took too long to start!")
	}
	log.Fatal(<-e.Err())
}
