package analysis

import (
	"errors"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/backward"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// backwardCmd represents the backward command
var backwardCmd = &cobra.Command{
	Use:   "backward",
	Short: "Apply backward heuristic to transaction",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Error("Analyze backward", errors.New("Missing transaction hash"), logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Analyze backward", "Analyzing...", logger.Params{"tx": args[0]})

		tx, err := dgraph.GetTx(args[0])
		if err != nil {
			logger.Error("Analyze backward", err, logger.Params{})
			os.Exit(-1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Heuristic", "Transaction", "Privacy"})
		table.SetBorder(false)

		if backward.Vulnerable(&tx) {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor})
			table.Append([]string{"backward", args[0], "✓"})
		} else {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor})
			table.Append([]string{"backward", args[0], "x"})
		}

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// backwardCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// backwardCmd.Flags().BoolP("toggle", "t", fabackwarde, "Help message for toggle")
// }
