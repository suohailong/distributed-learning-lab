package raft

import (
	"math/rand"
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
	State            int
	Msgs             chan *raftpb.Message
	electionTimeout  int
	electionElapsed  int
	heartbeatElapsed int
	heartbeatTimeout int
	Prs              map[uint64]*Node

	// 记录我投票的任期号
	voteTerm uint64
	// 记录我投票的节点
	voteId uint64

	// 记录票数
	voteNum int
	// 记录我的谁给我投票
	votePrs map[uint64]bool
}

func NewRaft(id uint64, elecTimeout, heartbeatTimeout int, prs []uint64) *Raft {
	r := &Raft{
		Id:               id,
		electionTimeout:  elecTimeout,
		heartbeatTimeout: heartbeatTimeout,
		Msgs:             make(chan *raftpb.Message, 10),
		RaftState: &RaftState{
			Term: 0,
		},
		Prs:             make(map[uint64]*Node),
		votePrs:         make(map[uint64]bool),
		State:           FOLLOWER,
		electionElapsed: 0,
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
		r.heartbeatElapsed++
		if r.heartbeatElapsed >= r.heartbeatTimeout {
			log.Debugf("node: %d, send heartbeat, heartbeatElapsed: %d, heartbeatTimout: %d", r.Id, r.heartbeatElapsed, r.heartbeatTimeout)
			r.Step(&raftpb.Message{
				Type: raftpb.MsgBeat,
			})
		}

	default:
		r.electionElapsed++
		// 这里可能会发起多起选举， 下次选举开始之前要把上次选举的中间结果清楚， 比如本机投给的节点记录等
		if r.electionElapsed >= r.electionTimeout {
			log.Debugf("node: %d, timeout, electionElapsed: %d, timeout: %d", r.Id, r.electionElapsed, r.electionTimeout)
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
		r.handleHub()
	case raftpb.MsgVote:
		r.handleVote(m)
	case raftpb.MsgVoteResp:
		r.handleVoteResponse(m)
	case raftpb.MsgBeat:
		r.handleMsgBeat()
	case raftpb.MsgHeartbeat:
		r.handleMsgHeartbeat(m)
	case raftpb.MsgHeartbeatResp:
		r.handleHeartbeatResponse(m)
	}

	return nil
}

func (r *Raft) handleHeartbeatResponse(m *raftpb.Message) {
	log.Debugf("node %d receive heartbeat response message from node %d, term: %d", r.Id, m.From, m.Term)
}

func (r *Raft) handleMsgHeartbeat(m *raftpb.Message) {
	log.Debugf("node %d receive heartbeat message from node %d, term: %d, myterm: %d", r.Id, m.From, m.Term, r.Term)
	r.electionElapsed = 0
	r.send(&raftpb.Message{
		Type: raftpb.MsgHeartbeatResp,
		To:   m.From,
		From: r.Id,
	})
}

func (r *Raft) handleMsgBeat() {
	// r.electionElapsed = 0
	for _, node := range r.Prs {
		if node.Id == r.Id {
			continue
		}
		r.send(&raftpb.Message{
			From: r.Id,
			Term: r.Term,
			To:   node.Id,
			Type: raftpb.MsgHeartbeat,
		})
	}
}

func (r *Raft) handleVoteResponse(m *raftpb.Message) {
	log.Debugf("node %d receive vote response message from node %d, term: %d", r.Id, m.From, m.Term)
	if m.Term < r.Term {
		return
	}
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
	log.Debugf("node %d receive vote message from node %d, term: %d", r.Id, m.From, m.Term)
	if r.Term < m.Term {
		// 拒绝
		r.send(&raftpb.Message{
			Term: m.Term,
			From: r.Id,
			To:   m.From,
			Type: raftpb.MsgVoteResp,
		})
		r.voteTerm = m.Term
		r.voteId = m.From
		r.becomeFlowwer(m.Term)
		return
	} else if m.Term == r.Term {
		if r.voteTerm == m.Term {
			// 已经给当前任期投过票, 而且节点也是同一个
			if r.voteId == m.From {
				log.Debugf("node %d already vote to node %d", r.Id, m.From)
				r.send(&raftpb.Message{
					Term: m.Term,
					From: r.Id,
					To:   m.From,
					Type: raftpb.MsgVoteResp,
				})
			}
			r.voteTerm = m.Term
			r.voteId = m.From
			r.becomeFlowwer(m.Term)
			return
		} else {
			// 如果没给当前任期投过票， 比较日志
			// 日志比较新， 投票
		}
	}
	log.Debugf("node %d reject vote to node %d, term: %d", r.Id, m.From, m.Term)
	r.send(&raftpb.Message{
		Term:   r.Term,
		From:   r.Id,
		To:     m.From,
		Type:   raftpb.MsgVoteResp,
		Reject: true,
	})
}

func (r *Raft) handleHub() {
	if r.State == LEADER {
		return
	}
	log.Debugf("node %d start campaign", r.Id)
	r.becomeCandidate()
	for _, node := range r.Prs {
		if node.Id == r.Id {
			continue
		}
		r.send(&raftpb.Message{
			From: r.Id,
			Term: r.Term,
			To:   node.Id,
			Type: raftpb.MsgVote,
		})
	}
}

func (r *Raft) becomeLeader() {
	r.State = LEADER
	r.electionElapsed = 0
	r.heartbeatElapsed = 0

	randomInt := rand.Intn(r.electionTimeout)
	r.electionTimeout = r.electionTimeout + randomInt
	log.Debugf("node %d become leader, electionElapsed: %d, term: %d", r.Id, r.electionElapsed, r.Term)
}

func (r *Raft) becomeCandidate() {
	r.voteNum = 0
	r.votePrs = make(map[uint64]bool)

	r.Term++

	r.voteId = r.Id
	r.voteNum++
	r.voteTerm = r.Term
	r.votePrs[r.Id] = true

	r.State = CADIDATE
	r.electionElapsed = 0
	r.heartbeatElapsed = 0

	randomInt := rand.Intn(r.electionTimeout)
	r.electionTimeout = r.electionTimeout + randomInt
	log.Debugf("node %d become candidate, electionElapsed: %d, term: %d", r.Id, r.electionElapsed, r.Term)
}

func (r *Raft) becomeFlowwer(term uint64) {
	r.Term = term

	r.State = FOLLOWER
	r.voteNum = 0
	r.votePrs = make(map[uint64]bool)

	r.electionElapsed = 0
	r.heartbeatElapsed = 0
	randomInt := rand.Intn(r.electionTimeout)
	r.electionTimeout = r.electionTimeout + randomInt
	log.Debugf("node %d become flowwer, electionElapsed: %d: term: %d", r.Id, r.electionElapsed, r.Term)
}

func (r *Raft) send(m *raftpb.Message) {
	log.Debugf("node %d send message to node %d, msgType: %v, term: %d", r.Id, m.To, m.Type, m.Term)
	r.Msgs <- m
}

func (r *Raft) GetMsgs() chan *raftpb.Message {
	return r.Msgs
}
