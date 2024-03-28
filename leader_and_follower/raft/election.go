package raft

type RaftState struct {
	Term int64
}

type Raft struct {
	*RaftState
	Node map[string]
}

func NewRaft() *Raft {
	return &Raft{&RaftState{0}}
}

func (r *Raft) Step(m *Message) {
}
