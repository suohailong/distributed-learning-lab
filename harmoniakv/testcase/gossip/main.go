package main

import (
	"log"
)

func Diff(fromState, toState map[string]string) map[string]string {
	diff := make(map[string]string)
	for key, value := range fromState {
		if butu, ok := toState[key]; !ok || value != butu {
			diff[key] = value
		}
	}
	return diff
}

type NodeId string
type NodeState map[string]string // Assuming NodeState is a map of strings. Please replace this with actual type
type GossipStateMessage struct {
	NodeStates map[NodeId]NodeState
}

type Gossip struct {
	ClusterMetadata map[NodeId]NodeState
}

func (g *Gossip) HandleGossipRequest(gossipStateMessage GossipStateMessage) {
	log.Printf("Received gossip state message %+v\n", gossipStateMessage)
	log.Printf("Current state %+v\n", g.ClusterMetadata)
	// log.Printf("Merging state from %s\n", conn.RemoteAddr().String())
	g.Merge(gossipStateMessage.NodeStates)

	log.Printf("after merge Current state %+v\n", g.ClusterMetadata)

	diff := Delta(g.ClusterMetadata, gossipStateMessage.NodeStates)
	log.Printf("Sending diff %+v\n", diff)
	diffResponse := GossipStateMessage{NodeStates: diff}

	log.Printf("Sending diff response %+v\n", diffResponse)
	// conn.Write(diffResponseBytes)
}

func Delta(fromMap, toMap map[NodeId]NodeState) map[NodeId]NodeState {
	delta := make(map[NodeId]NodeState)
	for key, fromState := range fromMap {
		toState, exists := toMap[key]
		if !exists {
			// 本地存在 远端不存在
			delta[key] = fromState
		} else {
			diffStates := Diff(fromState, toState) // Assuming Diff function exists
			if len(diffStates) > 0 {
				delta[key] = diffStates
			}
		}
	}
	return delta
}

func (g *Gossip) Merge(otherState map[NodeId]NodeState) {
	diff := Delta(otherState, g.ClusterMetadata)
	for diffKey, diffValue := range diff {
		currentState, exists := g.ClusterMetadata[diffKey]
		if !exists {
			g.ClusterMetadata[diffKey] = diffValue
		} else {
			for k, v := range diffValue { // Assuming NodeState is a map
				currentState[k] = v
			}
		}
	}
}

func main() {
	g := Gossip{
		ClusterMetadata: map[NodeId]NodeState{
			"node1": map[string]string{
				"counter": "0",
			},
			"node2": map[string]string{
				"counter": "0",
			},
			"node3": map[string]string{
				"counter": "0",
			},
		},
	}
	g.HandleGossipRequest(GossipStateMessage{
		NodeStates: map[NodeId]NodeState{
			"node1": map[string]string{
				"counter": "2",
			},
			"node2": map[string]string{
				"counter": "3",
			},
			// "node3": map[string]string{
			// 	"counter": "0",
			// },
		},
	})
}
