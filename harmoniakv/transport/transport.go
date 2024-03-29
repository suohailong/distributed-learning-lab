package transport

import (
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/node"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
)

var tran *defaultTransporter

var onece sync.Once
var mutex sync.Mutex

func init() {
	onece.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()
		tran = &defaultTransporter{
			ret: make(chan interface{}, 100),
		}
	})
}

type callback func(resp interface{})

type Transporter interface {
}

type defaultTransporter struct {
	ret  chan interface{}
	pool *ants.Pool
}

func collectResp(result interface{}, err error) {
	if err != nil {
		tran.ret <- result
	} else {
		tran.ret <- err
	}
}

// TODO: 这里也不知道返回什么
func (d *defaultTransporter) send(n *node.Node, cmd *node.KvCommand) (interface{}, error) {
	if n.GetId() == config.LocalId() {
		logrus.Debugf("local node: %v", n)
		return n.HandleCommand(cmd)
	}
	// TODO:  传递消息到其他节点
	return d.writeRemote(n, cmd)
}

func (d *defaultTransporter) asend(n *node.Node, cmd *node.KvCommand) error {
	return d.pool.Submit(
		func() {
			result, err := d.send(n, cmd)
			collectResp(result, err)
		})
}

func (d *defaultTransporter) waitResp(n int) []interface{} {
	resps := []interface{}{}
	i := 0
	for resp := range tran.ret {
		if i == n {
			break
		}
		resps = append(resps, resp)
		i++
	}
	logrus.Debugf("wait resp: %v", resps)
	return resps
}

func Send(node *node.Node, req *node.KvCommand) (interface{}, error) {
	return tran.send(node, req)
}

func AsyncSend(node *node.Node, req *node.KvCommand) {
	tran.asend(node, req)
}

func WaitResp(n int) []interface{} {
	//TODO: 在读取响应返回给调用方之后，状态机会等待一小段时间以接收任何未完成的响应。如果任何响应中返回了过期的版本，协调器将使用最新版本更新这些节点。这个过程被称为读修复（read repair），因为它在一个机会时机修复了错过最新更新的副本，并减轻了反熵协议的负担。
	return tran.waitResp(n)
}

// TODO: 这里也不知道返回什么
func (d *defaultTransporter) writeRemote(n *node.Node, message interface{}) (interface{}, error) {
	// 这里要重试

	// 获取client
	// 发送消息
	logrus.Debugf("send messge: [%v] to target: %v", message, n)

	return nil, nil
}
