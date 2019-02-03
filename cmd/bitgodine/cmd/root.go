package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var (
	cfgFile, network, blocksDir, outputDir string
	debug                                  bool
	BitcoinNet                             chaincfg.Params
)

var rootCmd = &cobra.Command{
	Use:   "bitgodine",
	Short: "Go implementation of Bitiodine",
	Long: `Go implementation of Bitcoin forensic analysis tool to	investigate blockchain and Bitcoin malicious flows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bitgodine.yaml)")

	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", chaincfg.MainNetParams.Name, "Specify blockchain network - mainnet - testnet3 - regtest [default: mainnet]")

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	rootCmd.PersistentFlags().StringVarP(&blocksDir, "blocksDir", "b", hd, "Sets the path to the bitcoind blocks directory")
	viper.SetDefault("blocksDir", hd)

	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	rootCmd.PersistentFlags().StringVarP(&outputDir, "outputDir", "o", wd, "Sets the path to the output clusters.csv file")
	viper.SetDefault("outputDir", wd)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.SetDefault("network", chaincfg.MainNetParams.Name)

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("blocksDir", rootCmd.PersistentFlags().Lookup("blocksDir"))
	viper.BindPFlag("outputDir", rootCmd.PersistentFlags().Lookup("outputDir"))

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
