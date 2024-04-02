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
	Term uint64
}

type Node struct {
	Id    uint64
	Index uint64
}

type Raft struct {
	Id uint64
	sync.Mutex
	*RaftState
	State int
	Nodes []*Node
	Msgs  []*raftpb.Message
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

func (r *Raft) handleVote(m *raftpb.Message) {
	if m.Term < r.Term {
		// 拒绝
		r.send(&raftpb.Message{
			Term:   r.Term,
			From:   r.Id,
			To:     m.From,
			Type:   raftpb.MsgVoteResp,
			Reject: true,
		})
	} else if m.Term == r.Term {
		// 判断日志
	} else {
		// 同意
		r.send(&raftpb.Message{
			Term:   r.Term,
			From:   r.Id,
			To:     m.From,
			Type:   raftpb.MsgVoteResp,
			Reject: false,
		})
	}
}

func (r *Raft) handleHub() {
	r.becomeCandidate()
	for _, node := range r.Nodes {
		if node.Id == r.Id {
			continue
		}
		r.send(&raftpb.Message{
			Term: r.Term,
			To:   node.Id,
			Type: raftpb.MsgVote,
		})
	}
}

func (r *Raft) becomeCandidate() {
	r.Lock()
	defer r.Unlock()
	r.Term++
	r.State = CADIDATE
}

func (r *Raft) send(m *raftpb.Message) {
	r.Msgs = append(r.Msgs, m)
}
