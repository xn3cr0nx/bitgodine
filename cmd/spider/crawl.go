package main

import (
	"errors"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/spider/bitcoinabuse"
	"github.com/xn3cr0nx/bitgodine_server/internal/spider/checkbitcoinaddress"
	"github.com/xn3cr0nx/bitgodine_server/pkg/migration"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
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
		if err := pg.Connect(); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
		}
		defer pg.Close()
		if err := migration.Migration(pg); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(-1)
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

		c.Run()
	},
}

func configCheck() error {
	if viper.GetString("spider.bitcoinabuse.url") == "" {
		return errors.New("missing bitcoinabuse endpoint")
	}
	if viper.GetString("spider.bitcoinabuse.api") == "" {
		return errors.New("missing bitcoinabuse api key")
	}
	if viper.GetString("spider.checkbitcoinaddress.url") == "" {
		return errors.New("missing checkbitcoinaddress endpoint")
	}

	return nil
}
