// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/cmd/block"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/cmd/transaction"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// eraseCmd represents the erase command
var eraseCmd = &cobra.Command{
	Use:   "erase",
	Short: "Removes all stored data",
	Long:  "Erases blocks stored on leveldb and transaction graph stored in dgraph",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("erase called")
		transaction.TransactionCmd.Run(cmd, args)
		block.BlockCmd.Run(cmd, args)
		logger.Info("Erase", "Bitgodine Erased", logger.Params{})
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
