package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/analysis"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/block"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/cluster"
	"github.com/xn3cr0nx/bitgodine_code/cmd/bitgodine/transaction"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"

	"github.com/btcsuite/btcd/chaincfg"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var (
	cfgFile, network, bitgodineDir, blocksDir, outputDir, dbDir, dgHost string
	dgPort                                                              int
	debug                                                               bool
	BitcoinNet                                                          chaincfg.Params
)

// DBConf exports the Config object to initialize indexing db
func DBConf() *db.Config {
	return &db.Config{
		Dir: viper.GetString("dbDir"),
	}
}

// DGraphConf exports the Config object to initialize indexing dgraph
func DGraphConf() *dgraph.Config {
	return &dgraph.Config{
		Host: viper.GetString("dgHost"),
		Port: viper.GetInt("dgPort"),
	}
}

var rootCmd = &cobra.Command{
	Use:   "bitgodine",
	Short: "Go implementation of Bitiodine",
	Long: `Go implementation of Bitcoin forensic analysis tool to	investigate blockchain and Bitcoin malicious flows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()

		// _, err := db.Instance(DBConf())
		// if err != nil {
		// 	logger.Error("Bitgodine", err, logger.Params{})
		// 	return
		// }

		dg := dgraph.Instance(DGraphConf())
		if err := dgraph.Setup(dg); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
		}

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

	// Adds subdirectories command
	rootCmd.AddCommand(block.BlockCmd)
	rootCmd.AddCommand(transaction.TransactionCmd)
	rootCmd.AddCommand(cluster.ClusterCmd)
	rootCmd.AddCommand(analysis.AnalysisCmd)

	// Adds root flags and persistent flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bitgodine.yaml)")
	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", chaincfg.MainNetParams.Name, "Specify blockchain network - mainnet - testnet3 - regtest [default: mainnet]")

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	rootCmd.PersistentFlags().StringVar(&bitgodineDir, "bitgodineDir", hd, "Sets the folder containing configuration files and stored data")
	viper.SetDefault("bitgodineDir", bitgodineFolder)

	rootCmd.PersistentFlags().StringVarP(&blocksDir, "blocksDir", "b", hd, "Sets the path to the bitcoind blocks directory")
	viper.SetDefault("blocksDir", hd)

	rootCmd.PersistentFlags().StringVarP(&outputDir, "outputDir", "o", bitgodineFolder, "Sets the path to the output clusters.csv file")
	viper.SetDefault("outputDir", bitgodineFolder)

	rootCmd.PersistentFlags().StringVar(&dbDir, "dbDir", filepath.Join(bitgodineFolder, "badger"), "Sets the path to the indexing db files")
	viper.SetDefault("dbDir", filepath.Join(hd, ".bitgodine", "badger"))
	rootCmd.PersistentFlags().StringVar(&dgHost, "dgHost", "localhost", "Sets the of host the indexing graph db")
	rootCmd.PersistentFlags().IntVar(&dgPort, "dgPort", 9080, "Sets the port  the indexing db files")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.SetDefault("network", chaincfg.MainNetParams.Name)
	viper.SetDefault("dgHost", "localhost")
	viper.SetDefault("dgPort", 9080)

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("bitgodineDir", rootCmd.PersistentFlags().Lookup("bitgodineDir"))
	viper.BindPFlag("blocksDir", rootCmd.PersistentFlags().Lookup("blocksDir"))
	viper.BindPFlag("outputDir", rootCmd.PersistentFlags().Lookup("outputDir"))
	viper.BindPFlag("dbDir", rootCmd.PersistentFlags().Lookup("dbDir"))
	viper.BindPFlag("dgHost", rootCmd.PersistentFlags().Lookup("dgHost"))
	viper.BindPFlag("dgPort", rootCmd.PersistentFlags().Lookup("dgPort"))

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".bitgodine" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bitgodine")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
