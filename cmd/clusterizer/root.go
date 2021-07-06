package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/clusterizer/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/disjoint/disk"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var (
	db, boltFile, output string
	debug, realtime      bool
)

var rootCmd = &cobra.Command{
	Use:   "clusterizer",
	Short: "Creates clusters from synced blocks",
	Long: `Fetch block by block and creates a persistent
	version of the cluster of addresses. The cluster is stored
	in a persistent way in storage layer.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Start", "Start called", logger.Params{})

		c, err := cache.NewCache(nil)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		db, err := kv.NewDB()
		defer db.Close()

		set, err := disk.NewDisjointSet(db, true, true)
		if err != nil {
			logger.Error("Start", err, logger.Params{})
			os.Exit(-1)
		}
		if err := disk.RestorePersistentSet(&set); err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
				logger.Error("Start", err, logger.Params{})
				os.Exit(-1)
			}
		}

		pg, err := postgres.NewPg(postgres.Conf())
		if err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
			os.Exit(-1)
		}

		interrupt := make(chan int)
		done := make(chan int)
		bc := bitcoin.NewClusterizer(&set, db, pg, c, interrupt, done)
		bc.Clusterize()

		cltzCount, err := bc.Done()
		if err != nil {
			logger.Error("Clusterizer", err, logger.Params{})
			os.Exit(-1)
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)
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

	if value, ok := os.LookupEnv("config"); ok {
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
