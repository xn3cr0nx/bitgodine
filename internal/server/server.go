package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/routes/address"
	"github.com/xn3cr0nx/bitgodine_code/internal/routes/block"
	"github.com/xn3cr0nx/bitgodine_code/internal/routes/tx"
	"github.com/xn3cr0nx/bitgodine_code/pkg/validator"

	"github.com/dgraph-io/dgo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Server struct initialized with port
type (
	Server struct {
		port   string
		router *echo.Echo
		db     *dgo.Dgraph
	}
)

var server *Server

// Instance singleton pattern that returnes pointer to server
func Instance(port int) *Server {
	if server != nil {
		return server
	}
	dg := dgraph.Instance(&dgraph.Config{
		Host: viper.GetString("dgHost"),
		Port: viper.GetInt("dgPort"),
	})
	if err := dgraph.Setup(dg); err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		logger.Error("Bitgodine", errors.New("You need to start dgraph"), logger.Params{})
		os.Exit(-1)
	}
	server = &Server{
		port:   fmt.Sprintf(":%d", port),
		router: echo.New(),
		db:     dg,
	}
	return server
}

// Listen initializes the echo webserver
func (s *Server) Listen() {
	s.router.HideBanner = true
	s.router.Debug = viper.GetBool("debug")
	s.router.Use(middleware.Logger())
	s.router.Logger.SetLevel(log.INFO)
	s.router.Validator = validator.NewValidator()

	s.router.HTTPErrorHandler = customHTTPErrorHandler

	s.router.Use(middleware.Recover())
	s.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	s.router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:4200"},
		AllowMethods: []string{"*"},
	}))

	s.router.Use(middleware.RequestID())

	s.router.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Welcome to Bitgodine Rest API")
	})

	api := s.router.Group("/api")
	tx.Routes(api)
	block.Routes(api)
	address.Routes(api)

	fmt.Println("ROUTES:")
	for _, route := range s.router.Routes() {
		fmt.Println(route.Method, route.Path)
	}

	go func() {
		if err := s.router.Start(s.port); err != nil {
			s.router.Logger.Fatal("shutting down the server")
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.router.Logger.Info("signal caught. gracefully shutting down...")
	if err := s.router.Shutdown(ctx); err != nil {
		s.router.Logger.Fatal(err)
	}
}

func customHTTPErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	message := map[string]interface{}{"code": code, "error": http.StatusText(code)}
	c.JSON(code, message)
}
