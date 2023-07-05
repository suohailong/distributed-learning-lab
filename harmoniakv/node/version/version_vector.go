package version

import v1 "distributed-learning-lab/harmoniakv/api/v1"

// 版本向量
type Vector struct {
	Versions map[string]uint64
}

func NewVector(versions map[string]uint64) *Vector {
	return &Vector{
		Versions: versions,
	}
}

func (ve *Vector) Increment(nodeId string) {
	if _, ok := ve.Versions[nodeId]; ok {
		ve.Versions[nodeId]++
	} else {
		ve.Versions[nodeId] = 1
	}
}

func (ve *Vector) Compare(in *Vector) {

}

// 版本向量值
type Value struct {
	KeyValue      *v1.Object
	VersionVector *Vector
}

func (va *Value) Equal() {}

func (va *Value) HasCode() int {}
