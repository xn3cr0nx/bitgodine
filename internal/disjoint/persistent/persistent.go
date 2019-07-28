package persistent

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/dgraph-io/dgo"
	"github.com/gosuri/uiprogress"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/memory"
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

	logger.Info("Persistent Disjoint Set", "Restoring the clusters", logger.Params{"size": clusters.Size})

	d.SetSize = clusters.Size
	logger.Debug("Persistent", "Restoring parents", logger.Params{"size": len(clusters.Parents)})
	for _, parent := range clusters.Parents {
		d.Parent = append(d.Parent, parent.Parent)
	}
	logger.Debug("Persistent", "Restoring ranks", logger.Params{"size": len(clusters.Ranks)})
	for _, rank := range clusters.Ranks {
		d.Rank = append(d.Rank, rank.Rank)
	}
	for _, cluster := range clusters.Set {
		for _, address := range cluster.Addresses {
			d.HashMap[visitor.Utxo(address.Address)] = cluster.Cluster
		}
	}
	return nil
}

// RecoverPersistentSet restores consisent clusters set based on blocks synced in dgraph
func (d *DisjointSet) RecoverPersistentSet(m *memory.DisjointSet) error {
	logger.Info("Persistent", "Recovering clusters", logger.Params{"size": strconv.Itoa(int(d.SetSize))})

	fmt.Println("")
	uiprogress.Start()
	bar := uiprogress.AddBar(int(d.SetSize)).AppendCompleted().PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Building new set")
	})

	var height int32
	for {
		block, err := dgraph.GetBlockFromHeight(height)
		if err != nil {
			if err.Error() == "Block not found" {
				break
			}
			return err
		}

		for _, tx := range block.Transactions {
			// TODO: need to check coinjoin too
			logger.Debug("Persistent", fmt.Sprintf("Restoring tx %v", tx.Hash), logger.Params{})
			if len(tx.Inputs) > 1 {
				lastAddress, err := tx.Inputs[0].GetAddress()
				if err != nil {
					return err
				}
				m.MakeSet(lastAddress)
				for _, input := range tx.Inputs {
					addr, err := input.GetAddress()
					if err != nil {
						return err
					}
					m.MakeSet(addr)
					m.Union(lastAddress, addr)
					lastAddress = addr

					bar.Incr()
				}
			}
		}
		height++
	}

	// replace broken persistent set
	if err := dgraph.UpdateSize(m.SetSize); err != nil {
		return err
	}

	// uiprogress.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	setBar := uiprogress.AddBar(len(m.HashMap)).AppendCompleted().PrependElapsed()
	setBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Updating set: %v/%v", b.Current(), len(m.HashMap))
	})
	go func() {
		defer wg.Done()
		for addr := range m.HashMap {
			if _, ok := d.HashMap[addr]; !ok {
				if err := dgraph.NewSet(addr.(string), d.SetSize); err != nil {
					logger.Error("Persistent", err, logger.Params{})
					os.Exit(-1)
				}
			}
			setBar.Incr()
		}
	}()

	wg.Add(1)
	rankBar := uiprogress.AddBar(len(m.Rank)).AppendCompleted().PrependElapsed()
	rankBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Updating ranks: %v/%v", b.Current(), len(m.Rank))
	})
	go func() {
		defer wg.Done()
		for i, rank := range m.Rank {
			if len(d.Rank) > i {
				if d.Rank[i] != m.Rank[i] {
					if err := dgraph.UpdateRank(uint32(i), rank); err != nil {
						logger.Error("Persistent", err, logger.Params{})
						os.Exit(-1)
					}
				}
			} else {
				if err := dgraph.AddRank(uint32(i), rank); err != nil {
					logger.Error("Persistent", err, logger.Params{})
					os.Exit(-1)
				}
			}
			rankBar.Incr()
		}
	}()

	wg.Add(1)
	parentBar := uiprogress.AddBar(len(m.Parent)).AppendCompleted().PrependElapsed()
	parentBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Updating parents: %v/%v", b.Current(), len(m.Parent))
	})
	go func() {
		defer wg.Done()
		for i, rank := range m.Rank {
			if len(d.Rank) > i {
				if d.Rank[i] != m.Rank[i] {
					if err := dgraph.UpdateParent(uint32(i), rank); err != nil {
						logger.Error("Persistent", err, logger.Params{})
						os.Exit(-1)
					}
				}
			} else {
				if err := dgraph.AddParent(uint32(i), rank); err != nil {
					logger.Error("Persistent", err, logger.Params{})
					os.Exit(-1)
				}
			}
			parentBar.Incr()
		}
	}()

	wg.Wait()
	uiprogress.Stop()

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
	if err := dgraph.NewSet(string(x.(visitor.Utxo)), d.SetSize); err != nil {
		logger.Error("Persistent Disjoint Set", err, logger.Params{})
		os.Exit(-1)
	}
	//

	d.Parent = append(d.Parent, d.SetSize)
	// persistence
	logger.Debug("Dgraph Cluster", "Parent length", logger.Params{"parent_length": len(d.Parent), "rank_length": len(d.Rank), "set_size": d.SetSize})
	if err := dgraph.AddParent(uint32(len(d.Parent)-1), d.SetSize); err != nil {
		logger.Error("Persistent Disjoint Set", err, logger.Params{})
		os.Exit(-1)
	}
	//

	d.Rank = append(d.Rank, 0)
	// persistence
	if err := dgraph.AddRank(uint32(len(d.Rank)-1), 0); err != nil {
		logger.Error("Persistent Disjoint Set", err, logger.Params{})
		os.Exit(-1)
	}
	//

	d.SetSize = d.SetSize + 1
	// persistence
	if err := dgraph.UpdateSize(d.SetSize); err != nil {
		logger.Error("Persistent Disjoint Set", err, logger.Params{})
		os.Exit(-1)
	}
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
		if err := dgraph.UpdateParent(yRoot, xRoot); err != nil {
			logger.Error("Persistent Disjoint Set", err, logger.Params{})
			os.Exit(-1)
		}
		//
		return xRoot, nil
	}
	d.Parent[xRoot] = yRoot
	// persistent
	if err := dgraph.UpdateParent(xRoot, yRoot); err != nil {
		logger.Error("Persistent Disjoint Set", err, logger.Params{})
		os.Exit(-1)
	}
	//
	if xRank == yRank {
		d.Rank[yRoot]++
		// persistent
		if err := dgraph.UpdateRank(yRoot, d.Rank[yRoot]); err != nil {
			logger.Error("Persistent Disjoint Set", err, logger.Params{})
			os.Exit(-1)
		}
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
