package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/migration"
	"github.com/xn3cr0nx/bitgodine/internal/server"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var (
	host, bitgodineDir, blocksDir, dbDir, bdg, analysis string
	port, dgPort                                        int
	debug                                               bool
)

var rootCmd = &cobra.Command{
	Use:   "bitgodine",
	Short: "Serve bitgodine web server",
	Long: `Serve web server instance
	exposing router to retrieve stored data about blocks and transactions.
	The server is bu default exposed on port 3000.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Setup()
	},
	Run: serve,
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

	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Sets logging level to Debug")
	rootCmd.PersistentFlags().StringVar(&bitgodineDir, "bitgodineDir", bitgodineFolder, "Sets the folder containing configuration files and stored data")
	rootCmd.PersistentFlags().StringVarP(&blocksDir, "blocksDir", "b", hd, "Sets the path to the bitcoind blocks directory")
	rootCmd.PersistentFlags().StringVar(&dbDir, "dbDir", filepath.Join(bitgodineFolder, "badger"), "Sets the path to the indexing db files")
	rootCmd.PersistentFlags().StringVar(&bdg, "badger", "/badger", "Sets the path to the badger stored files")
	rootCmd.PersistentFlags().StringVar(&analysis, "analysis", "/analysis", "Sets the path to the analysis stored files")
	rootCmd.PersistentFlags().IntVar(&port, "port", 3000, "bind http server to port")
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "bind http server to host")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.BindPFlag("http.port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("http.host", rootCmd.Flags().Lookup("host"))

	viper.AutomaticEnv()

	if value, ok := os.LookupEnv("config"); ok {
		viper.SetConfigFile(value)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/server/")
		viper.AddConfigPath("$HOME/.bitgodine/server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.ReadInConfig()
	f := viper.ConfigFileUsed()
	if f != "" {
		fmt.Printf("Found configuration file: %s \n", f)
	}
}

func serve(cmd *cobra.Command, args []string) {
	logger.Info("Bitgodine Serve", "Server Starting", logger.Params{"timestamp": time.Now()})

	// defer profile.Start(profile.MemProfile, profile.ProfilePath("./mem.pprof")).Stop()

	c, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	db, err := kv.NewDB()
	defer db.Close()

	pg, err := postgres.NewPg(postgres.Conf())
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}
	if err := migration.Migration(pg); err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	s := server.NewServer(viper.GetInt("server.http.port"), db, c, pg)
	s.Listen()
}
