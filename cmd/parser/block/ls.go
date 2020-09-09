package block

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/storage/redis"
	"github.com/xn3cr0nx/bitgodine/internal/storage/tikv"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		var db storage.DB
		if viper.GetString("db") == "tikv" {
			t, err := tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
			db, err = tikv.NewKV(t, nil)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()

		} else if viper.GetString("db") == "badger" {
			bdg, err := badger.NewBadger(badger.Conf(viper.GetString("badger")), false)
			db, err = badger.NewKV(bdg, nil)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()
		} else if viper.GetString("db") == "redis" {
			r, err := redis.NewRedis(redis.Conf(viper.GetString("redis")))
			db, err = redis.NewKV(r, nil)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()
		}

		blocks, err := db.GetStoredBlocks()
		if err != nil {
			logger.Error("blocks ls", err, logger.Params{})
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Height", "Block Hash"})

		for _, block := range blocks {
			table.Append([]string{strconv.Itoa(int(block.Height)), block.ID})
		}

		table.Render()
	},
}
