package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/clusterizer/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/disjoint/disk"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export stored clusters to csv",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("export", "export called", logger.Params{})

		viper.Set("sync.csv", true)

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		db, err := kv.NewDB()
		defer db.Close()

		set, err := disk.NewDisjointSet(db, true, true)
		if err != nil {
			logger.Error("export", err, logger.Params{})
			os.Exit(-1)
		}
		if err := disk.RestorePersistentSet(&set); err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
				logger.Error("export", err, logger.Params{})
				os.Exit(-1)
			}
		}

		pg, err := postgres.NewPg(postgres.Conf())
		if err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		interrupt := make(chan int)
		done := make(chan int)
		bc := bitcoin.NewClusterizer(&set, db, pg, c, interrupt, done)
		// bc.Clusterize()

		cltzCount, err := bc.Done()
		if err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)
	},
}
