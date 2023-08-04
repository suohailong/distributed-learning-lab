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

	Store storage.Storage
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
// TODO: 这里返回值暂且不知道填什么
func (n *Node) HandleCommand(cmd *KvCommand) (interface{}, error) {
	logrus.Debugf("handle local msg: %v", cmd)
	if cmd.Command == GET {
		// 3. 从本地数据库读取该key的所有版本
		// n.store.Get(cmd.Key)
		logrus.Debugf("get key: %s", cmd.Key)
		// 4. 将所有版本返回给协调节点
		return nil, nil
	} else if cmd.Command == PUT {
		logrus.Debugf("put key: %s, value: %v", cmd.Key, cmd.Value)
		// 2. 生成该key对应的版本向量, 写入本地数据库
		cmd.Value.VersionVector.Increment(n.ID)
		// 3. 写入对应的值到本地
		n.Store.Put(cmd.Key, cmd.Value)
		return nil, nil
	}
	return nil, nil
}
