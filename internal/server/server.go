package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xn3cr0nx/bitgodine/internal/abuse"
	"github.com/xn3cr0nx/bitgodine/internal/address"
	"github.com/xn3cr0nx/bitgodine/internal/analysis"
	"github.com/xn3cr0nx/bitgodine/internal/auth"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/cluster"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"github.com/xn3cr0nx/bitgodine/internal/trace"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/meter"
	"github.com/xn3cr0nx/bitgodine/pkg/pprof"
	"github.com/xn3cr0nx/bitgodine/pkg/tracer"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
	"go.opentelemetry.io/otel/label"

	v "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	otleTrace "go.opentelemetry.io/otel/trace"

	echoSwagger "github.com/swaggo/echo-swagger"
)

// Server struct initialized with port
type (
	Server struct {
		port   string
		router *echo.Echo
		db     kv.DB
		cache  *cache.Cache
		pg     *postgres.Pg
	}
)

var server *Server

// NewServer singleton pattern that returns pointer to server
func NewServer(port int, db kv.DB, c *cache.Cache, pg *postgres.Pg) *Server {
	if server != nil {
		return server
	}
	server = &Server{
		port:   fmt.Sprintf(":%d", port),
		router: echo.New(),
		db:     db,
		cache:  c,
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

	_, err := meter.NewMeter(&meter.Config{Name: "bitgodine_api"})
	if err != nil {
		panic(errors.Wrapf(err, "cannot setup meter"))
	}

	tracer, tracerMiddleware, err := tracer.NewTracer(&tracer.Config{Name: "bitgodine_api", Exporter: tracer.Jaeger})
	if err != nil {
		panic(errors.Wrapf(err, "cannot setup tracing"))
	}
	s.router.Use(tracerMiddleware)

	s.router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", s.db)
			c.Set("cache", s.cache)
			c.Set("pg", s.pg)
			c.Set("tracer", tracer)
			return next(c)
		}
	})

	s.router.Use(middleware.Recover())
	s.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "swagger")
		},
	}))

	s.router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: viper.GetStringSlice(("auth.cors")),
		AllowMethods: viper.GetStringSlice(("auth.methods")),
	}))

	s.router.Use(middleware.RequestID())

	s.router.GET("/", func(c echo.Context) error {
		_, span := (*tracer).Start(context.Background(), "status", otleTrace.WithAttributes(label.String("id", c.Request().Header.Get("X-Request-ID"))))
		defer span.End()
		return c.JSON(http.StatusOK, "Welcome to Bitgodine REST API")
	})

	s.router.GET("/swagger/*", echoSwagger.WrapHandler)

	api := s.router.Group("/api")

	abuseService := abuse.NewService(s.pg, s.cache)
	abuse.Routes(api, abuseService)
	addressService := address.NewService(s.db, s.cache)
	address.Routes(api, addressService)
	analysisService := analysis.NewService(s.db, s.cache)
	analysis.Routes(api, analysisService)
	authService := auth.NewService(s.pg)
	auth.Routes(api, authService)
	blockService := block.NewService(s.db, s.cache)
	block.Routes(api, blockService)
	clusterService := cluster.NewService(s.pg, s.cache)
	cluster.Routes(api, clusterService)
	tagService := tag.NewService(s.pg, s.cache)
	tag.Routes(api, tagService)
	traceService := trace.NewService(s.pg, s.db, s.cache)
	trace.Routes(api, traceService)
	txService := tx.NewService(s.db, s.cache)
	tx.Routes(api, txService)

	// fmt.Println("ROUTES:")
	// for _, route := range s.router.Routes() {
	// 	fmt.Println(route.Method, route.Path)
	// }

	go func() {
		if err := s.router.Start(s.port); err != nil {
			s.router.Logger.Error(err)
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
		} else {
			if stringError, ok := e.Message.(string); ok {
				m = stringError
			} else {
				m = err.Error()
			}
		}
	} else {
		if errors.Is(err, errorx.ErrKeyNotFound) {
			code = http.StatusNotFound
		}
	}

	message := map[string]interface{}{"code": code, "error": http.StatusText(code)}
	if m != "" && m != message["error"] {
		message["type"] = m
	}
	c.JSON(code, message)
}
