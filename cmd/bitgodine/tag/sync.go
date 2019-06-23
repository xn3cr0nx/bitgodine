package tag

import (
	"fmt"
	"path/filepath"

	"github.com/gocolly/colly"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/spider"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var target, output string

// syncCmd launch sync process of address tags
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run spiders to download and update address tags",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("tag.target") == "blockchain.com" {
			concurrent := make(chan bool)
			s1 := spider.Spider{
				colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
			}
			go s1.Sync("https://blockchain.com/btc/tags?filter=8", "submitted_links", concurrent)
			s2 := spider.Spider{
				colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
			}
			go s2.Sync("https://blockchain.com/btc/tags?filter=16", "signed_messages", concurrent)
			s3 := spider.Spider{
				colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
			}
			go s3.Sync("https://blockchain.com/btc/tags?filter=2", "bitcointalk_profiles", concurrent)
			s4 := spider.Spider{
				colly.NewCollector(colly.Async(true), colly.CacheDir(fmt.Sprintf("%s/colly", viper.GetString("bitgodineDir")))),
			}
			go s4.Sync("https://blockchain.com/btc/tags?filter=4", "bitcoinotc_profiles", concurrent)

			for i := 0; i < 4; i++ {
				<-concurrent
			}
			// s3 := spider.Spider{
			// 	colly.NewCollector(colly.Async(true)),
			// }
			// s3.Sync("https://blockchain.com/btc/tags?filter=2", "bitcointalk_profiles", concurrent)

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

	syncCmd.PersistentFlags().StringVarP(&output, "tagdir", "o", fmt.Sprintf("%s/tags", bitgodineFolder), "Specify the csv tags file output folder")
	viper.SetDefault("tag.output", fmt.Sprintf("%s/tags", bitgodineFolder))
	viper.BindPFlag("tag.output", syncCmd.PersistentFlags().Lookup("tagdir"))
}
