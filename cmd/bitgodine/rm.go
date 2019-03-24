package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/block"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/transaction"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes all stored data",
	Long:  "Removess blocks stored on badger and transaction graph stored in dgraph",
	Run: func(cmd *cobra.Command, args []string) {
		transactionRmCommand, _, err := transaction.TransactionCmd.Find([]string{"rm"})
		if err != nil {
			fmt.Println("error", err)
			return
		}
		transactionRmCommand.Run(cmd, args)

		blockRmCommand, _, err := block.BlockCmd.Find([]string{"rm"})
		if err != nil {
			fmt.Println("error", err)
			return
		}
		blockRmCommand.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
