package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var (
	db, boltFile, output string
	debug, realtime      bool
)

var rootCmd = &cobra.Command{
	Use:   "clusterizer",
	Short: "Clusterizer service",
	Long: `Clusterizer service in bitgodine architecture in charge of keeping
cluster of addresses up to date based on chain stored in local storage.`,
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

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(exportCmd)

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", bitgodineFolder, "Sets the path to output clusters.csv file")
	rootCmd.PersistentFlags().StringVar(&db, "db", "/badger", "Sets the path to the storage stored files")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetEnvPrefix("clusterizer")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if value, ok := os.LookupEnv("CONFIG_FILE"); ok {
		viper.SetConfigFile(value)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/clusterizer/")
		viper.AddConfigPath("$HOME/.bitgodine/clusterizer")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.ReadInConfig()
	f := viper.ConfigFileUsed()
	if f != "" {
		fmt.Printf("Found configuration file: %s \n", f)
	}
}
