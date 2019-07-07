package tag

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gocolly/colly"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/postgres"
	"github.com/xn3cr0nx/bitgodine_code/internal/spider"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var target, output, dest string

// syncCmd launch sync process of address tags
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run spiders to download and update address tags",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("tag.target") == "blockchain.com" {

			pg, err := postgres.NewPg(&postgres.Config{
				Host: "localhost",
				Port: 5432,
				Pass: "bitgodine",
				User: "bitgodine",
				Name: "bitgodine",
			})
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			if err := pg.Connect(); err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			defer pg.Close()
			if err := pg.Migration(); err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			var wg sync.WaitGroup
			wg.Add(4)
			s1 := spider.Spider{
				Spider: colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
				Target: "https://blockchain.com/btc/tags?filter=8",
				Pg:     pg,
			}
			go s1.Sync("submitted_links", &wg)
			s2 := spider.Spider{
				Spider: colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
				Target: "https://blockchain.com/btc/tags?filter=16",
				Pg:     pg,
			}
			go s2.Sync("signed_messages", &wg)
			s3 := spider.Spider{
				Spider: colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
				Target: "https://blockchain.com/btc/tags?filter=2",
				Pg:     pg,
			}
			go s3.Sync("bitcointalk_profiles", &wg)
			s4 := spider.Spider{
				Spider: colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
				Target: "https://blockchain.com/btc/tags?filter=4",
				Pg:     pg,
			}
			go s4.Sync("bitcoinotc_profiles", &wg)

			wg.Wait()

			// var wg sync.WaitGroup
			// wg.Add(4)
			// s4 := spider.Spider{
			// 	Spider: colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
			// 	Target: "https://blockchain.com/btc/tags?filter=4",
			// 	Pg:     pg,
			// }
			// go s4.Sync("bitcoinotc_profiles", &wg)
			// wg.Wait()

			logger.Info("Tag sync", "Sync process ended", logger.Params{})
		}
	},
}

func init() {
	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")

	syncCmd.PersistentFlags().StringVar(&target, "target", "blockchain.com", "Specify the spider's target")
	viper.SetDefault("tag.target", "blockchain.com")
	viper.BindPFlag("tag.target", syncCmd.PersistentFlags().Lookup("target"))

	syncCmd.PersistentFlags().StringVar(&dest, "dest", "db", "Specify the crawled data output destination | db or csv")
	viper.SetDefault("tag.dest", "db")
	viper.BindPFlag("tag.dest", syncCmd.PersistentFlags().Lookup("dest"))

	syncCmd.PersistentFlags().StringVarP(&output, "tagdir", "o", fmt.Sprintf("%s/tags", bitgodineFolder), "Specify the csv tags file output folder")
	viper.SetDefault("tag.output", fmt.Sprintf("%s/tags", bitgodineFolder))
	viper.BindPFlag("tag.output", syncCmd.PersistentFlags().Lookup("tagdir"))
}
