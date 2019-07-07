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

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store clusters to the database",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
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

		logger.Info("Cluster store", "storing clusters...", logger.Params{})
		set := persistent.NewDisjointSet(dgraph.Instance(nil))
		if err := persistent.RestorePersistentSet(&set); err != nil {
			if err.Error() != "Cluster not found" {
				logger.Error("Cluster store", err, logger.Params{})
				os.Exit(-1)
			}
		}
		cltz := visitor.NewClusterizer(&set)
		_, err = cltz.Done()
		if err != nil {
			logger.Error("Blockchain test", err, logger.Params{})
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

	},
}

// func init() {
// }
