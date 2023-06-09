package node

import (
	"distributed-learning-lab/harmoniakv/node/storage"
	"distributed-learning-lab/harmoniakv/node/version"

	"github.com/sirupsen/logrus"
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

const (
	GET = 1
	PUT = 2
)

type KvCommand struct {
	Command uint32
	Key     []byte
	Value   version.Value
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

//go:noinline
func (n *Node) HandleMsg(cmd *KvCommand, cb func(interface{})) {
	logrus.Debugf("handle local msg: %v", cmd)
	if cmd.Command == GET {
		// 3. 从本地数据库读取该key的所有版本

		// 4. 将所有版本返回给协调节点
		cb(cmd)
		return
	} else if cmd.Command == PUT {
		// 2. 生成该key对应的版本向量, 写入本地数据库
		cmd.Value.VersionVector.Increment(n.ID)
		// 3. 从本地数据库读取该key的所有版本
		n.store.Put(cmd.Key, cmd.Value)
		// 4. 将所有版本返回给协调节点
		cb(cmd)
		return
	}
}
