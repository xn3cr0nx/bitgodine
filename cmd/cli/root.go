package main

import (
	"fmt"
	"os"

	"github.com/xn3cr0nx/bitgodine/cmd/cli/address"
	"github.com/xn3cr0nx/bitgodine/cmd/cli/block"
	"github.com/xn3cr0nx/bitgodine/cmd/cli/transaction"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	_ "net/http/pprof"
)

var (
	host  string
	debug bool
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "Cli implementation of Bitgodine",
	Long: `Cli implementation of Bitcoin forensic analysis tool to	investigate blockchain and Bitcoin malicious flows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bitgodine host: %s\n", viper.GetString("host"))
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
	rootCmd.AddCommand(address.AddressCmd)

	// Adds root flags and persistent flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "Specify bitgodine host - [default: localhost]")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.SetDefault("host", "localhost")

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))

	viper.AutomaticEnv()
}
