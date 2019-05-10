package block

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// heightCmd represents the height command
var heightCmd = &cobra.Command{
	Use:   "height",
	Short: "Show the height of the last block stored",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		block, err := dgraph.LastBlock()
		if err != nil {
			logger.Error("blocks height", err, logger.Params{})
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Height", "Block Hash"})

		table.Append([]string{strconv.Itoa(int(block.Height)), block.Hash})

		table.Render()
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// heightCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// heightCmd.Flags().BoolP("toggle", "t", faheighte, "Help message for toggle")
// }
