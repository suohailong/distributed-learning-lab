package server

import (
	"distributed-learning-lab/HarmoniaKV/server/coordinator"

	"honnef.co/go/tools/config"
)

type AdminApi interface {
	AddNode()
	RemoveNode()
}

type UserApi interface {
	Put(key string, value string) error
	Get(key string) (value string, err error)
}

type KvServer interface {
	UserApi
	AdminApi
}

type server struct {
	config      config.Config
	coordinator coordinator.Coordinator
}

func New() KvServer {
	return &server{}
}

func (s *server) AddNode() {
	panic("not implemented") // TODO: Implement
	// 向coordinator添加一个Node

}

func (s *server) RemoveNode() {
	panic("not implemented") // TODO: Implement
}

func (h *server) Put(key string, value string) error {
	panic("not implemented") // TODO: Implement
	// 从coordinator中获取key存放在那些node上
	nodes := h.coordinator.GetNodes(key, h.config.Replicas)
	// 建立和其他node的连接, 存取key
	transport.SendRequest(nodes, req)

}

func (h *server) Get(key string) (value string, err error) {
	panic("not implemented") // TODO: Implement
}
