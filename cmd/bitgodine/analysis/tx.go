package analysis

import (
	"errors"
	"os"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/analysis"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// txCmd represents the tx command
var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Specify transaction hash on which apply implemented heuristics",
	Args:  cobra.ExactArgs(1),
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Panic("Analyze", errors.New("Missing transaction hash"), logger.Params{})
		}

		logger.Info("Analyze", "Analyzing...", logger.Params{"tx": args[0]})

		txHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			logger.Panic("Analyze peeling", err, logger.Params{})
		}
		tx, err := txs.Get(txHash)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Heuristic", "Privacy"})
		// table.SetBorder(false)

		privacy := analysis.Tx(&tx)

		for i, p := range privacy {
			if p {
				table.SetColumnColor(
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor})
				table.Append([]string{heuristics.Heuristic(i).String(), "✓"})
			} else {
				table.SetColumnColor(
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor})
				table.Append([]string{heuristics.Heuristic(i).String(), "x"})
			}
		}

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// txCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// txCmd.Flags().BoolP("toggle", "t", fatxe, "Help message for toggle")
// }
