package memory

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// DisjointSet implements disjoint set structure
type DisjointSet struct {
	SetSize uint32
	Parent  []uint32
	Rank    []uint32
	HashMap map[interface{}]uint32
}

// NewDisjointSet returnes a reference to a new istance of persisted disjoint set
func NewDisjointSet() DisjointSet {
	// const CAPACITY int = 1000000
	return DisjointSet{
		SetSize: 0,
		Parent:  []uint32{},
		Rank:    []uint32{},
		HashMap: map[interface{}]uint32{},
	}
}

// Size returnes the number of elements in the set
func (d DisjointSet) Size() uint32 {
	return d.SetSize
}

// GetHashMap returnes the set hashmap
func (d DisjointSet) GetHashMap() map[interface{}]uint32 {
	return d.HashMap
}

// GetParent returnes parent based on the passed tag
func (d DisjointSet) GetParent(tag uint32) uint32 {
	return d.Parent[tag]
}

// MakeSet creates a new set based adding the parameter passed as argument to the set
func (d DisjointSet) MakeSet(x interface{}) {
	if _, ok := d.HashMap[x]; ok {
		return
	}

	d.HashMap[x] = d.SetSize
	d.Parent = append(d.Parent, d.SetSize)
	d.Rank = append(d.Rank, 0)
	d.SetSize = d.SetSize + 1
}

// Find returnes the value of the set required as argument to the function
func (d DisjointSet) Find(x interface{}) (uint32, error) {
	pos, ok := d.HashMap[x]
	if !ok {
		return 0, errors.New("Element not found")
	}
	return d.FindInternal(d.Parent, pos), nil
}

// FindInternal recursively search for the element of depth n in the set
func (d DisjointSet) FindInternal(p []uint32, n uint32) uint32 {
	if p[n] != n {
		parent := p[n]
		p[n] = d.FindInternal(p, parent)
		return p[n]
	}
	return n
}

// Union returnes the common set to the elements passed as arguments
func (d DisjointSet) Union(x, y interface{}) (uint32, error) {
	var (
		xRoot,
		yRoot,
		xRank,
		yRank uint32
	)

	xRoot, err := d.Find(x)
	if err != nil {
		logger.Error("Disjoint Set", err, logger.Params{})
		return 0, err
	}
	xRank = d.Rank[xRoot]
	yRoot, err = d.Find(y)
	if err != nil {
		logger.Error("Disjoint Set", err, logger.Params{})
		return 0, err
	}
	yRank = d.Rank[yRoot]

	if xRoot == yRoot {
		return xRoot, nil
	}
	if xRank > yRank {
		d.Parent[yRoot] = xRoot
		return xRoot, nil
	}
	d.Parent[xRoot] = yRoot
	if xRank == yRank {
		d.Rank[yRoot]++
	}
	return yRoot, nil
}

// Finalize parses the entire set
func (d DisjointSet) Finalize() {
	for i := 0; uint32(i) < d.SetSize; i++ {
		d.FindInternal(d.Parent, uint32(i))
	}
}
