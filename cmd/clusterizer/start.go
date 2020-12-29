package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine/internal/clusterizer/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/disjoint/disk"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
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

		db, err := kv.NewDB()
		defer db.Close()

		set, err := disk.NewDisjointSet(db, true, true)
		if err != nil {
			logger.Error("Start", err, logger.Params{})
			os.Exit(-1)
		}
		if err := disk.RestorePersistentSet(&set); err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
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
