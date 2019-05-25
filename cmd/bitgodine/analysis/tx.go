package analysis

import (
	"errors"
	"os"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/analysis"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var changeOutput bool

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
		if len(tx.MsgTx().TxOut) <= 1 {
			logger.Error("Analysis Transaction", errors.New("The transaction cannot be analyzed, less than 2 tx output"), logger.Params{})
			return
		}

		table := tablewriter.NewWriter(os.Stdout)

		if viper.GetBool("analysis.tx.change") {
			table.SetHeader([]string{"Heuristic", "Vout"})
			// table.SetBorder(false)
			privacy := analysis.TxChange(&tx)
			for i, p := range privacy {
				table.SetColumnColor(
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor})
				table.Append([]string{heuristics.Heuristic(i).String(), p})
			}
		} else {
			table.SetHeader([]string{"Heuristic", "Privacy"})
			// table.SetBorder(false)
			privacy := analysis.TxSingleCore(&tx)
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
		}
		table.Render()
	},
}

func init() {
	txCmd.PersistentFlags().BoolVar(&changeOutput, "change", false, "Specify to print vout of foreseen change output instead of heuristic vulnerability")
	viper.SetDefault("analysis.tx.change", false)
	viper.BindPFlag("analysis.tx.change", txCmd.PersistentFlags().Lookup("change"))
}
