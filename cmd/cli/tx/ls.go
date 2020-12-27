package tx

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine/internal/storage/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		dg := dgraph.Instance(dgraph.Conf(), nil)
		if err := dg.Setup(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		txs, err := dg.GetStoredTxs()
		if err != nil {
			logger.Error("transactions ls", err, logger.Params{})
			return
		}
		fmt.Println("Number of transactions:", len(txs))
		for _, tx := range txs {
			fmt.Println(tx)
		}
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
