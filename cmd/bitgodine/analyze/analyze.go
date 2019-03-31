package analyze

import (
	"errors"
	"os"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/power"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/reuse"
	class "github.com/xn3cr0nx/bitgodine_code/internal/heuristics/type"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// AnalyzeCmd represents the Analyze command
var AnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze transactions",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Panic("Analyze", errors.New("Missing transaction hash"), logger.Params{})
		}

		logger.Info("Analyze peeling", "Analyzing...", logger.Params{"tx": args[0]})

		heuristics := []string{
			"Peeling Chain",
			"Power of Ten",
			"Optimal Change",
			"Address Type",
			"Address Reuse",
		}

		txHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			logger.Panic("Analyze peeling", err, logger.Params{})
		}
		tx, err := txs.Get(txHash)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Heuristic", "Privacy"})
		// table.SetBorder(false)

		var privacy []bool
		privacy = append(privacy, peeling.IsPeelingChain(&tx))
		privacy = append(privacy, power.Vulnerable(&tx))
		privacy = append(privacy, optimal.Vulnerable(&tx))
		privacy = append(privacy, class.Vulnerable(&tx))
		privacy = append(privacy, reuse.Vulnerable(&tx))

		for i, p := range privacy {
			if p {
				table.SetColumnColor(
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor})
				table.Append([]string{heuristics[i], "✓"})
			} else {
				table.SetColumnColor(
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor})
				table.Append([]string{heuristics[i], "x"})
			}
		}

		table.Render()
	},
}

func init() {
	AnalyzeCmd.AddCommand(peelingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// AnalyzeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// AnalyzeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
