package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/cmd/block"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/cmd/transaction"
)

// eraseCmd represents the erase command
var eraseCmd = &cobra.Command{
	Use:   "erase",
	Short: "Removes all stored data",
	Long:  "Erases blocks stored on badger and transaction graph stored in dgraph",
	Run: func(cmd *cobra.Command, args []string) {
		transactionRmCommand, _, err := transaction.TransactionCmd.Find([]string{"rm"})
		if err != nil {
			fmt.Println("error", err)
			return
		}
		go transactionRmCommand.Run(cmd, args)

		blockRmCommand, _, err := block.BlockCmd.Find([]string{"rm"})
		if err != nil {
			fmt.Println("error", err)
			return
		}
		blockRmCommand.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(eraseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eraseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eraseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
