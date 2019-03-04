package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// emptyGraphCmd represents the emptyGraph command
var emptyGraphCmd = &cobra.Command{
	Use:   "emptyGraph",
	Short: "Empty Graph database",
	Long:  "Command to erase store info in Dgraph instance of graph database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("emptyGraph called")
		err := dgraph.Empty()
		if err != nil {
			logger.Error("Empty Dgraph", err, logger.Params{})
			return
		}
		logger.Info("Empty Dgraph", "all txs are removed", logger.Params{})
	},
}

func init() {
	rootCmd.AddCommand(emptyGraphCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// emptyGraphCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// emptyGraphCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
