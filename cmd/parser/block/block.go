package block

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/storage/redis"
	"github.com/xn3cr0nx/bitgodine/internal/storage/tikv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
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

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		var db storage.DB
		if viper.GetString("db") == "tikv" {
			t, err := tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
			db, err = tikv.NewKV(t, c)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()

		} else if viper.GetString("db") == "badger" {
			bdg, err := badger.NewBadger(badger.Conf(viper.GetString("badger")), false)
			db, err = badger.NewKV(bdg, c)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()
		} else if viper.GetString("db") == "redis" {
			r, err := redis.NewRedis(redis.Conf(viper.GetString("redis")))
			db, err = redis.NewKV(r, c)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()
		}

		height, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("Block", errors.New("Cannot parse passed height"), logger.Params{})
			os.Exit(-1)
		}
		block, err := db.GetBlockFromHeight(int32(height))
		if err != nil {
			logger.Error("Block", err, logger.Params{})
			os.Exit(-1)
		}

		if viper.GetBool("block.verbose") {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"Hash", block.ID})
			table.Append([]string{"Height", fmt.Sprint(block.Height)})
			table.Append([]string{"PrevBlock", block.Previousblockhash})
			table.Append([]string{"Timestamp", fmt.Sprint(block.Timestamp)})
			table.Append([]string{"Merkle Root", block.MerkleRoot})
			table.Append([]string{"Bits", fmt.Sprint(block.Bits)})
			table.Append([]string{"Nonce", fmt.Sprint(block.Nonce)})
			table.Render()
		} else {
			fmt.Println("Block Hash", block.ID)
		}
	},
}

func init() {
	BlockCmd.AddCommand(lsCmd)
	BlockCmd.AddCommand(heightCmd)

	BlockCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Specify verbose output to show all block info")
	viper.SetDefault("block.verbose", false)
	viper.BindPFlag("block.verbose", BlockCmd.PersistentFlags().Lookup("verbose"))
}
