package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/clusterizer/bitcoin"
	"github.com/xn3cr0nx/bitgodine/pkg/badger"
	"github.com/xn3cr0nx/bitgodine/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/disjoint/disk"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

var (
	csv bool
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Creates clusters from synced blocks",
	Long: `Fetch block by block and creates a persistent
version of the cluster of addresses. The cluster is stored
in a persistent way in storage layer.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Start", "Start called", logger.Params{})

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		db, err := kv.NewKV(kv.Conf(viper.GetString("badger")), c, false)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer db.Close()

		set, err := disk.NewDisjointSet(badger.Conf(viper.GetString("disjoint")), true, true)
		if err != nil {
			logger.Error("Start", err, logger.Params{})
			os.Exit(-1)
		}
		if err := disk.RestorePersistentSet(&set); err != nil {
			// TODO: this is package implementation dependent, error should be generic
			if err.Error() != "Key not found" {
				logger.Error("Start", err, logger.Params{})
				os.Exit(-1)
			}
		}

		pg, err := postgres.NewPg(postgres.Conf())
		if err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
			os.Exit(-1)
		}
		if err := pg.Connect(); err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
			os.Exit(-1)
		}
		defer pg.Close()

		interrupt := make(chan int)
		done := make(chan int)
		bc := bitcoin.NewClusterizer(&set, db, pg, c, interrupt, done)
		bc.Clusterize()

		cltzCount, err := bc.Done()
		if err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
			os.Exit(-1)
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)
	},
}
