package server

import (
	"context"
	v1 "distributed-learning-lab/harmoniakv/api/v1"
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/coordinator"
	"distributed-learning-lab/harmoniakv/storage"
	"distributed-learning-lab/harmoniakv/transport"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/metadata"
)

type KvServer interface {
	v1.HarmoniakvServer
}

type server struct {
	v1.UnimplementedHarmoniakvServer
	config      config.Config
	coordinator coordinator.Coordinator
	storage     storage.Storage
	transer     transport.Transporter
}

func (s *server) Get(ctx context.Context, req *v1.GetRequest) (*v1.GetResponse, error) {
	//TODO: 如何处理ctx 之后再议
	objects, err := s.coordinator.HandleGet(req.Key)
	return &v1.GetResponse{
		Objects: objects,
	}, err
}

func (s *server) Put(ctx context.Context, req *v1.PutRequest) (*v1.PutResponse, error) {
	// s.coordinator.HandlePut(req.)
	var meta *coordinator.Metadata
	pairs, ok := metadata.FromIncomingContext(ctx)
	if ok {
		mapstructure.Decode(pairs, meta)
	}
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
