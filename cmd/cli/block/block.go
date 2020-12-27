package block

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/httpx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var verbose bool

// BlockCmd represents the block command
var BlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Get block by height",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || args[0] == "" {
			logger.Error("Block", errorx.ErrInvalidArgument, logger.Params{})
			os.Exit(1)
		}

		var resp string
		var err error
		if block.IsHash(args[0]) {
			viper.Set("verbose", true)
			resp, err = httpx.GET(fmt.Sprintf("%s/api/block/%s", viper.GetString("host"), args[0]), nil)
			if err != nil {
				logger.Error("bitgodine-cli", err, logger.Params{})
				os.Exit(1)
			}
		} else if block.IsHeight(args[0]) {
			resp, err = httpx.GET(fmt.Sprintf("%s/api/block-height/%s", viper.GetString("host"), args[0]), nil)
			if err != nil {
				logger.Error("bitgodine-cli", err, logger.Params{})
				os.Exit(1)
			}
		} else {
			logger.Error("Block", errorx.ErrInvalidArgument, logger.Params{})
			os.Exit(1)
		}

		var b block.BlockOut
		if err := json.Unmarshal([]byte(resp), &b); err != nil {
			logger.Error("bitgodine-cli", err, logger.Params{})
			os.Exit(1)
		}

		if viper.GetBool("verbose") {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"Hash", b.ID})
			table.Append([]string{"Height", fmt.Sprint(b.Height)})
			table.Append([]string{"PrevBlock", b.Previousblockhash})
			table.Append([]string{"Timestamp", fmt.Sprint(b.Timestamp)})
			table.Append([]string{"Merkle Root", b.MerkleRoot})
			table.Append([]string{"Bits", fmt.Sprint(b.Bits)})
			table.Append([]string{"Nonce", fmt.Sprint(b.Nonce)})
			table.Render()
		} else {
			fmt.Println("Block Hash", b.ID)
		}
	},
}

func init() {
	BlockCmd.AddCommand(lsCmd)
	BlockCmd.AddCommand(heightCmd)

	BlockCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Specify verbose output to show all block info")
	viper.SetDefault("verbose", false)
	viper.BindPFlag("verbose", BlockCmd.PersistentFlags().Lookup("verbose"))
}
