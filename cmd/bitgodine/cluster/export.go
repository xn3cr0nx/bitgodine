package cluster

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/postgres"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/persistent"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var target string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export clusters to the specified output",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Cluster Export", "exporting clusters...", logger.Params{})
		set := persistent.NewDisjointSet(dgraph.Instance(nil))
		if err := persistent.RestorePersistentSet(&set); err != nil {
			if err.Error() != "Cluster not found" {
				logger.Error("Cluster export", err, logger.Params{})
				os.Exit(-1)
			}
		}
		cltz := visitor.NewClusterizer(&set)
		if _, err := cltz.Done(); err != nil {
			logger.Error("Cluster export", err, logger.Params{})
			os.Exit(-1)
		}

		if viper.GetString("cluster.target") == "db" {
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
				logger.Error("Cluster store", err, logger.Params{})
				os.Exit(-1)
			}

			file, _ := os.Open(fmt.Sprintf("%s/clusters.csv", viper.GetString("csv.output")))
			defer file.Close()
			reader := csv.NewReader(bufio.NewReader(file))
			for {
				line, error := reader.Read()
				if error == io.EOF {
					break
				} else if error != nil {
					logger.Error("Cluster store", err, logger.Params{})
					os.Exit(-1)
				}
				value, err := strconv.Atoi(line[1])
				if err != nil {
					logger.Error("Cluster store", err, logger.Params{})
					os.Exit(-1)
				}
				cluster := postgres.Cluster{
					Cluster: value,
					Address: line[0],
				}
				if pg.DB.NewRecord(&cluster) {
					if res := pg.DB.Create(&cluster); res.Error != nil {
						logger.Error("Cluster store", res.Error, logger.Params{})
						os.Exit(-1)
					}
				}
			}
		}
	},
}

func init() {
	exportCmd.PersistentFlags().StringVar(&target, "target", "csv", "Specify the target of exporting cluster [csv, db]")
	viper.SetDefault("cluster.target", "csv")
	viper.BindPFlag("cluster.target", exportCmd.PersistentFlags().Lookup("target"))
}
