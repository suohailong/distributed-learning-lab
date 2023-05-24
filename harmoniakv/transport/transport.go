package transport

import (
	"distributed-learning-lab/harmoniakv/cluster"
	"distributed-learning-lab/harmoniakv/config"
	"sync"
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
	SendGet(nodes map[string]struct{})
}

type defaultTransporter struct {
	ret chan struct{}
}

func collectResp(result interface{}) {
	tran.ret <- struct{}{}
}

func Send(nodes map[string]struct{}, key []byte) {
	message := "aa"
	for id := range nodes {
		n := cluster.GetNodeByID(id)
		if id == config.LocalId() {
			n.HandleMsg(message, collectResp)
			continue
		}
		writeRemote(id, message, collectResp)
	}
}

func WaitResp(n int) []interface{} {
	resps := []interface{}{}
	i := 0
	for resp := range tran.ret {
		if i == n {
			break
		}
		resps = append(resps, resp)
	}
	return resps
}

func writeRemote(string, interface{}, callback) {}
