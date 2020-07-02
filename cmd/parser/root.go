package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xn3cr0nx/bitgodine/cmd/parser/address"
	"github.com/xn3cr0nx/bitgodine/cmd/parser/block"
	"github.com/xn3cr0nx/bitgodine/cmd/parser/transaction"

	"github.com/btcsuite/btcd/chaincfg"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	_ "net/http/pprof"
)

var (
	cfgFile, network, bitgodineDir, blocksDir, db, dbDir, btcClientHost, btcClientEp, btcClientUser, btcClientPass, btcClientCerts string
	skippedLimit, startFile, restoredBlocks                                                                                        int
	debug, realtime                                                                                                                bool
	BitcoinNet                                                                                                                     chaincfg.Params
)

var rootCmd = &cobra.Command{
	Use:   "parser",
	Short: "Go implementation of Bitiodine",
	Long: `Go implementation of Bitcoin forensic analysis tool to	investigate blockchain and Bitcoin malicious flows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// trace.Start(os.Stdout)
		// defer trace.Stop()

		logger.Setup()

		net, _ := cmd.Flags().GetString("network")
		switch net {
		case "mainnet":
			BitcoinNet = chaincfg.MainNetParams
		case "testnet3":
			BitcoinNet = chaincfg.TestNet3Params
		case "regtest":
			BitcoinNet = chaincfg.RegressionNetParams
		default:
			logger.Panic("Initializing network", errors.New("Network not found"), logger.Params{"provided": net})
		}
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

	rootCmd.AddCommand(startCmd)
	// Adds subdirectories command
	rootCmd.AddCommand(block.BlockCmd)
	rootCmd.AddCommand(transaction.TransactionCmd)
	rootCmd.AddCommand(address.AddressCmd)

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

	rootCmd.PersistentFlags().IntVar(&skippedLimit, "skipped", 50000, "Sets allowed number of skipped blocks")
	rootCmd.PersistentFlags().IntVar(&startFile, "file", 0, "Sets the data file to start parsing from")
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
	viper.SetDefault("file", 0)
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
	viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("restored", rootCmd.PersistentFlags().Lookup("restored"))

	viper.SetEnvPrefix("parser")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if value, ok := os.LookupEnv("CONFIG_FILE"); ok {
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
