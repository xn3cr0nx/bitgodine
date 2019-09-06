package main

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_server/internal/server"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
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

	s := server.Instance(viper.GetInt("http.port"))
	s.Listen()
}
