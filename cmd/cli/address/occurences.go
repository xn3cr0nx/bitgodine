package address

import (
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// occurencesCmd represents the list of address occurences cmd
var occurencesCmd = &cobra.Command{
	Use:   "occurences",
	Short: "Show the list of txs hash where the address appears",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Error("Address", errorx.ErrInvalidArgument, logger.Params{})
			os.Exit(1)
		}

		dg := dgraph.Instance(dgraph.Conf(), nil)
		if err := dg.Setup(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		address, err := btcutil.DecodeAddress(args[0], &chaincfg.MainNetParams)
		if err != nil {
			logger.Error("Address", err, logger.Params{})
			os.Exit(1)
		}
		occurences, err := dg.GetOccurences(address.String())
		if err != nil {
			logger.Error("occurences", err, logger.Params{})
			os.Exit(1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Occurence", "Transaction"})
		for i, a := range occurences {
			table.Append([]string{strconv.Itoa(i), a})
		}

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// occurencesCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// occurencesCmd.Flags().BoolP("toggle", "t", faheighte, "Help message for toggle")
// }
