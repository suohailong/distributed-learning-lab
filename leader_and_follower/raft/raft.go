package raft

import (
	"sync"

	"distributed-learning-lab/util/log"

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
	State           int
	Msgs            chan *raftpb.Message
	TimeOut         int
	electionElapsed int
	Prs             map[uint64]*Node

	// 记录我投票的任期号
	voteTerm uint64
	// 记录我投票的节点
	voteId uint64

	// 记录票数
	voteNum int
	votePrs map[uint64]bool
}

func NewRaft(id uint64, timeout int, prs []uint64) *Raft {
	r := &Raft{
		Id:      id,
		TimeOut: timeout,
		Msgs:    make(chan *raftpb.Message, 10),
		RaftState: &RaftState{
			Term: 0,
		},
		Prs:     make(map[uint64]*Node),
		votePrs: make(map[uint64]bool),
		State:   FOLLOWER,
	}

	for _, pr := range prs {
		r.Prs[pr] = &Node{
			Id:    pr,
			Index: 0,
		}
	}

	return r
}

func (r *Raft) Tick() {
	switch r.State {
	case LEADER:
		// TODO: 发送心跳
	default:
		r.electionElapsed++
		if r.electionElapsed >= r.TimeOut {
			r.Step(&raftpb.Message{
				Type: raftpb.MsgHup,
			})
		}
	}
}

// 用来驱动 Raft 状态机的主要方法
func (r *Raft) Step(m *raftpb.Message) error {
	switch m.Type {
	case raftpb.MsgHup:
		// TODO: 开始选举
		r.handleHub()
	case raftpb.MsgVote:
		// TODO: 处理投票请求
		r.handleVote(m)
	case raftpb.MsgVoteResp:
		r.handleVoteResponse(m)
	}

	return nil
}

func (r *Raft) handleVoteResponse(m *raftpb.Message) {
	log.Debugf("node %d receive vote response message from node %d", r.Id, m.From)
	if m.Reject {
		return
	}
	if ok := r.votePrs[m.From]; !ok {
		log.Debugf("node %d recive vote From node %d", r.Id, m.From)
		r.votePrs[m.From] = true
		r.voteNum++
		if r.voteNum > len(r.votePrs)/2 {
			r.becomeLeader()
		}
	}
}

func (r *Raft) handleVote(m *raftpb.Message) {
	log.Debugf("node %d receive vote message from node %d", r.Id, m.From)
	if r.Term < m.Term {
		// 拒绝
		r.send(&raftpb.Message{
			Term: r.Term,
			From: r.Id,
			To:   m.From,
			Type: raftpb.MsgVoteResp,
		})
		r.becomeFlowwer()
	} else if m.Term == r.Term {
		if r.voteTerm == m.Term {
			// 已经给当前任期投过票, 而且节点也是同一个
			if r.voteId == m.From {
				log.Debugf("node %d already vote to node %d", r.Id, m.From)
				r.send(&raftpb.Message{
					Term: r.Term,
					From: r.Id,
					To:   m.From,
					Type: raftpb.MsgVoteResp,
				})
			}
			r.becomeFlowwer()
		} else {
			// 如果没给当前任期投过票， 比较日志
			// 日志比较新， 投票
		}
	}
	log.Debugf("node %d reject vote to node %d", r.Id, m.From)
	r.send(&raftpb.Message{
		Term:   r.Term,
		From:   r.Id,
		To:     m.From,
		Type:   raftpb.MsgVoteResp,
		Reject: true,
	})
}

func (r *Raft) handleHub() {
	log.Debugf("node %d start campaign", r.Id)
	r.becomeCandidate()
	for _, node := range r.Prs {
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

func (r *Raft) becomeLeader() {
	r.Lock()
	defer r.Unlock()
	r.State = LEADER
	log.Debugf("node %d become leader", r.Id)
}

func (r *Raft) becomeCandidate() {
	r.Lock()
	defer r.Unlock()
	r.Term++
	r.State = CADIDATE
	log.Debugf("node %d become candidate", r.Id)
}

func (r *Raft) becomeFlowwer() {
	r.Lock()
	defer r.Unlock()
	r.State = FOLLOWER
	log.Debugf("node %d become follower", r.Id)
}

func (r *Raft) send(m *raftpb.Message) {
	log.Debugf("node %d send message to node %d, msgType: %v", r.Id, m.To, m.Type)
	r.Msgs <- m
}

func (r *Raft) GetMsgs() chan *raftpb.Message {
	return r.Msgs
}
