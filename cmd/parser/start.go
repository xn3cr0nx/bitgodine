package main

import (
	"os"
	"os/signal"

	// "github.com/pkg/profile"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/utxoset"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"

	"github.com/xn3cr0nx/bitgodine/internal/skipped"
	badgerStorage "github.com/xn3cr0nx/bitgodine/pkg/badger/storage"
	tikvStorage "github.com/xn3cr0nx/bitgodine/pkg/tikv/storage"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Parses the blockchain to sync it",
	Long: `Parses the blockchain, from the last point,
if the synced is being previously performed.
The parsing stores blocks and transaction and creates clusters to provide
data representation to analyze the blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		// defer profile.Start().Stop()
		// defer profile.Start(profile.MemProfile).Stop()

		logger.Info("Start", "Start called", logger.Params{})

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		var db storage.DB
		if viper.GetString("db") == "tikv" {
			db, err := tikvStorage.NewKV(tikvStorage.Conf(viper.GetString("tikv")), c)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()

		} else if viper.GetString("db") == "badger" {
			db, err := badgerStorage.NewKV(badgerStorage.Conf(viper.GetString("badger")), c, false)
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer db.Close()
		}

		skippedBlocksStorage := skipped.NewSkipped()
		utxoset := utxoset.Instance(utxoset.Conf("", true))

		b := blockchain.Instance(db, BitcoinNet)
		b.Read("")

		var client *rpcclient.Client
		if viper.GetBool("parser.realtime") {
			client, err = bitcoin.NewClient()
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			if err := client.NotifyBlocks(); err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
		}

		interrupt := make(chan int)
		done := make(chan int)

		bp := bitcoin.NewParser(b, client, db, skippedBlocksStorage, utxoset, c, interrupt, done)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go handleInterrupt(ch, interrupt, done)

		skipped := viper.GetInt("skipped")
		for {
			if cycleSkipped, err := bp.Walk(skipped); err != nil {
				logger.Error("Bitgodine", err, logger.Params{"skipped": cycleSkipped})
				if err.Error() == "too many skipped blocks, stopping process" {
					skippedBlocksStorage.Empty()
					continue
				}
				os.Exit(-1)
			}
			select {
			case <-done:
				return
			default:
			}

		}

	},
}

func handleInterrupt(c chan os.Signal, interrupt, done chan int) {
	for sig := range c {
		logger.Info("Sync", "Killing the application", logger.Params{"signal": sig})
		interrupt <- 1
		done <- 1
	}
}
