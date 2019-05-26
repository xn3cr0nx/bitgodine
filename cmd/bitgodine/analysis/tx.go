package analysis

import (
	"errors"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/analysis"
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
			logger.Error("Analyze", errors.New("Missing transaction hash"), logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Analyze", "Analyzing...", logger.Params{"tx": args[0]})

		tx, err := dgraph.GetTx(args[0])
		if err != nil {
			logger.Error("Analyze Transactions", err, logger.Params{})
			os.Exit(-1)
		}
		if len(tx.Outputs) <= 1 {
			logger.Error("Analysis Transaction", errors.New("The transaction cannot be analyzed, less than 2 tx output"), logger.Params{})
			os.Exit(-1)
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
