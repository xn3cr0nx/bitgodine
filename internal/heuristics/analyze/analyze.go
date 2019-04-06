package analyze

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	logger.Info("Analyze", fmt.Sprintf("Analyzing the transactions in blocks between block %d and block %d", from, to), logger.Params{})
	var analysis [][]bool
	for i := from; i <= to; i++ {
		block, err := dgraph.GetBlockHashFromHeight(i)
		if err != nil {
			logger.Error("Analyze", err, logger.Params{})
			return nil, err
		}
		logger.Debug("Analyze", fmt.Sprintf("Analyzing block %s", block), logger.Params{})
		hash, err := chainhash.NewHashFromStr(block)
		if err != nil {
			logger.Error("Analyze", err, logger.Params{})
			return nil, err
		}
		b, err := db.GetBlock(hash)
		if err != nil {
			logger.Error("Analyze", err, logger.Params{})
			return nil, err
		}
		for _, tx := range b.Transactions() {
			logger.Debug("Analyze", fmt.Sprintf("Analyzing transaction %s", tx.Hash().String()), logger.Params{})
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

func Percentages(analysis [][]bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Heuristic", "%"})
	// table.SetBorder(false)
	table.SetCaption(true, "Heuristics success rate")

	tot := len(analysis)

	for heuristic := range analysis[0] {
		counter := 0
		for _, a := range analysis {
			if a[heuristic] {
				counter++
			}
		}

		perc := float64(counter) / float64(tot)

		// table.SetColumnColor(
		// 	tablewriter.Colors{},
		// 	tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor})
		table.Append([]string{heuristics.Heuristic(heuristic).String(), fmt.Sprintf("%4.2f", perc)})
	}

	table.Render()
}
