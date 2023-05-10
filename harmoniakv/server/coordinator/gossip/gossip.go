package gossip

type Gossip interface {
	Start()
	Stop()
	Join(seedNodeAddress string)
	UpdateState(key string, value interface{})
	GetState(key string) (interface{}, bool)
	RegisterEventHandler(handler EventHandler)
}

type EventHandler func(event Event)

type Event struct {
	Type   string
	Key    string
	Value  interface{}
	NodeID string
}
