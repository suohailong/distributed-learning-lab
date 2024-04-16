package etcdleader

import (
	"fmt"
	"log"
	"testing"
	"time"

	"go.etcd.io/etcd/server/v3/embed"
)

func TestCampaign(t *testing.T) {
	// 启动一个嵌入式的etcd服务
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
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
