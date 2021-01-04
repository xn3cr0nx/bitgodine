package main

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var (
	bitgodineDir, blocksDir, dbDir, bdg, analysis string
	dgPort                                        int
	debug                                         bool
)

var rootCmd = &cobra.Command{
	Use:   "bitgodine",
	Short: "Go implementation of Bitiodine",
	Long: `Go implementation of Bitcoin forensic analysis tool to	investigate blockchain and Bitcoin malicious flows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()
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

	rootCmd.AddCommand(serveCmd)

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().StringVar(&bitgodineDir, "bitgodineDir", bitgodineFolder, "Sets the folder containing configuration files and stored data")
	rootCmd.PersistentFlags().StringVarP(&blocksDir, "blocksDir", "b", hd, "Sets the path to the bitcoind blocks directory")
	rootCmd.PersistentFlags().StringVar(&dbDir, "dbDir", filepath.Join(bitgodineFolder, "badger"), "Sets the path to the indexing db files")
	rootCmd.PersistentFlags().StringVar(&bdg, "badger", "/badger", "Sets the path to the badger stored files")
	rootCmd.PersistentFlags().StringVar(&analysis, "analysis", "/analysis", "Sets the path to the analysis stored files")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv()

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
