package disjoint

// DisjointSet implements disjoint set structure
type DisjointSet interface {
	Size() uint32
	GetHashMap() map[interface{}]uint32
	GetParent(uint32) uint32
	MakeSet(interface{})
	Find(interface{}) (uint32, error)
	FindInternal([]uint32, uint32) uint32
	Union(interface{}, interface{}) (uint32, error)
	Finalize()
}
