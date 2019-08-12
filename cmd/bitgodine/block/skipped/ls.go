package skipped

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/badger/skipped"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := skipped.NewSkipped(&badger.Config{
			Dir: viper.GetString("dbDir"),
		}, false)
		if err != nil {
			logger.Error("Block skipped ls", err, logger.Params{})
			os.Exit(-1)
		}
		stored, err := s.StoredBlocks()
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
