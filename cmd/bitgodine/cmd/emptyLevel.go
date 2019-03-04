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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// emptyLevelCmd represents the emptyLevel command
var emptyLevelCmd = &cobra.Command{
	Use:   "emptyLevel",
	Short: "Empty Level database",
	Long:  "Command to erase store info in Level instance of block database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("emptyLevel called")
		dir, err := ioutil.ReadDir(filepath.Join(DBConf().Dir, DBConf().Name))
		if err != nil {
			logger.Error("Emtpy Level DB", err, logger.Params{})
			return
		}
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{filepath.Join(DBConf().Dir, DBConf().Name), d.Name()}...))
		}
		logger.Info("Empty Level", "all blocks are removed", logger.Params{})
	},
}

func init() {
	rootCmd.AddCommand(emptyLevelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// emptyLevelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// emptyLevelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
