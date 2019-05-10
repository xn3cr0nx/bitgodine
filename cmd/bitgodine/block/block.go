package block

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var verbose bool

// BlockCmd represents the block command
var BlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage block",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logger.Error("Block", errors.New("Missing block hash or height"), logger.Params{})
		}

		height, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("Block", errors.New("Cannot parse passed height"), logger.Params{})
		}
		block, err := dgraph.GetBlockFromHeight(int32(height))
		if err != nil {
			logger.Error("Block", err, logger.Params{})
		}

		if viper.GetBool("block.verbose") {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"Hash", block.Hash})
			table.Append([]string{"Height", fmt.Sprint(block.Height)})
			table.Append([]string{"PrevBlock", block.PrevBlock})
			table.Append([]string{"Timestamp", fmt.Sprint(block.Time)})
			table.Append([]string{"Merkle Root", block.MerkleRoot})
			table.Append([]string{"Bits", fmt.Sprint(block.Bits)})
			table.Append([]string{"Nonce", fmt.Sprint(block.Nonce)})
			table.Render()
		} else {
			fmt.Println("Block Hash", block.Hash)
		}
	},
}

func init() {
	BlockCmd.AddCommand(lsCmd)
	BlockCmd.AddCommand(rmCmd)
	BlockCmd.AddCommand(heightCmd)

	BlockCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Specify verbose output to show all block info")
	viper.SetDefault("block.verbose", false)
	viper.BindPFlag("block.verbose", BlockCmd.PersistentFlags().Lookup("verbose"))
}
