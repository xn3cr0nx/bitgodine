package bitcoin

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
	"github.com/xn3cr0nx/bitgodine/pkg/task"
	"gorm.io/gorm"
)

const zeroHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Clusterize start fetching and generation of address clusters
func (c *Clusterizer) Clusterize() (err error) {
	var height int32
	if h := c.clusters.GetHeight(); h > 0 {
		height = h
	}
	logger.Info("Clusterizer", fmt.Sprintf("Starting clusterizer from block height %d", height), logger.Params{})

	for {
		syncedHeight, e := c.db.GetLastBlockHeight()
		if e != nil {
			return e
		}
		if syncedHeight <= height {
			logger.Info("Clusterizer", "Synced blocks and clusterizer to the same height", logger.Params{"height": syncedHeight})
			break
		}

		for ; height < syncedHeight; height++ {
			b, e := c.db.GetBlockFromHeight(height)
			if err != nil {
				return e
			}

			logger.Info("Clusterizer", fmt.Sprintf("Clusterizing block %d", b.Height), logger.Params{"height": b.Height, "hash": b.ID})

			for _, txID := range b.Transactions {
				tx, e := c.db.GetTx(txID)
				if e != nil {
					return e
				}

				logger.Debug("Clusterizer", "Clusterizing tx", logger.Params{"size": len(tx.Vout), "hash": tx.TxID})

				txItem := mapset.NewSet()

				pool := task.New(runtime.NumCPU() * 3)
				for _, in := range tx.Vin {
					pool.Do(&InputParser{c, in, &txItem})
				}
				if err = pool.Shutdown(); err != nil {
					return
				}

				logger.Debug("Clusterizer", "Updating cluster", logger.Params{"size": txItem.Cardinality()})
				err = c.UpdateCluster(&txItem)
			}

			logger.Warn("clusterizer", "Updating height", logger.Params{"height": b.Height})
			if err = c.clusters.UpdateHeight(b.Height); err != nil {
				logger.Error("Clusterizer", err, logger.Params{})
				return
			}
		}

	}

	return
}

// InputParser worker wrapper for parsing inputs in sync pool
type InputParser struct {
	c      *Clusterizer
	in     models.Input
	txItem *mapset.Set
}

// Work interface to execute input parser worker operations
func (w *InputParser) Work() (err error) {
	if w.in.TxID == zeroHash {
		return
	}
	spentTx, err := w.c.db.GetTx(w.in.TxID)
	if err != nil {
		return
	}

	utxoAddr := spentTx.Vout[w.in.Vout].ScriptpubkeyAddress
	if utxoAddr != "" {
		(*w.txItem).Add(utxoAddr)
	}
	return
}

// UpdateCluster implements first heuristic (all input are from the same user) and clusterize the input in the disjoint set
func (c *Clusterizer) UpdateCluster(txItem *mapset.Set) (err error) {
	// skip transactions with just one input
	// if (*txItem).Size() > 1 && !tx.IsCoinjoin() {
	if (*txItem).Cardinality() > 1 {
		// TODO: this is disjointset implementation dependent, should find a higher level implementation
		batch := sync.Map{}
		s := make([]byte, 8)
		binary.LittleEndian.PutUint64(s, c.clusters.GetSize())
		batch.Store("size", s)

		txInputs := (*txItem).ToSlice()
		lastAddress := txInputs[0]

		logger.Debug("Clusterizer", "Enhancing disjoint set", logger.Params{"last_address": lastAddress, "size": c.clusters.GetSize()})
		c.clusters.PrepareMakeSet(lastAddress, &batch)

		for _, address := range txInputs {
			c.clusters.PrepareMakeSet(address, &batch)
			if _, err = c.clusters.PrepareUnion(lastAddress, address, &batch); err != nil {
				return
			}
			lastAddress = address
		}

		if err = c.clusters.BulkUpdate(&batch); err != nil {
			return
		}
	}

	return
}

// Done worker group struct to concurrently execute final export operations
type Done struct {
	c       *Clusterizer
	tag     interface{}
	address interface{}
}

// Cluster cluster struct with validation
type Cluster struct {
	gorm.Model
	Address string `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"index"`
	Cluster uint64 `json:"cluster" validate:"" gorm:"index"`
}

// Work executes the sql insert
func (w *Done) Work() (err error) {
	// err = w.c.pg.DB.Save(&Cluster{Cluster: int(w.tag.(uint64)), Address: w.address.(string)}).Error
	err = w.c.pg.DB.FirstOrCreate(&Cluster{Cluster: w.tag.(uint64), Address: w.address.(string)}, &Cluster{Cluster: w.tag.(uint64), Address: w.address.(string)}).Error
	return
}

// Done finalizes the operations of the clusterizer exporting its content to a csv file
func (c *Clusterizer) Done() (size uint64, err error) {
	c.clusters.Finalize()
	logger.Info("Clusterizer", "Exporting clusters to CSV", logger.Params{"size": c.clusters.GetSize()})
	if viper.GetBool("sync.csv") {
		file, err := os.Create(fmt.Sprintf("%s/clusters.csv", viper.GetString("sync.output")))
		if err != nil {
			return 0, err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		c.clusters.GetHashMap().Range(func(address, tag interface{}) bool {
			writer.Write([]string{address.(string), strconv.Itoa(int(c.clusters.GetParent(tag.(uint64))))})
			return true
		})
	} else {
		// FIXME: I shouldn't start from the first each time
		pool := task.New(runtime.NumCPU() * 3)
		c.clusters.GetHashMap().Range(func(address, tag interface{}) bool {
			pool.Do(&Done{c, tag, address})
			return true
		})
		if err = pool.Shutdown(); err != nil {
			return
		}
	}

	logger.Info("Clusterizer", "Exported clusters", logger.Params{"size": c.clusters.GetSize()})
	size = c.clusters.GetSize()
	return
}
