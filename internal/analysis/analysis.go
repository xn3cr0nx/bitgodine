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
		heuristics.ApplyConditionSet(db, tx, &addr)
		vuln = addr
	} else {
		vuln = make(heuristics.Map)
		addr := vuln.(heuristics.Map)
		heuristics.ApplyChangeSet(db, tx, heuristicsList, &addr)
		heuristics.ApplyChangeConditionSet(db, tx, &addr)
		vuln = addr
	}

	return
}

// ExtractLikelihoodOutput function to return most probable output between applied heuristics results
func ExtractLikelihoodOutput(analyzed heuristics.Map) (vout uint32, err error) {
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

func recursive(prefix heuristics.Mask, h []heuristics.Heuristic, result []heuristics.Mask) []heuristics.Mask {
	for i, e := range h {
		mask := heuristics.MaskFromPower(e)
		if prefix[0] > 0 {
			mask = heuristics.MergeMasks(mask, prefix)
		}
		result = append(result, mask)
		recursive(mask, h[i+1:], result)
	}
	return result
}

func getCombinations(h []heuristics.Heuristic) (result []heuristics.Mask) {
	result = recursive(heuristics.Mask{}, h, result)
	return
}

// MajorityLikelihood extract majority output sets with likelihood percentages
func MajorityLikelihood(v heuristics.Map) (likelihood map[uint32]map[heuristics.Mask]float64) {
	majority := make(heuristics.Map, len(v))
	for key, value := range v {
		majority[key] = value
	}

	for _, n := range []heuristics.Heuristic{5, 6, 8, 9, 10, 11, 17, 18, 19, 20} {
		delete(majority, n)
	}

	clusters := make(map[uint32][]heuristics.Heuristic)
	for heuristic, change := range majority {
		clusters[change] = append(clusters[change], heuristic)
	}
	if len(clusters) == 0 {
		return
	}

	likelihood = make(map[uint32]map[heuristics.Mask]float64, len(clusters))
	for output, list := range clusters {
		combinations := getCombinations(list)
		max := float64(-1)
		index := 0
		for i, combination := range combinations {
			if perc, ok := heuristics.MajorityLikelihood[combination[0]]; ok {
				if max < perc {
					index = i
					max = perc
				}
			}
		}
		if max > 0 {
			likelihood[output] = make(map[heuristics.Mask]float64, 1)
			likelihood[output][combinations[index]] = max
		}
	}

	return
}

// MajorityVotingOutput return map with probability of change output for each output
func MajorityVotingOutput(analyzed heuristics.Map) (likelihood map[uint32]map[heuristics.Mask]float64, err error) {
	if len(analyzed) == 0 {
		return nil, errors.New("Not feasible transaction")
	}
	likelihood = make(map[uint32]map[heuristics.Mask]float64, 1)
	if out, ok := analyzed[heuristics.Index("Address Reuse")]; ok {
		likelihood[out] = make(map[heuristics.Mask]float64, 1)
		likelihood[out][heuristics.MaskFromPower(heuristics.Index("Address Reuse"))] = 100
		return
	}
	if out, ok := analyzed[heuristics.Index("Shadow")]; ok {
		if off, ok := analyzed[heuristics.Index("OffByOne")]; ok && off == 0 {
			if out == 0 {
				likelihood[1] = make(map[heuristics.Mask]float64, 1)
				likelihood[1][heuristics.MaskFromPower(heuristics.Index("Shadow"))] = 100
			} else {
				likelihood[0] = make(map[heuristics.Mask]float64, 1)
				likelihood[0][heuristics.MaskFromPower(heuristics.Index("Shadow"))] = 100
			}
		}
		return
	}

	likelihood = MajorityLikelihood(analyzed)
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
	vuln MaskGraph
}

// Work method to make ApplicabilityWorker compatible with task pool worker interface
func (w *ApplicabilityWorker) Work() (err error) {
	var v heuristics.Mask
	heuristics.ApplySet(w.db, w.tx, w.heuristicsList, &v)
	heuristics.ApplyConditionSet(w.db, w.tx, &v)

	w.lock.Lock()
	w.vuln[w.height][w.tx.TxID] = v
	w.lock.Unlock()
	return
}

// ReliabilityWorker wrapper to partecipate in task pool
type ReliabilityWorker struct {
	Worker
	vuln OutputGraph
}

// Work method to make ReliabilityWorker compatible with task pool worker interface
func (w *ReliabilityWorker) Work() (err error) {
	v := make(heuristics.Map, len(w.heuristicsList.ToList()))
	heuristics.ApplyChangeSet(w.db, w.tx, w.heuristicsList, &v)
	heuristics.ApplyChangeConditionSet(w.db, w.tx, &v)

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

	// // define tx analysis conditions based on analysis type and criteria
	// conditions := newConditionsSet()
	// conditions.fillConditionsSet(criteria)

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
				vuln.(OutputGraph)[block.Height] = make(map[string]heuristics.Map, len(block.Transactions))
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
					})
				} else {
					pool.Do(&ReliabilityWorker{
						w,
						vuln.(OutputGraph),
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

	err = generateOutput(vuln, chart, criteria, heuristicsList, from, to)
	return
}
