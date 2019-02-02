package disjoint

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

type DisjointSet struct {
	SetSize int
	Parent  []int
	Rank    []int
	HashMap map[interface{}]int
}

func NewDisjointSet() *DisjointSet {
	// const CAPACITY int = 1000000
	return &DisjointSet{
		SetSize: 0,
		Parent:  []int{},
		Rank:    []int{},
		HashMap: map[interface{}]int{},
	}
}

func (d *DisjointSet) Size() int {
	return d.SetSize
}

func (d *DisjointSet) MakeSet(x interface{}) {
	if _, ok := d.HashMap[x]; ok {
		return
	}

	d.HashMap[x] = d.SetSize
	d.Parent = append(d.Parent, d.SetSize)
	d.Rank = append(d.Rank, 0)
	d.SetSize = d.SetSize + 1
}

func (d *DisjointSet) Find(x interface{}) (int, error) {
	pos, ok := d.HashMap[x]
	if !ok {
		return 0, errors.New("Element not found")
	}
	return d.FindInternal(d.Parent, pos), nil
}

func (d *DisjointSet) FindInternal(p []int, n int) int {
	if p[n] != n {
		parent := p[n]
		p[n] = d.FindInternal(p, parent)
		return p[n]
	}
	return n
}

func (d *DisjointSet) Union(x, y interface{}) (int, error) {
	var xRoot int
	var yRoot int
	var xRank int
	var yRank int

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

func (d *DisjointSet) Finalize() {
	for i := 0; i < d.SetSize; i++ {
		d.FindInternal(d.Parent, i)
	}
}
