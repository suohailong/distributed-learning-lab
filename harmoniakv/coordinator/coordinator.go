package coordinator

import (
	v1 "distributed-learning-lab/harmoniakv/api/v1"
	"distributed-learning-lab/harmoniakv/cluster"
	"distributed-learning-lab/harmoniakv/config"
)

type Metadata struct{}

type Coordinator interface {
	HandleGet(key []byte) ([]*v1.Object, error)
	HandlePut(meta *Metadata, key []byte) error
}

type defaultCoordinator struct {
	cluster cluster.Cluster
}

func (d *defaultCoordinator) HandleGet(key []byte) ([]*v1.Object, error) {
	//1. 协调节点从该key的首选列表中排名最高的N个可达节点请求该key的所有数据版本
	nodes := d.cluster.GetReplicas(key, config.Replicas())
	//2. 判断自己是否是该key的协调节点, 不是则转发请求
	if n, ok := nodes[config.LocalId()]; !ok {
		d.transport.Send()
	}
	//3. 等待R个节点返回响应. 如果协调节点收到的响应少于R个,它将重试,直到收到R个响应
	//4. 如果协调节点最终搜集到多个数据版本,它将返回它认为因果无关的版本
}

func (d *defaultCoordinator) HandlePut(meta *Metadata, key []byte) error {
	// 1. 判断自己是否为首选节点,如果不是直接转发请求,到指定节点
	// 2. 生成该key对应的版本向量, 写入本地数据库
	// 3. 发送给其他N-1个可达的副本. 不可达的副本采用Hinted Handoff的方式实现副本一致性
	// 4. 等待w个个响应, 如果收到的响应少于w个,则重试,直到收到w个响应
	// 5. 返回客户端
}
