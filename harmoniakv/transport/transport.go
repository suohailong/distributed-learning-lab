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
			ret: make(chan struct{}, 100),
		}
	})
}

type callback func(resp interface{})

type Transporter interface {
}

type defaultTransporter struct {
	ret chan struct{}
}

func collectResp(result interface{}) {
	tran.ret <- struct{}{}
}

func (d *defaultTransporter) send(nodes []*node.Node, cmd *node.KvCommand) {
	for _, peer := range nodes {
		n := peer
		_ = ants.Submit(
			func() {
				if n.GetId() == config.LocalId() {
					logrus.Debugf("local node: %v", n)
					n.HandleMsg(cmd, collectResp)
					// continue
					return
				}
				d.writeRemote(n, cmd, collectResp)
			})
	}
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

func Send(nodes []*node.Node, req *node.KvCommand) {
	tran.send(nodes, req)
}

func WaitResp(n int) []interface{} {
	//TODO: 在读取响应返回给调用方之后，状态机会等待一小段时间以接收任何未完成的响应。如果任何响应中返回了过期的版本，协调器将使用最新版本更新这些节点。这个过程被称为读修复（read repair），因为它在一个机会时机修复了错过最新更新的副本，并减轻了反熵协议的负担。
	return tran.waitResp(n)
}

func (d *defaultTransporter) writeRemote(n *node.Node, message interface{}, cb callback) {
	// 这里要重试

	// 获取client
	// 发送消息
	logrus.Debugf("send messge: [%v] to target: %v", message, n)

	cb(n.GetId())
}
