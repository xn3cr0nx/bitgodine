package analysis

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_server/internal/plot"
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

// Worker wrapper to partecipate in task pool
type Worker struct {
	db             storage.DB
	height         int32
	tx             models.Tx
	lock           *sync.RWMutex
	vuln           Graph
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

// Graph alias for struct describing blockchain graph based on vulnerabilities mask
type Graph map[int32]map[string]byte

// Chunk struct with info on previous analyzed blocks slice
type Chunk struct {
	Range          `json:"range,omitempty"`
	Vulnerabilites Graph `json:"vulnerabilities,omitempty"`
}

func restorePreviousAnalysis(kv *badger.Badger, from, to, interval int32) (intervals []Chunk) {
	if to-from >= interval {
		upper := upperBoundary(from, interval)
		lower := lowerBoundary(to, interval)
		fmt.Println("restoring in range", upper, lower, interval)
		for i := upper; i < lower; i += interval {
			r, err := kv.Read(fmt.Sprintf("int%d-%d", i, i+interval))
			fmt.Println("read range", i, i+interval, err)
			if err != nil {
				break
			}
			var analyzed Chunk
			err = encoding.Unmarshal(r, &analyzed)
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				break
			}
			intervals = append(intervals, analyzed)
		}
	} else {
		lower := lowerBoundary(from, interval)
		upper := upperBoundary(to, interval)
		r, err := kv.Read(fmt.Sprintf("int%d-%d", lower, upper))
		if err != nil {
			return
		}
		var analyzed Chunk
		err = encoding.Unmarshal(r, &analyzed)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
		}
		analyzed.Vulnerabilites = subGraph(analyzed.Vulnerabilites, from, to)
		intervals = []Chunk{analyzed}
	}
	return
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, heuristicsList []string, export bool) (vuln Graph, err error) {
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
	ranges := updateRange(from, to, analyzed)
	fmt.Println("updated ranges", ranges)

	pool := task.New(runtime.NumCPU() * 3)
	lock := sync.RWMutex{}
	vuln = make(map[int32]map[string]byte, to-from+1)
	for _, r := range ranges {
		if r.From == r.To {
			continue
		}
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
	for _, r := range ranges {
		if e := storeRange(kv, r, interval, vuln); e != nil {
			logger.Error("Analysis", e, logger.Params{})
		}
	}

	analyzed = append(analyzed, Chunk{Vulnerabilites: vuln})
	vuln = mergeChunks(analyzed...).Vulnerabilites

	if export {
		data := heuristics.ExtractPercentages(vuln, heuristicsList, from, to)
		err = PlotHeuristicsTimeline(data, from, heuristicsList)
	} else {
		data := heuristics.ExtractGlobalPercentages(vuln, heuristicsList, from, to)
		err = GlobalPercentages(data, heuristicsList)
	}

	return
}

// GlobalPercentages prints a table with percentages of heuristics success rate based on passed analysis
func GlobalPercentages(data []float64, heuristicsList []string) (err error) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")
	for h, perc := range data {
		table.Append([]string{heuristicsList[h], fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
	return
}

// PlotHeuristicsTimeline plots timeseries of heuristics percentage effectiveness for each block representing time series
func PlotHeuristicsTimeline(data map[int32][]float64, min int32, heuristicsList []string) (err error) {
	coordinates := make(map[string]plot.Coordinates)
	x := make([]float64, len(data))
	for height := range data {
		x[height-min] = float64(height)
	}

	for h, heuristic := range heuristicsList {
		y := make([]float64, len(data))
		for height, vulnerabilites := range data {
			y[int(height-min)] = vulnerabilites[h] + float64(h)
		}
		coordinates[heuristic] = plot.Coordinates{X: x, Y: y}
	}

	title := "Heuristics timeline"
	if len(heuristicsList) == 1 {
		title = heuristicsList[0] + " timeline"
	}
	err = plot.MultipleLineChart(title, "blocks", "heuristics effectiveness", coordinates)

	return
}
