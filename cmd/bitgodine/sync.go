package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/allegro/bigcache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/cache"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/dbblocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/persistent"
	"github.com/xn3cr0nx/bitgodine_code/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var (
	csv    bool
	output string
)

// BadgerConf exports the Config object to initialize indexing dgraph
func BadgerConf() *db.Config {
	return &db.Config{
		Dir: viper.GetString("dbDir"),
	}
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Parses the blockchain to sync it",
	Long: `Parses the blockchain, from the last point,
if the synced is being previously performed.
The parsing stores blocks and transaction and creates clusters to provide
data representation to analyze the blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Sync", "Sync called", logger.Params{})

		skippedBlocksStorage, err := dbblocks.NewDbBlocks(BadgerConf())
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		bigCache, err := cache.Instance(bigcache.DefaultConfig(2 * time.Minute))
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		b := blockchain.Instance(BitcoinNet)
		b.Read()
		set := persistent.NewDisjointSet(dgraph.Instance(nil))

		if err := persistent.RestorePersistentSet(&set); err != nil {
			if err.Error() != "Cluster not found" {
				logger.Error("Blockchain", err, logger.Params{})
				os.Exit(-1)
			}
		}

		cltz := visitor.NewClusterizer(&set)
		interrupt := make(chan int)
		done := make(chan int)

		bp := bitcoin.NewParser(b, cltz, skippedBlocksStorage, bigCache, interrupt, done)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go handleInterrupt(cltz, c, interrupt, done)

		bp.Walk()

		if viper.GetBool("sync.csv") {
			cltzCount, err := cltz.Done()
			if err != nil {
				logger.Error("Blockchain test", err, logger.Params{})
			}
			fmt.Printf("Exported Clusters: %v\n", cltzCount)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.PersistentFlags().BoolVar(&csv, "csv", false, "Creates output csv file with cluster when program ends")
	viper.SetDefault("sync.csv", false)
	viper.BindPFlag("sync.csv", syncCmd.PersistentFlags().Lookup("csv"))
}

func handleInterrupt(cltz visitor.BlockchainVisitor, c chan os.Signal, interrupt, done chan int) {
	for sig := range c {
		logger.Info("Sync", "Killing the application", logger.Params{"signal": sig})
		interrupt <- 1
		if viper.GetBool("sync.csv") {
			cltzCount, err := cltz.Done()
			if err != nil {
				logger.Error("Sync", err, logger.Params{})
			}
			logger.Info("Sync", fmt.Sprintf("Exported Clusters: %v\n", cltzCount), logger.Params{})
		}
		done <- 1
	}
}
