package main

import (
	"os"
	"time"

	// "github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/server"
	badgerStorage "github.com/xn3cr0nx/bitgodine/pkg/badger/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/migration"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
	redisStorage "github.com/xn3cr0nx/bitgodine/pkg/redis/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"
	tikvStorage "github.com/xn3cr0nx/bitgodine/pkg/tikv/storage"

	"github.com/xn3cr0nx/bitgodine/pkg/badger"
	"github.com/xn3cr0nx/bitgodine/pkg/redis"
	"github.com/xn3cr0nx/bitgodine/pkg/tikv"
)

var (
	port       int
	mode, host string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve bitgodine web server",
	Long: `Serve web server instance
exposing router to retrieve stored data about blocks and transactions.
The server is bu default exposed on port 3000.`,
	Run: start,
}

func init() {
	serveCmd.Flags().IntVar(&port, "port", 3000, "bind http server to port")
	serveCmd.Flags().StringVar(&host, "host", "localhost", "bind http server to host")
	serveCmd.Flags().StringVar(&mode, "mode", "debug", "http server mode (release for production)")

	viper.BindPFlag("http.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("http.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("http.mode", serveCmd.Flags().Lookup("mode"))
}

func start(cmd *cobra.Command, args []string) {
	logger.Info("Bitgodine Serve", "Server Starting", logger.Params{"timestamp": time.Now()})

	// defer profile.Start(profile.MemProfile, profile.ProfilePath("./mem.pprof")).Stop()

	c, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	var db storage.DB
	var kvdb kv.KV
	if viper.GetString("db") == "tikv" {
		db, err = tikvStorage.NewKV(tikvStorage.Conf(viper.GetString("tikv")), c)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer db.Close()

		kvdb, err = tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer kvdb.Close()

	} else if viper.GetString("db") == "badger" {
		db, err = badgerStorage.NewKV(badgerStorage.Conf(viper.GetString("badger")), c, false)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer db.Close()

		kvdb, err = badger.NewBadger(badger.Conf(viper.GetString("clusterizer.disjoint")), false)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer kvdb.Close()
	} else if viper.GetString("db") == "redis" {
		db, err = redisStorage.NewKV(redisStorage.Conf(viper.GetString("redis")), c)
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer db.Close()

		kvdb, err = redis.NewRedis(redis.Conf(viper.GetString("redis")))
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer kvdb.Close()
	}

	pg, err := postgres.NewPg(postgres.Conf())
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}
	if err := pg.Connect(); err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}
	defer pg.Close()
	if err := migration.Migration(pg); err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	s := server.Instance(viper.GetInt("server.http.port"), db, c, kvdb, pg)
	s.Listen()
}
