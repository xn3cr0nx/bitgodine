package analysis

import (
	"errors"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// optimalCmd represents the optimal command
var optimalCmd = &cobra.Command{
	Use:   "optimal",
	Short: "Apply optimal change heuristic to transaction",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Panic("Analyze", errors.New("Missing transaction hash"), logger.Params{})
		}

		logger.Info("Analyze optimal", "Analyzing...", logger.Params{"tx": args[0]})

		tx, err := dgraph.GetTx(args[0])
		if err != nil {
			logger.Panic("Analyze optimal", err, logger.Params{})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Heuristic", "Transaction", "Privacy"})
		table.SetBorder(false)

		if optimal.Vulnerable(&tx) {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor})
			table.Append([]string{"optimal Chain", args[0], "✓"})
		} else {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor})
			table.Append([]string{"optimal Chain", args[0], "x"})
		}

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// optimalCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// optimalCmd.Flags().BoolP("toggle", "t", faoptimale, "Help message for toggle")
// }
