package analysis

import (
	"errors"
	"fmt"
	"math"
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
	if res, ok := ca.Get(txid); ok {
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

	ApplyHeuristics(db, &tx, &vuln)

	if err = kv.Store(txid, []byte{vuln}); err != nil {
		return
	}
	if !ca.Set(txid, vuln, 1) {
		(*c).Logger().Error(err)
	}
	return
}

// Worker wrapper to partecipate in task pool
type Worker struct {
	c         *echo.Context
	tx        *models.Tx
	vuln      map[string]byte
	store     map[string][]byte
	vulnLock  *sync.RWMutex
	storeLock *sync.RWMutex
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

		var wg sync.WaitGroup
		wg.Add(len(block.Transactions))
		for i := range block.Transactions {
			go func(tx models.Tx) {
				defer wg.Done()
				logger.Debug("Analysis", fmt.Sprintf("Analyzing transaction %s", tx.TxID), logger.Params{})
				worker := Worker{
					tx:        &tx,
					c:         c,
					vuln:      vuln,
					store:     store,
					vulnLock:  &vulnLock,
					storeLock: &storeLock,
				}
				pool.Do(&worker)
			}(block.Transactions[i])
		}
		wg.Wait()

		logger.Debug("Analysis", fmt.Sprintf("Blocks untill %d analyzed", i), logger.Params{})
	}

	pool.Shutdown()
	if err := kv.StoreBatch(store); err != nil {
		logger.Error("Analysis", err, logger.Params{})
	}

	var res []byte
	for _, v := range vuln {
		res = append(res, v)
	}
	err = Percentages(res, export)

	return
}

// ApplyHeuristics applies the set of heuristics the the passed transaction
func ApplyHeuristics(db storage.DB, tx *models.Tx, vuln *byte) {
	for h := 0; h < heuristics.SetCardinality(); h++ {
		if heuristics.VulnerableFunction(heuristics.Heuristic(h).String())(db, tx) {
			(*vuln) += byte(math.Pow(2, float64(h+1)))
		}
	}
}

// Percentages prints a table with percentages of heuristics success rate based on passed analysis
func Percentages(analysis []byte, export bool) (err error) {
	var percentages []float64
	for h := 0; h < heuristics.SetCardinality(); h++ {
		if len(analysis) == 0 {
			percentages = append(percentages, 0)
			continue
		}

		counter := 0
		for _, a := range analysis {
			if a&byte(math.Pow(2, float64(h))) > 0 {
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

// func Plot(analysis [][]bool, start, end int) {
// 	logger.Info("Analysis", "Generating plots..", logger.Params{})

// 	line := charts.NewLine()
// 	line.SetGlobalOptions(charts.TitleOpts{Title: "Heuristics Success Rate"}, charts.YAxisOpts{SplitLine: charts.SplitLineOpts{Show: true}})

// 	// var heuristicsAxis []string
// 	// for i := 0; i < 9; i++ {
// 	// 	heuristicsAxis = append(heuristicsAxis, heuristics.Heuristic(i).String())
// 	// }
// 	// fmt.Println("axis", heuristicsAxis)
// 	// line.AddXAxis(heuristicsAxis).AddYAxis("Peeling", []int{90, 70, 90, 70, 90, 70, 90, 70, 90})

// 	line.AddXAxis(generateHeightSeries(start, end))

// 	for heuristic := range analysis[0] {
// 		var series []int
// 		for _, a := range analysis {
// 			if a[heuristic] {
// 				series = append(series, (heuristic*20)+10)
// 			} else {
// 				series = append(series, (heuristic*20)+0)
// 			}
// 		}

// 		// line = line.AddYAxis(heuristics.Heuristic(heuristic).String(), series, charts.AreaStyleOpts{Opacity: 0.2}, charts.LineOpts{Step: true})
// 		line = line.AddYAxis(heuristics.Heuristic(heuristic).String(), series, charts.LineOpts{Step: true})
// 	}

// 	f, err := os.Create("line.html")
// 	if err != nil {
// 		logger.Error("Analysis", err, logger.Params{})
// 		return
// 	}
// 	line.Render(f)
// }
// func generateHeightSeries(start, end int) (series []int) {
// 	length := end - start
// 	// steps := length / 10.0

// 	for i := 0; i <= length; i++ {
// 		series = append(series, start+i)
// 	}

// 	return
// }
