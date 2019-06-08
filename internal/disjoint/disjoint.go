package disjoint

// DisjointSet implements disjoint set structure
type DisjointSet interface {
	Size() int
	GetHashMap() map[interface{}]int
	GetParent(int) int
	MakeSet(interface{})
	Find(interface{}) (int, error)
	FindInternal([]int, int) int
	Union(interface{}, interface{}) (int, error)
	Finalize()
}
