package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/migration"
	"github.com/xn3cr0nx/bitgodine/internal/spider/bitcoinabuse"
	"github.com/xn3cr0nx/bitgodine/internal/spider/checkbitcoinaddress"
	"github.com/xn3cr0nx/bitgodine/internal/spider/walletexplorer"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var (
	debug, cr bool
	target    string
)

var rootCmd = &cobra.Command{
	Use:   "spider",
	Short: "Spider service to sync addresses tags resources",
	Long:  `Spider service crawling many web resources to sync and update addresses tags storage`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Spider", "Spider crawling...", logger.Params{})

		if err := configCheck(); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		pg, err := postgres.NewPg(postgres.Conf())
		if err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}
		if err := migration.Migration(pg); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		target := viper.GetString("target")
		if target != "" {
			if target == "bitcoinabuse" {
				btcabuse := bitcoinabuse.NewSpider(pg)
				if err := btcabuse.Sync(); err != nil {
					logger.Error("Spider", err, logger.Params{})
					os.Exit(-1)
				}
				logger.Info("Spider", "bitcoinabuse sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com"})
				return
			}

			if target == "checkbitcoinaddress" {
				checkbtcaddr := checkbitcoinaddress.NewSpider(pg)
				if err := checkbtcaddr.Sync(); err != nil {
					logger.Error("Spider", err, logger.Params{})
					os.Exit(-1)
				}
				logger.Info("Spider", "checkbitcoinaddress sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com"})
				return
			}

			if target == "walletexplorer" {
				wexplorer := walletexplorer.NewSpider(pg)
				if err := wexplorer.Sync(); err != nil {
					logger.Error("Spider", err, logger.Params{})
					os.Exit(-1)
				}
				logger.Info("Spider", "walletexplorer sync ended, waiting for next schedule", logger.Params{"target": "walletexplorer.com"})
				return
			}
		}

		if !viper.GetBool("cron") {
			btcabuse := bitcoinabuse.NewSpider(pg)
			if err := btcabuse.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "bitcoinabuse sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com", "crontime": viper.GetString("spider.crontime")})

			checkbtcaddr := checkbitcoinaddress.NewSpider(pg)
			if err := checkbtcaddr.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "checkbitcoinaddress sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com", "crontime": viper.GetString("spider.crontime")})

			wexplorer := walletexplorer.NewSpider(pg)
			if err := wexplorer.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "walletexplorer sync ended, waiting for next schedule", logger.Params{"target": "walletexplorer.com", "crontime": viper.GetString("spider.crontime")})
			return
		}

		c := cron.New()
		defer c.Stop()

		logger.Info("Spider", "Scheduling spider crawler", logger.Params{"target": "bitcoinabuse.com", "crontime": viper.GetString("spider.crontime")})
		if _, err = c.AddFunc(viper.GetString("spider.crontime"), func() {
			btcabuse := bitcoinabuse.NewSpider(pg)
			if err := btcabuse.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "bitcoinabuse sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com", "crontime": viper.GetString("spider.crontime")})
		}); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Spider", "Scheduling spider crawler", logger.Params{"target": "checkbitcoinaddress.com", "crontime": viper.GetString("spider.crontime")})
		if _, err = c.AddFunc(viper.GetString("spider.crontime"), func() {
			checkbtcaddr := checkbitcoinaddress.NewSpider(pg)
			if err := checkbtcaddr.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "checkbitcoinaddress sync ended, waiting for next schedule", logger.Params{"target": "bitcoinabuse.com", "crontime": viper.GetString("spider.crontime")})
		}); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Spider", "Scheduling spider crawler", logger.Params{"target": "walletexplorer.com", "crontime": viper.GetString("spider.crontime")})
		if _, err = c.AddFunc(viper.GetString("spider.crontime"), func() {
			wexplorer := walletexplorer.NewSpider(pg)
			if err := wexplorer.Sync(); err != nil {
				logger.Error("Spider", err, logger.Params{})
				os.Exit(-1)
			}
			logger.Info("Spider", "walletexplorer sync ended, waiting for next schedule", logger.Params{"target": "walletexplorer.com", "crontime": viper.GetString("spider.crontime")})
		}); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}

		c.Run()
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
	rootCmd.PersistentFlags().BoolVar(&cr, "cron", true, "Sets if spider should be started as cron or just run once")
	rootCmd.PersistentFlags().StringVar(&target, "target", "", "Sets if spider should run once with a specific target")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("debug", false)
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetDefault("cron", true)
	viper.BindPFlag("cron", rootCmd.PersistentFlags().Lookup("cron"))

	viper.SetDefault("target", "")
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))

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

func configCheck() error {
	if viper.GetString("spider.bitcoinabuse.url") == "" {
		return fmt.Errorf("%w: missing bitcoinabuse endpoint", errorx.ErrInvalidArgument)
	}
	if viper.GetString("spider.bitcoinabuse.api") == "" {
		return fmt.Errorf("%w: missing bitcoinabuse api key", errorx.ErrInvalidArgument)
	}
	if viper.GetString("spider.checkbitcoinaddress.url") == "" {
		return fmt.Errorf("%w: missing checkbitcoinaddress endpoint", errorx.ErrInvalidArgument)
	}

	return nil
}
