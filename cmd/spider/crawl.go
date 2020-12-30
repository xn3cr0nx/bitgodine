package main

import (
	"fmt"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/spider/bitcoinabuse"
	"github.com/xn3cr0nx/bitgodine/internal/spider/checkbitcoinaddress"
	"github.com/xn3cr0nx/bitgodine/internal/spider/walletexplorer"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/migration"
)

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl resources as cronjob",
	Long: `Dispatch a series of spiders to gracefully crawl the set of specified
resources, and sync the library of address tags with new reports.`,
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
