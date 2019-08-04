package skipped

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/dbblocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		skippedBlocksStorage, err := dbblocks.NewDbBlocks(&db.Config{
			Dir: viper.GetString("dbDir"),
		})
		if err != nil {
			logger.Error("Block skipped ls", err, logger.Params{})
			os.Exit(-1)
		}
		stored, err := skippedBlocksStorage.StoredBlocks()
		if err != nil {
			logger.Error("Block skipped ls", err, logger.Params{})
			os.Exit(-1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Prev Block Hash"})
		for _, block := range stored {
			table.Append([]string{block})
		}
		table.Render()
	},
}
