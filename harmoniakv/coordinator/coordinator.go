package coordinator

import (
	v1 "distributed-learning-lab/harmoniakv/api/v1"
	"distributed-learning-lab/harmoniakv/cluster"
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/coordinator/version"
	"distributed-learning-lab/harmoniakv/node"
	"distributed-learning-lab/harmoniakv/transport"

	"github.com/sirupsen/logrus"
)

type Metadata struct {
	VersionVector *version.Vector
}

type Coordinator interface {
	HandleGet(key []byte) ([]*v1.Object, error)
	HandlePut(meta *Metadata, key []byte) error
}

type defaultCoordinator struct {
	cluster cluster.Cluster
}

func (d *defaultCoordinator) isInPrimaryList(replicas []*node.Node, nodeId string) bool {
	for _, n := range replicas {
		if n.GetId() == nodeId {
			return true
		}
	}
	return false
}

func (d *defaultCoordinator) HandleGet(key []byte) ([]*v1.Object, error) {
	//1. 协调节点从该key的首选列表中排名最高的N个可达节点请求该key的所有数据版本
	nodes := d.cluster.GetReplicas(key, config.Replicas())
	//2. 判断自己是否是该key的协调节点, 不是则转发请求
	transport.Send(nodes, key)
	//3. 等待R个节点返回响应. 如果协调节点收到的响应少于R个,它将重试,直到收到R个响应
	resps := transport.WaitResp(config.ReadConcern())
	//4. 如果协调节点最终搜集到多个数据版本,它将返回它认为因果无关的版本
	logrus.Infof("get %s, resps: %v", key, resps)

	return []*v1.Object{}, nil
}

func (d *defaultCoordinator) HandlePut(meta *Metadata, key []byte) error {
	nodes := d.cluster.GetReplicas(key, config.Replicas())
	if len(nodes) < config.Replicas() {
		// 可用节点小于replicas
	}
	// 1. 判断自己是否为首选节点,如果不是直接转发请求,到指定节点
	if !d.isInPrimaryList(nodes, config.LocalId()) {
		// 选择第一转发请求
		primaryNode := nodes[0]
		transport.Send([]*node.Node{primaryNode}, []byte("ss"))
	}
	// 2. 生成该key对应的版本向量, 写入本地数据库
	meta.VersionVector.Increment(config.LocalId())
	// 3. 发送给其他N-1个可达的副本. 不可达的副本采用Hinted Handoff的方式实现副本一致性
	transport.Send()
	// 4. 等待w个个响应, 如果收到的响应少于w个,则重试,直到收到w个响应
	// 5. 返回客户端
}
