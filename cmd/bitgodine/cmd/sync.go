package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Check sync status of blockchain",
	Long:  `Check sync status of blockchain and provides info`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called " + BitcoinNet.Name)
		database, err := db.DB(DBConf())
		if err != nil {
			logger.Error("sync", err, logger.Params{})
			return
		}
		// defer os.RemoveAll(filepath.Join(DBConf().Dir, DBConf().Name))
		defer (*database).Close()

		b := blockchain.Instance(BitcoinNet)
		b.Read()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
