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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
