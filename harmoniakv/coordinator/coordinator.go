package coordinator

import (
	v1 "distributed-learning-lab/harmoniakv/api/v1"
	"distributed-learning-lab/harmoniakv/cluster"
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/node"
	"distributed-learning-lab/harmoniakv/node/version"
	"distributed-learning-lab/harmoniakv/transport"

	"github.com/sirupsen/logrus"
)

// type Metadata struct {
// 	VersionVector *version.Vector
// }

type Coordinator interface {
	HandleGet(key []byte) ([]*v1.Object, error)
	HandlePut(key []byte, value *version.Value) error
}

type defaultCoordinator struct {
	cluster cluster.Cluster
}

func (d *defaultCoordinator) isInPrimaryList(replicas []*node.Node, nodeId string) int {
	for index, n := range replicas {
		if n.GetId() == nodeId {
			return index
		}
	}
	return -1
}

func (d *defaultCoordinator) HandleGet(key []byte) ([]*v1.Object, error) {
	//1. 协调节点从该key的首选列表中排名最高的N个可达节点请求该key的所有数据版本
	nodes := d.cluster.GetReplicas(key, config.Replicas())
	// 2. 判断自己是否为首选节点,如果不是直接转发请求,到指定节点
	index := d.isInPrimaryList(nodes, config.LocalId())
	if index < 0 {
		// 选择第一转发请求
		primaryNode := nodes[0]
		transport.Send(primaryNode, &node.KvCommand{})
	}

	for i := 0; i < len(nodes); i++ {
		transport.AsyncSend(nodes[i], &node.KvCommand{})
	}
	//3. 等待R个节点返回响应. 如果协调节点收到的响应少于R个,它将重试,直到收到R个响应
	resps := transport.WaitResp(config.ReadConcern())
	//4. 如果协调节点最终搜集到多个数据版本,它将返回它认为因果无关的版本
	logrus.Infof("get %s, resps: %v", key, resps)
	return []*v1.Object{}, nil
}

func (d *defaultCoordinator) HandlePut(key []byte, value *version.Value) error {
	// 获取节点及副本，并跳过那些访问不到的节点
	nodes := d.cluster.GetReplicas(key, config.Replicas())
	if len(nodes) < config.Replicas() {
		// 可用节点小于replicas
		// 集群不可用
	}
	// 1. 判断自己是否为首选节点,如果不是直接转发请求,到指定节点
	index := d.isInPrimaryList(nodes, config.LocalId())
	if index < 0 {
		// 选择第一转发请求
		primaryNode := nodes[0]
		_, err := transport.Send(primaryNode, &node.KvCommand{})
		if err != nil {
			logrus.Errorf("send to primary node error: %v", err)
			return err
		}
	}
	// 2. 生成该key对应的版本向量, 写入本地数据库
	acceptValue, err := transport.Send(nodes[index], &node.KvCommand{
		Command: node.PUT,
		Key:     key,
		Value:   *value,
	})
	if err != nil {
		logrus.Errorf("send to primary node error: %v", err)
		return err
	}

	otherNodes := []*node.Node{}
	otherNodes = append(otherNodes, nodes[:index]...)
	otherNodes = append(otherNodes, nodes[index+1:]...)
	// 3. 发送给其他N-1个可达的副本. 不可达的副本采用Hinted Handoff的方式实现副本一致性
	for i := 0; i < len(otherNodes); i++ {
		transport.AsyncSend(nodes[i], &node.KvCommand{
			Command: node.PUT,
			Key:     key,
			Value:   acceptValue.(version.Value),
		})
	}
	transport.WaitResp(config.WriteConcern())
	// 4. 等待w个个响应, 如果收到的响应少于w个,则重试,直到收到w个响应

	// 5. 返回客户端

	return nil
}
