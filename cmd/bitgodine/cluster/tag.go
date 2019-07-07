package cluster

import (
	"errors"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/addresses"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/postgres"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var full bool

// tagCmd represents the store command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Shows tagged clusters available in the database",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			logger.Error("Cluster tag", errors.New("Missing parameter"), logger.Params{})
			os.Exit(-1)
		}

		pg, err := postgres.NewPg(&postgres.Config{
			Host: "localhost",
			Port: 5432,
			Pass: "bitgodine",
			User: "bitgodine",
			Name: "bitgodine",
		})
		if err != nil {
			logger.Error("Cluster tag", err, logger.Params{})
			os.Exit(-1)
		}
		if err := pg.Connect(); err != nil {
			logger.Error("Cluster tag", err, logger.Params{})
			os.Exit(-1)
		}
		defer pg.Close()
		if err := pg.Migration(); err != nil {
			logger.Error("Cluster tag", err, logger.Params{})
			os.Exit(-1)
		}

		type TaggedCluster struct {
			Cluster  int    `gorm:"not null"`
			Address  string `gorm:"primary_key;size:64;index;not null;unique"`
			Tag      string `gorm:"index;not null"`
			Verified bool   `gorm:"default:false"`
		}
		var clusters []TaggedCluster

		address := addresses.IsBitcoinAddress(args[0])
		if address {
			if err := pg.DB.Raw(`select t.address, cluster, tag, verified from tags right join 
			(select address, cluster from clusters where cluster = (select cluster from clusters where address = ?)) t 
			on tags.address = t.address;`, args[0]).Find(&clusters).Error; err != nil {
				logger.Error("Cluster tag", err, logger.Params{})
				os.Exit(-1)
			}
		} else {
			if err := pg.DB.Raw(`select c.address, c.cluster, t.tag, t.verified from clusters c left join tags t on t.address = c.address 
					where cluster = (select cluster from clusters where address = (select address from tags where tag = ?));`, args[0]).Find(&clusters).Error; err != nil {
				logger.Error("Cluster tag", err, logger.Params{})
				os.Exit(-1)
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Address", "Tag", "Cluster", "Verified"})
		red := color.New(color.FgRed)
		green := color.New(color.FgGreen)

		if viper.GetBool("tag.full") || !address {
			for _, cluster := range clusters {
				value := strconv.Itoa(cluster.Cluster)
				if cluster.Verified == true {
					table.Append([]string{cluster.Address, cluster.Tag, value, green.Sprint("✓")})
				} else {
					table.Append([]string{cluster.Address, cluster.Tag, value, "-"})
				}
			}
		} else {
			for _, cluster := range clusters {
				if cluster.Tag != "" {
					value := strconv.Itoa(cluster.Cluster)
					if cluster.Verified == true {
						table.Append([]string{cluster.Address, cluster.Tag, value, green.Sprint("✓")})
					} else {
						table.Append([]string{cluster.Address, cluster.Tag, value, red.Sprint("x")})
					}
				}
			}
		}

		table.Render()
	},
}

func init() {
	tagCmd.PersistentFlags().BoolVar(&full, "full", false, "Specify whether show the entire list of address in the cluster")
	viper.SetDefault("tag.full", "")
	viper.BindPFlag("tag.full", tagCmd.PersistentFlags().Lookup("full"))
}
