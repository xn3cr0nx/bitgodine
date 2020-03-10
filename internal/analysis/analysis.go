package analysis

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/behaviour"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/locktime"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/power"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/reuse"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/shadow"
	class "github.com/xn3cr0nx/bitgodine_server/internal/heuristics/type"
	"github.com/xn3cr0nx/bitgodine_server/internal/task"
)

// HeuristicChangeAnalysis analysis output map for heuristics change output
type HeuristicChangeAnalysis map[heuristics.Heuristic]uint32

// AnalyzeTx applies all the heuristics to the transaction returning a byte mask representing bool condition on vulnerabilites
func AnalyzeTx(c *echo.Context, txid string) (vuln heuristics.Mask, err error) {
	ca := (*c).Get("cache").(*cache.Cache)
	if res, ok := ca.Get("v_" + txid); ok {
		vuln = res.(heuristics.Mask)
		return
	}
	kv := (*c).Get("kv").(*badger.Badger)
	if v, e := kv.Read(txid); e == nil {
		vuln = heuristics.MaskFromBytes(v)
		return
	}
	db := (*c).Get("db").(storage.DB)
	tx, err := db.GetTx(txid)
	if err != nil {
		if err.Error() == "transaction not found" {
			err = echo.NewHTTPError(http.StatusNotFound, err)
		}
		return
	}

	heuristics.ApplyFullSet(db, tx, &vuln)

	if err = kv.Store(txid, vuln.Bytes()); err != nil {
		return
	}
	if !ca.Set("v_"+txid, vuln, 1) {
		(*c).Logger().Error(err)
	}
	return
}

// TxChange apply tx analysis and inferes transaction change output
func TxChange(c *echo.Context, txid string, heuristicsList heuristics.Mask, unfeasibleConditions ...func(tx models.Tx) bool) (vout HeuristicChangeAnalysis, err error) {
	// ca := (*c).Get("cache").(*cache.Cache)
	// if res, ok := ca.Get("v_" + txid); ok {
	// 	vuln = res.(byte)
	// 	return
	// }
	// kv := (*c).Get("kv").(*badger.Badger)
	// if v, e := kv.Read(txid); e == nil {
	// 	vuln = v[0]
	// 	return
	// }
	db := (*c).Get("db").(storage.DB)
	tx, err := db.GetTx(txid)
	if err != nil {
		if err.Error() == "transaction not found" {
			err = echo.NewHTTPError(http.StatusNotFound, err)
		}
		return
	}

	vout = make(map[heuristics.Heuristic]uint32, len(heuristicsList.ToList()))

	for _, condition := range unfeasibleConditions {
		if condition(tx) {
			err = errors.New("Not feasible transaction")
			return
		}
	}

	if len(tx.Vin) > 1 && len(tx.Vout) <= 1 {
		fmt.Println("Possible self transfer")
		vout[0] = 0
		return
	}

	for _, heuristic := range heuristicsList.ToList() {
		switch heuristic {
		case 0:
			c, e := locktime.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 1:
			c, e := peeling.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			vout[heuristic] = c
		case 2:
			c, e := power.ChangeOutput(&tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 3:
			c, e := optimal.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 4:
			c, e := class.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 5:
			c, e := reuse.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 6:
			c, e := shadow.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case 7:
			c, e := behaviour.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
			// case "Forward":
			// 	c, e := forward.ChangeOutput(db, &tx)
			// 	if e != nil {
			// 		break
			// 	}
			// case "Backward":
			// 	c, e := backward.ChangeOutput(db, &tx)
			// 	if e != nil {
			// 		break
			// 	}
			// }
		}
	}

	// if err = kv.Store(txid, []byte{vuln}); err != nil {
	// 	return
	// }
	// if !ca.Set("v_"+txid, vuln, 1) {
	// 	(*c).Logger().Error(err)
	// }
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

// Worker wrapper to partecipate in task pool
type Worker struct {
	db             storage.DB
	height         int32
	tx             models.Tx
	lock           *sync.RWMutex
	vuln           MaskGraph
	heuristicsList heuristics.Mask
}

// Work method to make Worker compatible with task pool worker interface
func (w *Worker) Work() {
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
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, heuristicsList heuristics.Mask, force bool, chart string) (vuln MaskGraph, err error) {
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

	tip, err := db.LastBlock()
	if err != nil {
		return
	}
	if to > tip.Height {
		to = tip.Height
	}
	interval := int32(10000)
	analyzed := restorePreviousAnalysis(kv, from, to, interval)
	fmt.Println("prev analyzed chunks", len(analyzed))
	ranges := updateRange(from, to, analyzed, force)
	fmt.Println("updated ranges", ranges)

	pool := task.New(runtime.NumCPU() * 3)
	lock := sync.RWMutex{}
	vuln = make(MaskGraph, to-from+1)
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
			vuln[block.Height] = make(map[string]heuristics.Mask, len(block.Transactions))
			lock.Unlock()
			for _, txID := range block.Transactions {
				tx, e := db.GetTx(txID)
				if e != nil {
					err = e
					return
				}

				pool.Do(&Worker{
					height:         block.Height,
					tx:             tx,
					db:             db,
					lock:           &lock,
					vuln:           vuln,
					heuristicsList: heuristicsList,
				})
			}
		}
	}
	pool.Shutdown()

	fmt.Println("storing ranges", ranges)
	if force {
		fmt.Println("updating ranges")
		newVuln := updateStoredRanges(kv, interval, analyzed, vuln)
		vuln = mergeGraphs(vuln, newVuln)
		for _, r := range ranges {
			if e := storeRange(kv, r, interval, vuln); e != nil {
				logger.Error("Analysis", e, logger.Params{})
			}
		}
	} else {
		for _, r := range ranges {
			if e := storeRange(kv, r, interval, vuln); e != nil {
				logger.Error("Analysis", e, logger.Params{})
			}
		}

		analyzed = append(analyzed, Chunk{Vulnerabilites: vuln})
		vuln = mergeChunks(analyzed...).Vulnerabilites
	}

	err = generateOutput(vuln, chart, heuristicsList, from, to)
	return
}

func offByOneAnalysis(c *echo.Context, from, to int32, heuristicsList heuristics.Mask, chart string) (err error) {
	db := (*c).Get("db").(storage.DB)
	if db == nil {
		err = errors.New("db not initialized")
		return
	}

	vuln := make(HeuristicGraph, from-to+1)
	for i := from; i <= to; i++ {
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

		vuln[block.Height] = make(map[string]HeuristicChangeAnalysis, len(block.Transactions))
		for _, txID := range block.Transactions {
			changeOutputs, e := TxChange(c, txID, heuristicsList, coinbaseCondition, offByOneBugCondition)
			if e != nil {
				if e.Error() == "Not feasible transaction" {
					continue
				}
				err = e
				return
			}
			vuln[block.Height][txID] = changeOutputs
		}
	}

	err = generateOutput(vuln, chart, heuristicsList, from, to)
	return
}
