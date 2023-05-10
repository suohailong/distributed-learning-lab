package server

import (
	"distributed-learning-lab/HarmoniaKV/server/coordinator"
	"distributed-learning-lab/harmoniakv/server/config"
	"distributed-learning-lab/harmoniakv/server/storage"

	"github.com/coreos/etcd/pkg/transport"
	"github.com/docker/docker/pkg/plugins/transport"
)

type AdminApi interface {
	AddNode()
	RemoveNode()
}

type UserApi interface {
	Delete(key string) error
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
	storage     storage.Storage
	transer     transport.Transporter
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
	// 从coordinator中获取key 所在的node
	// 建立和其他node的连接, 存取key
	// 根据一致性策略判断返回是否符合要求

	return nil
}

func (h *server) Get(key string) (value string, err error) {
	// 从coordinator中获取node
	// 建立和其他node的连接, 存取key
	// 根据一致性策略判断返回是否符合要求

	// 返回值

	return value, nil
}

func (h *server) Delete(key string) error {

	return nil
}
