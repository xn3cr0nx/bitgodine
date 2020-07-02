package block

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"

	badgerStorage "github.com/xn3cr0nx/bitgodine/pkg/badger/storage"
	tikvStorage "github.com/xn3cr0nx/bitgodine/pkg/tikv/storage"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		var db storage.DB
		if viper.GetString("db") == "tikv" {
			db, err := tikvStorage.NewKV(tikvStorage.Conf(viper.GetString("tikv")), nil)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()

		} else if viper.GetString("db") == "badger" {
			db, err := badgerStorage.NewKV(badgerStorage.Conf(viper.GetString("badger")), nil, false)
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
