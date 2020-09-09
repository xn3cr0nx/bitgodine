package disk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// DisjointSet implements disjoint set logic in a persistent way using key value storage
type DisjointSet struct {
	size    uint64
	height  int32
	parent  []uint64
	rank    []uint64
	hashMap sync.Map
	storage storage.KV
}

// NewDisjointSet creates a new instance of DisjointSet
func NewDisjointSet(db storage.KV, disk, memory bool) (d DisjointSet, err error) {
	const CAPACITY uint64 = 2147483647
	d = DisjointSet{
		size:    0,
		height:  0,
		parent:  make([]uint64, CAPACITY),
		rank:    make([]uint64, CAPACITY),
		hashMap: sync.Map{},
		storage: db,
	}
	return
}

// RestorePersistentSet initialize the disjoint set with the persisted state
func RestorePersistentSet(d *DisjointSet) (err error) {
	size, err := d.GetStoredSize()
	d.size = size
	logger.Info("Persistent Disjoint Set", "Restoring the clusters", logger.Params{"size": d.size})

	height, err := d.GetStoredHeight()
	if err != nil {
		return
	}
	d.height = height

	for i := uint64(0); i < size; i++ {
		p, e := d.storage.Read(fmt.Sprintf("p%d", i))
		if e != nil {
			return e
		}
		d.parent[i] = binary.LittleEndian.Uint64(p)

		r, e := d.storage.Read(fmt.Sprintf("r%d", i))
		if e != nil {
			return e
		}
		d.rank[i] = binary.LittleEndian.Uint64(r)
	}

	addresses, err := d.storage.ReadKeysWithPrefix("addr")
	for _, address := range addresses {
		cluster, e := d.storage.Read(address)
		if e != nil {
			return e
		}
		d.hashMap.Store(address[4:], binary.LittleEndian.Uint64(cluster))
	}

	return
}

// GetSize returnes the number of elements in the set
func (d *DisjointSet) GetSize() uint64 {
	return d.size
}

// GetHeight returnes the number of elements in the set
func (d *DisjointSet) GetHeight() int32 {
	return d.height
}

// GetStoredHeight returnes the number of elements in the set
func (d *DisjointSet) GetStoredHeight() (int32, error) {
	height, err := d.storage.Read("height")
	if err != nil {
		return 0, err
	}
	h := int32(binary.LittleEndian.Uint64(height))
	return h, nil
}

// GetStoredSize returnes the number of elements in the set
func (d *DisjointSet) GetStoredSize() (uint64, error) {
	size, err := d.storage.Read("size")
	if err != nil {
		return 0, err
	}
	s := binary.LittleEndian.Uint64(size)
	return s, nil
}

// GetStoredParents returnes the number of elements in the set
func (d *DisjointSet) GetStoredParents() (parents []uint64, err error) {
	p, err := d.storage.ReadPrefix("p")
	for _, parent := range p {
		parents = append(parents, binary.LittleEndian.Uint64(parent))
	}
	return
}

// GetStoredRanks returnes the number of elements in the set
func (d *DisjointSet) GetStoredRanks() (ranks []uint64, err error) {
	r, err := d.storage.ReadPrefix("r")
	for _, rank := range r {
		ranks = append(ranks, binary.LittleEndian.Uint64(rank))
	}
	return
}

// GetStoredClusters returnes the number of elements in the set
func (d *DisjointSet) GetStoredClusters() (clusters map[string][]byte, err error) {
	clusters, err = d.storage.ReadPrefixWithKey("addr")
	return
}

// GetHashMap returnes the set hashmap
func (d *DisjointSet) GetHashMap() *sync.Map {
	return &d.hashMap
}

// GetParent returnes parent based on the passed tag
func (d *DisjointSet) GetParent(tag uint64) uint64 {
	return d.parent[tag]
}

// MakeSet creates a new set based adding the parameter passed as argument to the set
func (d *DisjointSet) MakeSet(x interface{}) {}

// PrepareMakeSet creates a new set based adding the parameter passed as argument to the set
func (d *DisjointSet) PrepareMakeSet(x interface{}, batch *sync.Map) {
	if _, ok := d.hashMap.Load(x); ok {
		return
	}

	d.hashMap.Store(x, d.size)
	d.parent[d.size] = d.size
	d.rank[d.size] = 0

	s := make([]byte, 8)
	binary.LittleEndian.PutUint64(s, d.size)
	batch.Store("addr"+x.(string), s)

	batch.Store(fmt.Sprintf("p%d", d.size), s)
	batch.Store(fmt.Sprintf("r%d", d.size), make([]byte, 8))

	d.size = d.size + 1

	ns := make([]byte, 8)
	binary.LittleEndian.PutUint64(ns, d.size)
	batch.Store("size", ns)
}

// Find returnes the value of the set required as argument to the function
func (d *DisjointSet) Find(x interface{}, batch *sync.Map) (uint64, error) {
	pos, ok := d.hashMap.Load(x)
	if !ok {
		return 0, errors.New("Element not found")
	}
	return d.FindInternal(d.parent, pos.(uint64), batch), nil
}

// FindInternal recursively search for the element of depth n in the set
func (d *DisjointSet) FindInternal(p []uint64, n uint64, batch *sync.Map) uint64 {
	if p[n] != n {
		parent := p[n]
		res := d.FindInternal(p, parent, batch)
		p[n] = res
		if batch != nil {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, res)
			batch.Store(fmt.Sprintf("p%d", n), b)
		}
		return p[n]
	}
	return n
}

// Union returnes the common set to the elements passed as arguments
func (d *DisjointSet) Union(x, y interface{}) (uint64, error) {
	return 0, nil
}

// PrepareUnion returnes the common set to the elements passed as arguments
func (d *DisjointSet) PrepareUnion(x, y interface{}, batch *sync.Map) (uint64, error) {
	var (
		xRoot,
		yRoot,
		xRank,
		yRank uint64
	)

	xRoot, err := d.Find(x, batch)
	if err != nil {
		logger.Error("Disjoint Set", err, logger.Params{})
		return 0, err
	}
	xRank = d.rank[xRoot]
	yRoot, err = d.Find(y, batch)
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

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, xRoot)
		batch.Store(fmt.Sprintf("p%d", yRoot), b)

		return xRoot, nil
	}
	d.parent[xRoot] = yRoot

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, yRoot)
	batch.Store(fmt.Sprintf("p%d", xRoot), b)

	if xRank == yRank {
		d.rank[yRoot]++

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, d.rank[yRoot])
		batch.Store(fmt.Sprintf("r%d", yRoot), b)
	}
	return yRoot, nil
}

// BulkUpdate updates cluster synced height
func (d *DisjointSet) BulkUpdate(batch *sync.Map) error {
	b := make(map[string][]byte)
	batch.Range(func(k, v interface{}) bool {
		b[k.(string)] = v.([]byte)
		return true
	})
	return d.storage.StoreBatch(b)
}

// UpdateHeight updates cluster synced height
func (d *DisjointSet) UpdateHeight(height int32) error {
	d.height = height

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(height))
	return d.storage.Store("height", b)
}

// UpdateSize updates cluster synced height
func (d *DisjointSet) UpdateSize(size uint64) error {
	d.size = size

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, size)
	return d.storage.Store("size", b)
}

// Finalize parses the entire set
func (d *DisjointSet) Finalize() {
	for i := uint64(0); i < d.size; i++ {
		d.FindInternal(d.parent, i, nil)
	}
}
