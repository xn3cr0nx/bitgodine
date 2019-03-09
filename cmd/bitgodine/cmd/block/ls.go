package block

import (
	"fmt"

	"github.com/spf13/cobra"
	bdg "github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		blocks, err := bdg.StoredBlocks()
		if err != nil {
			logger.Error("blocks ls", err, logger.Params{})
			return
		}
		fmt.Println("Number of blocks:", len(blocks))
		fmt.Println("Height - Hash")
		for i, b := range blocks {
			fmt.Println(i, "     - ", b)
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
