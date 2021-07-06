package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	_ "net/http/pprof"
)

var (
	cfgFile, network, bitgodineDir, blocksDir, db, dbDir, btcClientHost, btcClientEp, btcClientUser, btcClientPass, btcClientCerts string
	startFile, restoredBlocks                                                                                                      int
	debug, realtime                                                                                                                bool
)

var rootCmd = &cobra.Command{
	Use:   "parser",
	Short: "Parses the blockchain to sync blocks",
	Long: `Parses the blockchain, from the last point,
	if the synced is being previously performed.
	The parsing stores blocks and transaction and creates clusters to provide
	data representation to analyze the blockchain.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// trace.Start(os.Stdout)
		// defer trace.Stop()
		logger.Setup()
	},
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

		logger.Info("Bitgodine", "Parsing completed", logger.Params{})
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Adds root flags and persistent flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bitgodine.yaml)")
	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", chaincfg.MainNetParams.Name, "Specify blockchain network - mainnet - testnet3 - regtest [default: mainnet]")
	rootCmd.PersistentFlags().BoolVarP(&realtime, "realtime", "r", true, "Specify whether real time parsing is performed (client connection)")

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	rootCmd.PersistentFlags().StringVar(&bitgodineDir, "bitgodineDir", bitgodineFolder, "Sets the folder containing configuration files and stored data")

	rootCmd.PersistentFlags().StringVarP(&blocksDir, "blocksDir", "b", hd, "Sets the path to the bitcoind blocks directory")

	rootCmd.PersistentFlags().StringVar(&db, "db", filepath.Join(bitgodineFolder, "badger"), "Sets the path to the indexing db files")
	rootCmd.PersistentFlags().StringVar(&dbDir, "dbDir", filepath.Join(bitgodineFolder, "badger", "skipped"), "Sets the path to the indexing db files")

	rootCmd.PersistentFlags().StringVar(&btcClientHost, "btcHost", "localhost:8333", "Specify bitcoin client host")
	rootCmd.PersistentFlags().StringVar(&btcClientEp, "btcEp", "ws", "Specify bitcoin client endpoint protocol")
	rootCmd.PersistentFlags().StringVar(&btcClientUser, "btcUser", "bitcoinrpc", "Specify bitcoin client connection user")
	rootCmd.PersistentFlags().StringVar(&btcClientPass, "btcPass", "pass", "Specify bitcoin client connection password")
	rootCmd.PersistentFlags().StringVar(&btcClientCerts, "btcCerts", "~/.bitcoin/rpc.cert", "Specify bitcoin client connection certificates")

	rootCmd.PersistentFlags().IntVar(&restoredBlocks, "restored", 50000, "Sets the number of blocks to restore before the current synced height")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	viper.SetDefault("debug", false)
	viper.SetDefault("realtime", false)
	viper.SetDefault("network", chaincfg.MainNetParams.Name)
	viper.SetDefault("bitgodineDir", bitgodineFolder)
	viper.SetDefault("blocksDir", hd)
	viper.SetDefault("db", filepath.Join(bitgodineFolder, "badger"))
	viper.SetDefault("dbDir", filepath.Join(bitgodineFolder, "badger", "skipped"))
	viper.SetDefault("btcHost", "localhost:8333")
	viper.SetDefault("btcEp", "ws")
	viper.SetDefault("btcUser", "bitcoinrpc")
	viper.SetDefault("btcPass", "pass")
	viper.SetDefault("btcCerts", "~/.bitcoin/rpc.cert")
	viper.SetDefault("skipped", 50000)
	viper.SetDefault("restored", 50000)

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("realtime", rootCmd.PersistentFlags().Lookup("realtime"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("bitgodineDir", rootCmd.PersistentFlags().Lookup("bitgodineDir"))
	viper.BindPFlag("blocksDir", rootCmd.PersistentFlags().Lookup("blocksDir"))
	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("dbDir", rootCmd.PersistentFlags().Lookup("dbDir"))
	viper.BindPFlag("bitcoin.client.btcHost", rootCmd.PersistentFlags().Lookup("btcHost"))
	viper.BindPFlag("bitcoin.client.btcEp", rootCmd.PersistentFlags().Lookup("btcEp"))
	viper.BindPFlag("bitcoin.client.btcUser", rootCmd.PersistentFlags().Lookup("btcUser"))
	viper.BindPFlag("bitcoin.client.btcPass", rootCmd.PersistentFlags().Lookup("btcPass"))
	viper.BindPFlag("bitcoin.client.btcCerts", rootCmd.PersistentFlags().Lookup("btcCerts"))
	viper.BindPFlag("skipped", rootCmd.PersistentFlags().Lookup("skipped"))
	viper.BindPFlag("restored", rootCmd.PersistentFlags().Lookup("restored"))

	viper.SetEnvPrefix("parser")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if value, ok := os.LookupEnv("config"); ok {
		viper.SetConfigFile(value)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/server/")
		viper.AddConfigPath("$HOME/.bitgodine/server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.ReadInConfig()
	f := viper.ConfigFileUsed()
	if f != "" {
		fmt.Printf("Found configuration file: %s \n", f)
	}
}
