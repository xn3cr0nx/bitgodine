package disjoint

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	"github.com/btcsuite/btcutil"
)

type DisjointSet struct {
	SetSize uint
	Parent  []uint
	rank    []uint
	HashMap map[btcutil.Address]uint
}

func NewDisjointSet() *DisjointSet {
	const CAPACITY int = 1000000
	return &DisjointSet{
		SetSize: 0,
		Parent:  make([]uint, CAPACITY),
		rank:    make([]uint, CAPACITY),
		HashMap: make(map[btcutil.Address]uint, CAPACITY),
	}
}

func (d *DisjointSet) Size() uint {
	return d.SetSize
}

func (d *DisjointSet) MakeSet(x btcutil.Address) {
	if _, ok := d.HashMap[x]; ok {
		return
	}

	d.HashMap[x] = d.SetSize
	d.Parent = append(d.Parent, d.SetSize)
	d.rank = append(d.rank, 0)
	d.SetSize = d.SetSize + 1
}

func (d *DisjointSet) Find(x btcutil.Address) (uint, error) {
	pos, ok := d.HashMap[x]
	if !ok {
		return 0, errors.New("Address not found")
	}
	return d.FindInternal(d.Parent, pos), nil
}

func (d *DisjointSet) FindInternal(p []uint, n uint) uint {
	if p[n] != n {
		Parent := p[n]
		p[n] = d.FindInternal(p, Parent)
		return p[n]
	}
	return n
}

func (d *DisjointSet) Union(x, y btcutil.Address) (uint, error) {
	var xRoot uint
	var yRoot uint
	var xRank uint
	var yRank uint

	xRoot, err := d.Find(x)
	if err != nil {
		logger.Error("Finding in Union", err, logger.Params{})
		return 0, err
	}
	xRank = d.rank[xRoot]
	yRoot, err = d.Find(y)
	if err != nil {
		logger.Error("Finding in Union", err, logger.Params{})
		return 0, err
	}
	yRank = d.rank[yRoot]

	if xRoot == yRoot {
		return xRoot, nil
	}
	if xRank > yRank {
		d.Parent[yRoot] = xRoot
		return xRoot, nil
	} else {
		d.Parent[xRoot] = yRoot
		if xRank == yRank {
			d.rank[yRoot]++
		}
		return yRoot, nil
	}
}

func (d *DisjointSet) Finalize() {
	for i := uint(0); i < d.SetSize; i++ {
		d.FindInternal(d.Parent, i)
	}
}
