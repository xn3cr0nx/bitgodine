package analysis

import (
	"encoding/json"
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

	heuristics.ApplySet(db, &tx, &vuln)

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
	c      *echo.Context
	height int32
	tx     *models.Tx
	lock   *sync.RWMutex
	store  map[string][]byte
	vuln   map[int32][]byte
}

// Work method to make Worker compatible with task pool worker interface
func (w *Worker) Work() {
	if len(w.tx.Vout) <= 1 {
		// TODO: we are not considering coinbase 1 output txs in heuristics analysis
		w.lock.Lock()
		defer w.lock.Unlock()
		w.store[w.tx.TxID] = []byte{0}
		w.vuln[w.height] = append(w.vuln[w.height], 0)
		return
	}
	v, err := AnalyzeTx(w.c, w.tx.TxID)
	if err != nil {
		return
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	w.store[w.tx.TxID] = []byte{v}
	w.vuln[w.height] = append(w.vuln[w.height], v)
}

func upperBoundary(n, interval int32) (r int32) {
	diff := n % interval
	if diff == 0 {
		diff = interval
	}
	r = n + (interval - diff)
	return
}

func lowerBoundary(n, interval int32) (r int32) {
	r = n - (n % interval)
	return
}

// Analyzed struct with info on previous analyzed blocks slice
type Analyzed struct {
	Range          `json:"range,omitempty"`
	Vulnerabilites map[int32][]byte `json:"vulnerabilities,omitempty"`
}

// Range wrapper for blocks interval boundaries
type Range struct {
	From int32 `json:"from,omitempty"`
	To   int32 `json:"to,omitempty"`
}

func restorePreviousAnalysis(kv *badger.Badger, from, to, interval int32) (intervals []Analyzed) {
	upper := upperBoundary(from, interval)
	lower := lowerBoundary(to, interval)
	if lower-upper >= interval {
		for i := upper; i < lower; i += interval {
			r, err := kv.Read(fmt.Sprintf("int%d-%d", upper, lower))
			if err != nil {
				break
			}
			var analyzed Analyzed
			err = json.Unmarshal(r, &analyzed)
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				break
			}
			intervals = append(intervals, analyzed)
		}
	}
	return
}

func extractRange(kv *badger.Badger, r Range, interval int32, vuln map[int32][]byte) (err error) {
	upper := upperBoundary(r.From, interval)
	lower := lowerBoundary(r.To, interval)
	if lower-upper >= interval {
		var analyzed Analyzed
		analyzed.From = upper
		analyzed.To = lower
		analyzed.Vulnerabilites = vuln
		var a []byte
		a, err = json.Marshal(analyzed)
		if err != nil {
			return
		}
		err = kv.Store(fmt.Sprintf("int%d-%d", upper, lower), a)
	}

	return
}

func mergeMaps(args ...map[int32][]byte) (merged map[int32][]byte) {
	merged = make(map[int32][]byte)
	for _, arg := range args {
		for height, perc := range arg {
			merged[height] = perc
		}
	}
	return
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, export bool) (vuln map[int32][]byte, err error) {
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

	pool := task.New(runtime.NumCPU() * 3)
	lock := sync.RWMutex{}
	store := make(map[string][]byte)
	vuln = make(map[int32][]byte)

	interval := int32(50000)
	ranges := []Range{Range{from, to}}
	analyzed := restorePreviousAnalysis(kv, from, to, interval)
	for i, a := range analyzed {
		if i == 0 {
			if a.From > from {
				ranges[0].To = a.From - 1
			} else {
				ranges[0].To = a.From
			}
		}
		if i == len(analyzed)-1 && a.To < to {
			ranges = append(ranges, Range{a.To + 1, to})
		}

		fmt.Println("updated from to", ranges)
	}

	for _, r := range ranges {
		from, to := r.From, r.To
		for i := from; i <= to; i++ {
			block, e := db.GetBlockFromHeight(i)
			if e != nil {
				if e.Error() == "Key not found" {
					break
				}
				err = e
				return
			}
			logger.Debug("Analysis", "Analyzing block", logger.Params{"height": block.Height, "hash": block.ID})

			for _, tx := range block.Transactions {
				logger.Debug("Analysis", fmt.Sprintf("Analyzing transaction %s", tx.TxID), logger.Params{})
				worker := Worker{
					height: block.Height,
					tx:     &tx,
					c:      c,
					lock:   &lock,
					store:  store,
					vuln:   vuln,
				}
				pool.Do(&worker)
			}

			logger.Debug("Analysis", fmt.Sprintf("Blocks untill %d analyzed", i), logger.Params{})
		}
	}

	pool.Shutdown()

	if _, ok := store[""]; ok {
		delete(store, "")
	}
	if err := kv.StoreBatch(store); err != nil {
		logger.Error("Analysis", err, logger.Params{})
	}

	for _, r := range ranges {
		if e := extractRange(kv, r, interval, vuln); e != nil {
			logger.Error("Analysis", e, logger.Params{})
		}
	}

	maps := []map[int32][]byte{vuln}
	for _, a := range analyzed {
		maps = append(maps, a.Vulnerabilites)
	}
	vuln = mergeMaps(maps...)

	if export {
		data := heuristics.ExtractPercentages(vuln, from, to)
		err = PlotHeuristicsTimeline(data, from)
		// err = HeuristicsPercentages(data, from)
	} else {
		data := heuristics.ExtractGlobalPercentages(vuln, from, to)
		err = GlobalPercentages(data, export)
	}

	return
}

// GlobalPercentages prints a table with percentages of heuristics success rate based on passed analysis
func GlobalPercentages(data []float64, export bool) (err error) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")
	for h, perc := range data {
		table.Append([]string{heuristics.Heuristic(h).String(), fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
	return
}

// PlotHeuristicsTimeline plots timeseries of heuristics percentage effectiveness for each block representing time series
func PlotHeuristicsTimeline(data map[int32][]float64, min int32) (err error) {
	coordinates := make(map[string]plot.Coordinates)

	x := make([]float64, len(data))
	for height := range data {
		x[height-min] = float64(height)
	}

	for h := 0; h < heuristics.SetCardinality(); h++ {
		y := make([]float64, len(data))
		for height, vulnerabilites := range data {
			y[int(height-min)] = vulnerabilites[h] + float64(h)
		}
		coordinates[heuristics.Heuristic(h).String()] = plot.Coordinates{X: x, Y: y}
	}

	err = plot.MultipleLineChart("Heuristics timeline", "blocks", "heuristics effectiveness", coordinates)

	return
}
