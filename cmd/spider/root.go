package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
)

var (
	debug, cr bool
)

var rootCmd = &cobra.Command{
	Use:   "spider",
	Short: "Spider service to sync addresses tags resources",
	Long:  `Spider service crawling many web resources to sync and update addresses tags storage`,
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
	rootCmd.AddCommand(crawlCmd)

	// Adds root flags and persistent flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().BoolVar(&cr, "cron", true, "Sets if spider should be started as cron or just run once")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetDefault("cron", true)
	viper.BindPFlag("cron", rootCmd.PersistentFlags().Lookup("cron"))

	viper.SetEnvPrefix("spider")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if value, ok := os.LookupEnv("CONFIG_FILE"); ok {
		viper.SetConfigFile(value)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/spider/")
		viper.AddConfigPath("$HOME/.bitgodine/spider")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.ReadInConfig()
	f := viper.ConfigFileUsed()
	if f != "" {
		fmt.Printf("Found configuration file: %s \n", f)
	}

}
