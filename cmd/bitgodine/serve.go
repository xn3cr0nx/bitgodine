package main

import (
	"os"
	"time"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/server"
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

	defer profile.Start(profile.MemProfile).Stop()

	c, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	db, err := kv.NewKV(kv.Conf(viper.GetString("badger")), c, false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	// dg := dgraph.Instance(&dgraph.Config{
	// 	Host: viper.GetString("dgHost"),
	// 	Port: viper.GetInt("dgPort"),
	// }, c)
	// if err := dg.Setup(); err != nil {
	// 	logger.Error("Bitgodine", err, logger.Params{})
	// 	logger.Error("Bitgodine", errors.New("You need to start dgraph"), logger.Params{})
	// 	os.Exit(-1)
	// }

	bdg, err := badger.NewBadger(badger.Conf("/analysis"), false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	// s := server.Instance(viper.GetInt("http.port"), dg, c, bdg)
	s := server.Instance(viper.GetInt("http.port"), db, c, bdg)
	s.Listen()
}
