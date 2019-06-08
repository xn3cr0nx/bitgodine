package persistent

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/dgo"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// DisjointSet implements disjoint set logic in a persistent way using key value storage
type DisjointSet struct {
	SetSize uint32
	Parent  []uint32
	Rank    []uint32
	HashMap map[interface{}]uint32
	storage *dgo.Dgraph
}

// NewDisjointSet returnes a reference to a new istance of persisted disjoint set
func NewDisjointSet(db *dgo.Dgraph) DisjointSet {
	if _, err := dgraph.GetClusterUID(); err != nil {
		if err.Error() == "Cluster not found" {
			if err := dgraph.NewClusters(); err != nil {
				logger.Error("Persistent Disjoint Set", err, logger.Params{})
				os.Exit(-1)
			}
		}
	}

	// const CAPACITY int = 1000000
	return DisjointSet{
		SetSize: 0,
		Parent:  []uint32{},
		Rank:    []uint32{},
		HashMap: map[interface{}]uint32{},
		storage: db,
	}
}

// RestorePersistentSet initialize the disjoint set with the persisted state
func RestorePersistentSet(d *DisjointSet) error {
	clusters, err := dgraph.GetClusters()
	if err != nil {
		return err
	}
	fmt.Println("found cluster")

	fmt.Println("size", clusters.Size)
	d.SetSize = clusters.Size
	fmt.Println("parents", clusters.Parents)
	for _, parent := range clusters.Parents {
		d.Parent[parent.Pos] = parent.Parent
	}
	fmt.Println("ranks", clusters.Ranks)
	for _, rank := range clusters.Ranks {
		d.Rank[rank.Pos] = rank.Rank
	}
	fmt.Println("sets", clusters.Set)
	for _, cluster := range clusters.Set {
		for _, address := range cluster.Addresses {
			d.HashMap[address.Address] = cluster.Cluster
		}
	}
	return nil
}

// Size returnes the number of elements in the set
func (d *DisjointSet) Size() uint32 {
	return d.SetSize
}

// GetHashMap returnes the set hashmap
func (d *DisjointSet) GetHashMap() map[interface{}]uint32 {
	return d.HashMap
}

// GetParent returnes parent based on the passed tag
func (d *DisjointSet) GetParent(tag uint32) uint32 {
	return d.Parent[tag]
}

// MakeSet creates a new set based adding the parameter passed as argument to the set
func (d *DisjointSet) MakeSet(x interface{}) {
	if _, ok := d.HashMap[x]; ok {
		return
	}

	d.HashMap[x] = d.SetSize
	// persistence
	dgraph.NewSet(string(x.(visitor.Utxo)), d.SetSize)
	//

	d.Parent = append(d.Parent, d.SetSize)
	// persistence
	dgraph.AddParent(uint32(len(d.Parent)), d.SetSize)
	//

	d.Rank = append(d.Rank, 0)
	// persistence
	dgraph.AddRank(uint32(len(d.Rank)), 0)
	//

	d.SetSize = d.SetSize + 1
	// persistence
	dgraph.UpdateSize(d.SetSize + 1)
	//
}

// Find returnes the value of the set required as argument to the function
func (d *DisjointSet) Find(x interface{}) (uint32, error) {

	pos, ok := d.HashMap[x]
	if !ok {
		return 0, errors.New("Element not found")
	}
	return d.FindInternal(d.Parent, pos), nil
}

// FindInternal recursively search for the element of depth n in the set
func (d *DisjointSet) FindInternal(p []uint32, n uint32) uint32 {
	if p[n] != n {
		parent := p[n]
		p[n] = d.FindInternal(p, parent)
		return p[n]
	}
	return n
}

// Union returnes the common set to the elements passed as arguments
func (d *DisjointSet) Union(x, y interface{}) (uint32, error) {
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
		// persistent
		dgraph.UpdateParent(yRoot, xRoot)
		//
		return xRoot, nil
	}
	d.Parent[xRoot] = yRoot
	if xRank == yRank {
		d.Rank[yRoot]++
		// persistent
		dgraph.UpdateRank(yRoot, d.Rank[yRoot])
		//
	}
	return yRoot, nil
}

// Finalize parses the entire set
func (d *DisjointSet) Finalize() {
	for i := 0; uint32(i) < d.SetSize; i++ {
		d.FindInternal(d.Parent, uint32(i))
	}
}
