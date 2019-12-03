package analysis

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
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
	c         *echo.Context
	height    int32
	tx        *models.Tx
	vuln      map[string]byte
	vulnLock  *sync.RWMutex
	store     map[string][]byte
	storeLock *sync.RWMutex
	chain     map[int32]map[string]byte
	chainLock *sync.RWMutex
}

// Work method to make Worker compatible with task pool worker interface
func (w *Worker) Work() {
	if len(w.tx.Vout) <= 1 {
		// TODO: we are not considering coinbase 1 output txs in heuristics analysis
		// w.vulnLock.Lock()
		// defer w.vulnLock.Unlock()
		// w.vuln[w.tx.TxID] = 0
		w.storeLock.Lock()
		defer w.storeLock.Unlock()
		w.store[w.tx.TxID] = []byte{0}
		return
	}
	v, err := AnalyzeTx(w.c, w.tx.TxID)
	if err != nil {
		return
	}
	w.vulnLock.Lock()
	defer w.vulnLock.Unlock()
	w.vuln[w.tx.TxID] = v
	w.storeLock.Lock()
	defer w.storeLock.Unlock()
	w.store[w.tx.TxID] = []byte{v}
	w.chainLock.Lock()
	defer w.chainLock.Unlock()
	w.chain[w.height][w.tx.TxID] = v
}

// AnalyzeBlocks fetches stored block progressively and apply heuristics in contained transactions
func AnalyzeBlocks(c *echo.Context, from, to int32, step int32, export bool) (vuln map[string]byte, err error) {
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
	numCPUs := runtime.NumCPU()
	pool := task.New(numCPUs * 3)
	vuln = make(map[string]byte)
	vulnLock := sync.RWMutex{}
	store := make(map[string][]byte)
	storeLock := sync.RWMutex{}
	chain := make(map[int32]map[string]byte)
	chainLock := sync.RWMutex{}

	for i := from; i <= to; i++ {
		block, e := db.GetBlockFromHeight(i)
		if e != nil {
			if e.Error() == "Key not found" {
				break
			}
			err = e
			return
		}
		logger.Info("Analysis", "Analyzing block", logger.Params{"height": block.Height, "hash": block.ID})

		txLen := len(block.Transactions)
		chain[block.Height] = make(map[string]byte, txLen)
		var wg sync.WaitGroup
		wg.Add(txLen)
		for i := range block.Transactions {
			go func(height int32, tx models.Tx) {
				defer wg.Done()
				logger.Debug("Analysis", fmt.Sprintf("Analyzing transaction %s", tx.TxID), logger.Params{})
				worker := Worker{
					height:    height,
					tx:        &tx,
					c:         c,
					vuln:      vuln,
					vulnLock:  &vulnLock,
					store:     store,
					storeLock: &storeLock,
					chain:     chain,
					chainLock: &chainLock,
				}
				pool.Do(&worker)
			}(block.Height, block.Transactions[i])
		}
		wg.Wait()

		logger.Debug("Analysis", fmt.Sprintf("Blocks untill %d analyzed", i), logger.Params{})
	}

	pool.Shutdown()

	if err := kv.StoreBatch(store); err != nil {
		logger.Error("Analysis", err, logger.Params{})
	}

	// var res []byte
	// for _, v := range vuln {
	// 	res = append(res, v)
	// }
	// err = GlobalPercentages(res, export)

	err = PlotHeuristicsTimeline(chain)

	return
}

// GlobalPercentages prints a table with percentages of heuristics success rate based on passed analysis
func GlobalPercentages(analysis []byte, export bool) (err error) {
	var percentages []float64
	for h := 0; h < heuristics.SetCardinality(); h++ {
		if len(analysis) == 0 {
			percentages = append(percentages, 0)
			continue
		}

		counter := 0
		for _, a := range analysis {
			if heuristics.VulnerableMask(a, h) {
				counter++
			}
		}
		percentages = append(percentages, float64(counter)/float64(len(analysis)))
	}

	if export {
		err = plot.HeuristicsPercentages(percentages)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")
	for h, perc := range percentages {
		table.Append([]string{heuristics.Heuristic(h).String(), fmt.Sprintf("%4.2f", perc*100)})
	}
	table.Render()
	return
}

// PlotHeuristicsTimeline plots timeseries of heuristics percentage effectiveness for each block representing time series
func PlotHeuristicsTimeline(data map[int32]map[string]byte) (err error) {
	coordinates := make(map[string][]plot.Coordinates)

	keys := make([]int, len(data))
	for k := range data {
		keys[k] = int(k)
	}
	sort.Ints(keys)

	for h := 0; h < heuristics.SetCardinality(); h++ {
		heuristic := heuristics.Heuristic(h).String()
		for _, k := range keys {
			if len(data[int32(k)]) == 0 {
				coordinates[heuristic] = append(coordinates[heuristic], plot.Coordinates{X: float64(k), Y: 0})
				continue
			}
			counter := 0
			for _, a := range data[int32(k)] {
				if heuristics.VulnerableMask(a, h) {
					counter++
				}
			}
			percentage := float64(counter) / float64(len(data[int32(k)]))
			coordinates[heuristic] = append(coordinates[heuristic], plot.Coordinates{X: float64(k), Y: percentage})

		}
	}

	err = plot.LineChart("Heuristics timeline", "blocks", "heuristics effectiveness", coordinates)

	return
}
