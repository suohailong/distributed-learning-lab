package node

import (
	"distributed-learning-lab/harmoniakv/node/storage"
)

const (
	ONLINE  = 1 // 0
	OFFLINE = -1
)

type Message struct {
	MessageType uint32
	From        string
	to          string
}

type Node struct {
	ID      string
	Address string
	//TODO: 这里会一直增长吗
	Counter   uint64
	HeartTime int64
	State     int8
	SendTime  int64

	store storage.Storage
}

func NewNode(id string, address string) *Node {
	return &Node{
		ID:        id,
		Address:   address,
		Counter:   0,
		State:     ONLINE,
		HeartTime: 0,
		SendTime:  0,
	}
}

func (n *Node) GetId() string {
	return n.ID
}

func (n *Node) HandleMessage() {

}
