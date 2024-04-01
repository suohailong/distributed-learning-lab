package raft

import (
	"sync"

	"go.etcd.io/etcd/raft/v3/raftpb"
)

const (
	MSG_HUP = iota
	MSG_VOTE
)

const (
	FOLLOWER = iota
	LEADER
	CADIDATE
)

type Message struct {
	MsgType int
}

type RaftState struct {
	Term int64
}

type Raft struct {
	sync.Mutex
	*RaftState
	State int
}

func NewRaft() *Raft {
	return nil
}

// 用来驱动 Raft 状态机的主要方法
func (r *Raft) Step(m *raftpb.Message) error {
	switch m.Type {
	case MSG_HUP:
		// TODO: 开始选举
		r.handleHub()
	case MSG_VOTE:
		// TODO: 处理投票请求
		r.handleVote()
	}
	return nil
}

func (r *Raft) handleVote() {}

func (r *Raft) handleHub() {
	r.becomeCandidate()
	r.send(&raftpb.Message{Type: raftpb.MsgPreVote})
}

func (r *Raft) becomeCandidate() {
	r.Lock()
	defer r.Unlock()
	r.State = CADIDATE
}

func (r *Raft) send(m *raftpb.Message) {}
