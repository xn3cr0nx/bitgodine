package analysis

import (
	"errors"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/forward"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// forwardCmd represents the forward command
var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Apply forward heuristic to transaction",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Error("Analyze forward", errors.New("Missing transaction hash"), logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Analyze forward", "Analyzing...", logger.Params{"tx": args[0]})

		tx, err := dgraph.GetTx(args[0])
		if err != nil {
			logger.Error("Analyze forward", err, logger.Params{})
			os.Exit(-1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Heuristic", "Transaction", "Privacy"})
		table.SetBorder(false)

		if forward.Vulnerable(&tx) {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor})
			table.Append([]string{"Forward", args[0], "✓"})
		} else {
			table.SetColumnColor(
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor})
			table.Append([]string{"Forward", args[0], "x"})
		}

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// forwardCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// forwardCmd.Flags().BoolP("toggle", "t", faforwarde, "Help message for toggle")
// }
