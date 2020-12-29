package memory

import (
	"sync"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// DisjointSet implements disjoint set structure
type DisjointSet struct {
	size    uint64
	parent  []uint64
	rank    []uint64
	hashMap sync.Map
}

// NewDisjointSet returns a reference to a new istance of persisted disjoint set
func NewDisjointSet() DisjointSet {
	const CAPACITY int = 2147483647
	return DisjointSet{
		size:    0,
		parent:  []uint64{},
		rank:    []uint64{},
		hashMap: sync.Map{},
	}
}

// GetSize returns the number of elements in the set
func (d *DisjointSet) GetSize() uint64 {
	return d.size
}

// GetHashMap returns the set hashmap
func (d *DisjointSet) GetHashMap() *sync.Map {
	return &d.hashMap
}

// GetParent returns parent based on the passed tag
func (d *DisjointSet) GetParent(tag uint64) uint64 {
	return d.parent[tag]
}

// GetParents returns parents list
func (d *DisjointSet) GetParents() []uint64 {
	return d.parent
}

// GetRanks returns ranks list
func (d *DisjointSet) GetRanks() []uint64 {
	return d.rank
}

// MakeSet creates a new set based adding the parameter passed as argument to the set
func (d *DisjointSet) MakeSet(x interface{}) {
	if _, ok := d.hashMap.Load(x); ok {
		return
	}

	d.hashMap.Store(x, d.size)
	d.parent = append(d.parent, d.size)
	d.rank = append(d.rank, 0)
	d.size = d.size + 1
}

// PrepareMakeSet is a method declared for the persistent version of the package. Fallback to MakeSet
func (d *DisjointSet) PrepareMakeSet(x interface{}, clusters *dgraph.Clusters, lock *sync.RWMutex) {
	d.MakeSet(x)
}

// Find returns the value of the set required as argument to the function
func (d *DisjointSet) Find(x interface{}) (uint64, error) {
	pos, ok := d.hashMap.Load(x)
	if !ok {
		return 0, errorx.ErrNotFound
	}
	return d.FindInternal(d.parent, pos.(uint64)), nil
}

// FindInternal recursively search for the element of depth n in the set
func (d *DisjointSet) FindInternal(p []uint64, n uint64) uint64 {
	if p[n] != n {
		parent := p[n]
		p[n] = d.FindInternal(p, parent)
		return p[n]
	}
	return n
}

// Union returns the common set to the elements passed as arguments
func (d *DisjointSet) Union(x, y interface{}) (uint64, error) {
	var (
		xRoot,
		yRoot,
		xRank,
		yRank uint64
	)

	xRoot, err := d.Find(x)
	if err != nil {
		logger.Error("Disjoint Set", err, logger.Params{})
		return 0, err
	}
	xRank = d.rank[xRoot]
	yRoot, err = d.Find(y)
	if err != nil {
		logger.Error("Disjoint Set", err, logger.Params{})
		return 0, err
	}
	yRank = d.rank[yRoot]

	if xRoot == yRoot {
		return xRoot, nil
	}
	if xRank > yRank {
		d.parent[yRoot] = xRoot
		return xRoot, nil
	}
	d.parent[xRoot] = yRoot
	if xRank == yRank {
		d.rank[yRoot]++
	}
	return yRoot, nil
}

// PrepareUnion is a method declared for the persistent version of the package. Fallback to Union
func (d *DisjointSet) PrepareUnion(x, y interface{}, clusters *dgraph.Clusters, lock *sync.RWMutex) (uint64, error) {
	return d.Union(x, y)
}

// Finalize parses the entire set
func (d *DisjointSet) Finalize() {
	for i := 0; uint64(i) < d.size; i++ {
		d.FindInternal(d.parent, uint64(i))
	}
}
