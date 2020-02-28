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

// AnalyzeTx applies all the heuristics to the transaction returning a byte mask representing bool condition on vulnerabilites
func AnalyzeTx(c *echo.Context, txid string) (vuln byte, err error) {
	ca := (*c).Get("cache").(*cache.Cache)
	if res, ok := ca.Get("v_" + txid); ok {
		vuln = res.(byte)
		return
	}
	kv := (*c).Get("kv").(*badger.Badger)
	if v, e := kv.Read(txid); e == nil {
		vuln = v[0]
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

	if err = kv.Store(txid, []byte{vuln}); err != nil {
		return
	}
	if !ca.Set("v_"+txid, vuln, 1) {
		(*c).Logger().Error(err)
	}
	return
}

// TxChange apply tx analysis and inferes transaction change output
func TxChange(c *echo.Context, txid string, heuristicsList []string) (vout map[string]uint32, err error) {
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

	if (len(tx.Vin) == 1 && tx.Vin[0].IsCoinbase) || len(tx.Vout) <= 1 {
		err = errors.New("Not feasible transaction")
		return
	}

	vout = make(map[string]uint32, len(heuristicsList))
	for _, heuristic := range heuristicsList {
		switch heuristic {
		case "Locktime":
			c, e := locktime.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Peeling Chain":
			c, e := peeling.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			vout[heuristic] = c
		case "Power of Ten":
			c, e := power.ChangeOutput(&tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Optimal Change":
			c, e := optimal.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Address Type":
			c, e := class.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Address Reuse":
			c, e := reuse.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Shadow":
			c, e := shadow.ChangeOutput(db, &tx)
			if e != nil {
				break
			}
			if len(c) == 1 {
				vout[heuristic] = c[0]
			}
		case "Client Behaviour":
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

// Worker wrapper to partecipate in task pool
type Worker struct {
	db             storage.DB
	height         int32
	tx             models.Tx
	lock           *sync.RWMutex
	vuln           MaskedGraph
	heuristicsList []string
}

// Work method to make Worker compatible with task pool worker interface
func (w *Worker) Work() {
	if len(w.tx.Vout) <= 1 {
		// TODO: we are not considering coinbase 1 output txs in heuristics analysis
		w.lock.Lock()
		w.vuln[w.height][w.tx.TxID] = 0
		w.lock.Unlock()
		return
	}

	var v byte
	heuristics.ApplySet(w.db, w.tx, w.heuristicsList, &v)

	w.lock.Lock()
	w.vuln[w.height][w.tx.TxID] = v
	w.lock.Unlock()
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, heuristicsList []string, force bool, chart string) (vuln MaskedGraph, err error) {
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
	vuln = make(map[int32]map[string]byte, to-from+1)
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
			vuln[block.Height] = make(map[string]byte, len(block.Transactions))
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

func offByOneAnalysis(c *echo.Context, from, to int32, heuristicsList []string, chart string) (err error) {
	db := (*c).Get("db").(storage.DB)
	if db == nil {
		err = errors.New("db not initialized")
		return
	}
	// kv := (*c).Get("kv").(*badger.Badger)
	// if kv == nil {
	// 	err = errors.New("kv storage not initialized")
	// 	return
	// }

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

		vuln[block.Height] = make(map[string]map[string]uint32, len(block.Transactions))
		for _, txID := range block.Transactions {
			changeOutputs, e := TxChange(c, txID, heuristicsList)
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
