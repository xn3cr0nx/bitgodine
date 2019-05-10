package address

import (
	"errors"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// occurencesCmd represents the list of address occurences cmd
var occurencesCmd = &cobra.Command{
	Use:   "occurences",
	Short: "Show the list of txs hash where the address appears",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Error("Address", errors.New("Missing address"), logger.Params{})
			os.Exit(1)
		}

		address, err := btcutil.DecodeAddress(args[0], &chaincfg.MainNetParams)
		if err != nil {
			logger.Error("Address", err, logger.Params{})
			os.Exit(1)
		}
		occurences, err := dgraph.GetAddressOccurences(&address)
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
