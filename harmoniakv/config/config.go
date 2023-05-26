package config

import "sync"

var defaultConfig *config
var onece sync.Once
var mutex sync.Mutex

type config struct {
	Replicas    uint32
	ReadConcern uint32
}

func init() {
	onece.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()
		defaultConfig = &config{
			Replicas: 1,
		}
	})
}

func ReadConcern() int {
	return int(defaultConfig.ReadConcern)
}

func Replicas() int {
	return int(defaultConfig.Replicas)
}
func LocalId() string {
	//TODO
	return ""
}
