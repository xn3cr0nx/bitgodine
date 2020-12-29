package main

import (
	"os"

	// "github.com/pkg/profile"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
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

		net, _ := cmd.Flags().GetString("network")
		var network chaincfg.Params
		switch net {
		case "mainnet":
			network = chaincfg.MainNetParams
		case "testnet3":
			network = chaincfg.TestNet3Params
		case "regtest":
			network = chaincfg.RegressionNetParams
		default:
			logger.Panic("Initializing network", errorx.ErrInvalidArgument, logger.Params{"provided": net})
		}

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		db, err := kv.NewDB()
		defer db.Close()

		skippedBlocksStorage := bitcoin.NewSkipped()
		chain := bitcoin.NewBlockchain(db, network)
		// if err := chain.Read(""); err != nil {
		// 	logger.Error("Bitgodine", err, logger.Params{})
		// 	os.Exit(-1)
		// }

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
		bp := bitcoin.NewParser(chain, client, db, skippedBlocksStorage, nil, c, interrupt)

		if err := bp.InfinitelyParse(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
	},
}
