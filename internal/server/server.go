package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/internal/address"
	"github.com/xn3cr0nx/bitgodine_server/internal/analysis"
	"github.com/xn3cr0nx/bitgodine_server/internal/block"
	chttp "github.com/xn3cr0nx/bitgodine_server/internal/http"
	"github.com/xn3cr0nx/bitgodine_server/internal/tag"
	"github.com/xn3cr0nx/bitgodine_server/internal/trace"
	"github.com/xn3cr0nx/bitgodine_server/internal/tx"
	"github.com/xn3cr0nx/bitgodine_server/pkg/pprof"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"
	"github.com/xn3cr0nx/bitgodine_spider/pkg/postgres"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	v "gopkg.in/go-playground/validator.v9"
)

// Server struct initialized with port
type (
	Server struct {
		port   string
		router *echo.Echo
		db     storage.DB
		cache  *cache.Cache
		kv     *badger.Badger
		pg     *postgres.Pg
	}
)

var server *Server

// Instance singleton pattern that returnes pointer to server
func Instance(port int, db storage.DB, c *cache.Cache, bdg *badger.Badger, pg *postgres.Pg) *Server {
	if server != nil {
		return server
	}
	server = &Server{
		port:   fmt.Sprintf(":%d", port),
		router: echo.New(),
		db:     db,
		cache:  c,
		kv:     bdg,
		pg:     pg,
	}
	return server
}

// Listen initializes the echo webserver
func (s *Server) Listen() {
	pprof.Wrap(s.router)

	s.router.HideBanner = true
	s.router.Debug = viper.GetBool("debug")
	s.router.Use(middleware.Logger())
	s.router.Logger.SetLevel(log.INFO)
	s.router.Validator = validator.NewValidator()

	s.router.HTTPErrorHandler = customHTTPErrorHandler

	s.router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", s.db)
			c.Set("cache", s.cache)
			c.Set("kv", s.kv)
			c.Set("pg", s.pg)
			return next(c)
		}
	})

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
		return c.JSON(http.StatusOK, "Welcome to Bitgodine REST API")
	})

	api := s.router.Group("/api")
	tx.Routes(api)
	block.Routes(api)
	address.Routes(api)
	analysis.Routes(api)
	trace.Routes(api)
	tag.Routes(api)

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
	m := ""
	if e, ok := err.(*echo.HTTPError); ok {
		code = e.Code
		if httpError, ok := e.Message.(*echo.HTTPError); ok {
			m = httpError.Message.(string)
		} else if _, ok := e.Message.(v.ValidationErrors); ok {
			m = "Bad Request"
		} else {
			if customError, ok := e.Message.(chttp.Error); ok {
				m = customError.Type
			} else if stringError, ok := e.Message.(string); ok {
				m = stringError
			} else {
				// TODO: manipulate string and extract just message
				m = err.Error()
			}
		}
	}

	message := map[string]interface{}{"code": code, "error": http.StatusText(code)}
	if m != "" && m != message["error"] {
		message["type"] = m
	}
	c.JSON(code, message)
}
