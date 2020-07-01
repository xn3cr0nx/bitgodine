package disjoint

import (
	"sync"
)

// DisjointSet implements disjoint set structure
type DisjointSet interface {
	GetSize() uint64
	GetHeight() int32
	GetHashMap() *sync.Map
	GetParent(uint64) uint64
	MakeSet(interface{})
	Find(interface{}, *sync.Map) (uint64, error)
	FindInternal([]uint64, uint64, *sync.Map) uint64
	Union(interface{}, interface{}) (uint64, error)
	UpdateHeight(int32) error
	PrepareMakeSet(interface{}, *sync.Map)
	PrepareUnion(interface{}, interface{}, *sync.Map) (uint64, error)
	BulkUpdate(*sync.Map) error
	Finalize()
}
