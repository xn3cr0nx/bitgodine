package analysis

import (
	"fmt"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/chenjiandongx/go-echarts/charts"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/backward"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/behaviour"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/forward"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/locktime"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/power"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/reuse"
	class "github.com/xn3cr0nx/bitgodine_code/internal/heuristics/type"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Range applies heuristics to transaction contained in blocks specified in the range
func Range(from, to int32) ([][]bool, error) {
	logger.Info("Analysis", fmt.Sprintf("Analyzing the transactions in blocks between block %d and block %d", from, to), logger.Params{})
	var analysis [][]bool
	for i := from; i <= to; i++ {
		block, err := dgraph.GetBlockHashFromHeight(i)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
			return nil, err
		}
		logger.Debug("Analysis", fmt.Sprintf("Analyzing block %s", block), logger.Params{})
		hash, err := chainhash.NewHashFromStr(block)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
			return nil, err
		}
		b, err := db.GetBlock(hash)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
			return nil, err
		}
		for _, tx := range b.Transactions() {
			logger.Debug("Analysis", fmt.Sprintf("Analyzing transaction %s", tx.Hash().String()), logger.Params{})
			if len(tx.MsgTx().TxOut) <= 1 {
				continue
			}
			res := Tx(&txs.Tx{Tx: *tx})
			analysis = append(analysis, res)
		}
	}

	return analysis, nil
}

// Tx applies all the heuristics to the passed transaction returning a boolean value for each of them
// representing in vulnerable or not
func Tx(tx *txs.Tx) (privacy []bool) {
	privacy = append(privacy, peeling.IsPeelingChain(tx))
	privacy = append(privacy, power.Vulnerable(tx))
	privacy = append(privacy, optimal.Vulnerable(tx))
	privacy = append(privacy, class.Vulnerable(tx))
	privacy = append(privacy, reuse.Vulnerable(tx))
	privacy = append(privacy, locktime.Vulnerable(tx))
	privacy = append(privacy, behaviour.Vulnerable(tx))
	privacy = append(privacy, forward.Vulnerable(tx))
	privacy = append(privacy, backward.Vulnerable(tx))
	return privacy
}

// TxChange applies all the heuristics to the passed transaction returning the vout of the change output for each of them
func TxChange(tx *txs.Tx) (privacy []string) {
	output, err := peeling.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = power.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = optimal.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = class.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = reuse.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = locktime.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = behaviour.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = forward.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	output, err = backward.ChangeOutput(tx)
	if err != nil {
		privacy = append(privacy, "unknown")
	} else {
		privacy = append(privacy, strconv.Itoa(int(output)))
	}
	return privacy
}

func Percentages(analysis [][]bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")

	for heuristic := range analysis[0] {
		counter := 0
		for _, a := range analysis {
			if a[heuristic] {
				counter++
			}
		}

		perc := float64(counter) / float64(len(analysis))

		// table.SetColumnColor(
		// 	tablewriter.Colors{},
		// 	tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor})
		table.Append([]string{heuristics.Heuristic(heuristic).String(), fmt.Sprintf("%4.2f", perc*100)})
	}

	table.Render()
}

func Plot(analysis [][]bool, start, end int) {
	logger.Info("Analysis", "Generating plots..", logger.Params{})

	line := charts.NewLine()
	line.SetGlobalOptions(charts.TitleOpts{Title: "Heuristics Success Rate"}, charts.YAxisOpts{SplitLine: charts.SplitLineOpts{Show: true}})

	// var heuristicsAxis []string
	// for i := 0; i < 9; i++ {
	// 	heuristicsAxis = append(heuristicsAxis, heuristics.Heuristic(i).String())
	// }
	// fmt.Println("axis", heuristicsAxis)
	// line.AddXAxis(heuristicsAxis).AddYAxis("Peeling", []int{90, 70, 90, 70, 90, 70, 90, 70, 90})

	line.AddXAxis(generateHeightSeries(start, end))

	for heuristic := range analysis[0] {
		var series []int
		for _, a := range analysis {
			if a[heuristic] {
				series = append(series, (heuristic*20)+10)
			} else {
				series = append(series, (heuristic*20)+0)
			}
		}

		// line = line.AddYAxis(heuristics.Heuristic(heuristic).String(), series, charts.AreaStyleOpts{Opacity: 0.2}, charts.LineOpts{Step: true})
		line = line.AddYAxis(heuristics.Heuristic(heuristic).String(), series, charts.LineOpts{Step: true})
	}

	f, err := os.Create("line.html")
	if err != nil {
		logger.Error("Analysis", err, logger.Params{})
		return
	}
	line.Render(f)
}

func generateHeightSeries(start, end int) (series []int) {
	length := end - start
	// steps := length / 10.0

	for i := 0; i <= length; i++ {
		series = append(series, start+i)
	}

	return
}
