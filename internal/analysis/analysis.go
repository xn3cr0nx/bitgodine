// Package analysis applicability vs reliability: applicability describes in a deterministic
// way if an heuristic is applicable to a transaction based on its conditions while
// reliability is the degree of reliability the heuristic provides whether it is applicable
// to the transaction.
package analysis

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// HeuristicChangeAnalysis analysis output map for heuristics change output
type HeuristicChangeAnalysis map[heuristics.Heuristic]uint32

// AnalyzeTx applies all the heuristics to the transaction returning a byte mask representing bool condition on vulnerabilites
func AnalyzeTx(c *echo.Context, txid string, heuristicsList heuristics.Mask, analysisType string) (vuln interface{}, err error) {
	db := (*c).Get("db").(storage.DB)
	tx, err := db.GetTx(txid)
	if err != nil {
		if err.Error() == "transaction not found" {
			err = echo.NewHTTPError(http.StatusNotFound, err)
		}
		return
	}

	if analysisType == "applicability" {
		vuln = heuristics.FromListToMask(nil)
		addr := vuln.(heuristics.Mask)
		heuristics.ApplySet(db, tx, heuristicsList, &addr)
		vuln = addr
	} else {
		vuln = make(map[heuristics.Heuristic]uint32)
		addr := vuln.(map[heuristics.Heuristic]uint32)
		heuristics.ApplyChangeSet(db, tx, heuristicsList, &addr)
		vuln = addr
	}

	return
}

// ExtractLikelihoodOutput function to return most probable output between applied heuristics results
func ExtractLikelihoodOutput(analyzed HeuristicChangeAnalysis) (vout uint32, err error) {
	fmt.Println("from heuristics analysis", analyzed)
	if len(analyzed) == 0 {
		return 0, errors.New("No effective analysis")
	}
	if len(analyzed) == 1 {
		for _, v := range analyzed {
			return v, nil
		}
	}
	if out, ok := analyzed[heuristics.Index("Optimal Change")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Address Reuse")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Shadow")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Peeling Chain")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Address Type")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Power of Ten")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Client Behaviour")]; ok {
		return out, nil
	}
	if out, ok := analyzed[heuristics.Index("Locktime")]; ok {
		return out, nil
	}
	return
}

// Worker basic worker to partecipate in task pool
type Worker struct {
	db             storage.DB
	height         int32
	tx             models.Tx
	lock           *sync.RWMutex
	heuristicsList heuristics.Mask
}

// ApplicabilityWorker wrapper to partecipate in task pool
type ApplicabilityWorker struct {
	Worker
	vuln       MaskGraph
	conditions ConditionsSet
}

// Work method to make ApplicabilityWorker compatible with task pool worker interface
func (w *ApplicabilityWorker) Work() (err error) {
	// excluding coinbase transactions in analysis
	if coinbaseCondition(w.tx) {
		// w.lock.Lock()
		// w.vuln[w.height][w.tx.TxID] = heuristics.MaskFromPower(0)
		// w.lock.Unlock()
		return
	}

	var v heuristics.Mask
	heuristics.ApplySet(w.db, w.tx, w.heuristicsList, &v)

	w.lock.Lock()
	w.vuln[w.height][w.tx.TxID] = v
	w.lock.Unlock()
	return
}

// ReliabilityWorker wrapper to partecipate in task pool
type ReliabilityWorker struct {
	Worker
	vuln       OutputGraph
	conditions ConditionsSet
}

// Work method to make ReliabilityWorker compatible with task pool worker interface
func (w *ReliabilityWorker) Work() (err error) {
	for _, condition := range w.conditions {
		if condition(w.tx) {
			return
		}
	}

	// if len(tx.Vout) == 1 {
	// 	fmt.Println("Possible self transfer")
	// 	vout[0] = 0
	// 	return
	// }

	v := make(map[heuristics.Heuristic]uint32, len(w.heuristicsList.ToList()))
	heuristics.ApplyChangeSet(w.db, w.tx, w.heuristicsList, &v)

	w.lock.Lock()
	w.vuln[w.height][w.tx.TxID] = v
	w.lock.Unlock()
	return
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, heuristicsList heuristics.Mask, analysisType, criteria, chart string, force bool) (err error) {
	db := (*c).Get("db").(storage.DB)
	if db == nil {
		err = errors.New("db not initialized")
		return
	}
	kv := (*c).Get("kv").(*badger.Badger)
	if kv == nil {
		err = errors.New("kv storage not initialized")
		return
	}

	gob.Register(MaskGraph{})
	gob.Register(OutputGraph{})

	tip, err := db.LastBlock()
	if err != nil {
		return
	}
	if to > tip.Height {
		to = tip.Height
	}
	interval := int32(10000)
	analyzed := restorePreviousAnalysis(kv, from, to, interval, analysisType)
	fmt.Println("prev analyzed chunks", len(analyzed))
	ranges := updateRange(from, to, analyzed, force)
	fmt.Println("updated ranges", ranges)

	// define tx analysis conditions based on analysis type and criteria
	conditions := newConditionsSet()
	conditions.fillConditionsSet(criteria)

	var vuln Graph
	if analysisType == "applicability" {
		vuln = make(MaskGraph, to-from+1)
	} else {
		vuln = make(OutputGraph, to-from+1)
	}

	pool := task.New(runtime.NumCPU() * 3)
	lock := sync.RWMutex{}

	for _, r := range ranges {
		for i := r.From; i <= r.To; i++ {
			block, e := db.GetBlockFromHeight(i)
			if e != nil {
				if e.Error() == "Key not found" {
					break
				}
				err = e
				return
			}
			if block.Height%1000 == 0 {
				logger.Info("Analysis", "Analyzing block", logger.Params{"height": block.Height, "hash": block.ID})
			}

			lock.Lock()
			if analysisType == "applicability" {
				vuln.(MaskGraph)[block.Height] = make(map[string]heuristics.Mask, len(block.Transactions))
			} else {
				vuln.(OutputGraph)[block.Height] = make(map[string]HeuristicChangeAnalysis, len(block.Transactions))
			}
			lock.Unlock()

			for _, txID := range block.Transactions {
				tx, e := db.GetTx(txID)
				if e != nil {
					err = e
					return
				}

				w := Worker{
					height:         block.Height,
					tx:             tx,
					db:             db,
					lock:           &lock,
					heuristicsList: heuristicsList,
				}
				if analysisType == "applicability" {
					pool.Do(&ApplicabilityWorker{
						w,
						vuln.(MaskGraph),
						conditions,
					})
				} else {
					pool.Do(&ReliabilityWorker{
						w,
						vuln.(OutputGraph),
						conditions,
					})
				}

			}
		}
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	fmt.Println("storing ranges", ranges)
	if force {
		fmt.Println("updating ranges")
		newVuln := vuln.updateStoredRanges(kv, interval, analyzed)
		vuln = vuln.mergeGraphs(newVuln)
		for _, r := range ranges {
			if e := storeRange(kv, r, interval, vuln, analysisType); e != nil {
				logger.Error("Analysis", e, logger.Params{})
			}
		}
	} else {
		for _, r := range ranges {
			if e := storeRange(kv, r, interval, vuln, analysisType); e != nil {
				logger.Error("Analysis", e, logger.Params{})
			}
		}
		vuln = vuln.mergeChunks(analyzed...).Vulnerabilites
	}

	err = generateOutput(vuln, chart, heuristicsList, from, to)
	return
}
