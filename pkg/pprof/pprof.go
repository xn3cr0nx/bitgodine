package pprof

import (
	"net/http/pprof"
	"strings"

	"github.com/labstack/echo/v4"
)

// Wrap adds several routes from package `net/http/pprof` to *echo.Echo object.
func Wrap(e *echo.Echo) {
	WrapGroup("", e.Group(""))
}

// Wrapper make sure we are backward compatible.
var Wrapper = Wrap

// WrapGroup adds several routes from package `net/http/pprof` to *echo.Group object.
func WrapGroup(prefix string, g *echo.Group) {
	routers := []struct {
		Method  string
		Path    string
		Handler echo.HandlerFunc
	}{
		{"GET", "/debug/pprof/", Handler("index")},
		{"GET", "/debug/pprof", Handler("index")},
		{"GET", "/debug/pprof/heap", Handler("heap")},
		{"GET", "/debug/pprof/goroutine", Handler("goroutine")},
		{"GET", "/debug/pprof/block", Handler("block")},
		{"GET", "/debug/pprof/threadcreate", Handler("threadcreate")},
		{"GET", "/debug/pprof/cmdline", Handler("cmdline")},
		{"GET", "/debug/pprof/profile", Handler("profile")},
		{"GET", "/debug/pprof/symbol", Handler("symbol")},
		{"POST", "/debug/pprof/symbol", Handler("symbol")},
		{"GET", "/debug/pprof/trace", Handler("trace")},
		{"GET", "/debug/pprof/mutex", Handler("mutex")},
	}

	for _, r := range routers {
		switch r.Method {
		case "GET":
			g.GET(strings.TrimPrefix(r.Path, prefix), r.Handler)
		case "POST":
			g.POST(strings.TrimPrefix(r.Path, prefix), r.Handler)
		}
	}
}

// Handler dispatch the correct profiler based on passed argument
func Handler(handler string) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		switch {
		case handler == "index":
			pprof.Index(ctx.Response(), ctx.Request())
		case handler == "profile":
			pprof.Profile(ctx.Response(), ctx.Request())
		case handler == "cmdline":
			pprof.Cmdline(ctx.Response(), ctx.Request())
		case handler == "symbol":
			pprof.Symbol(ctx.Response(), ctx.Request())
		case handler == "trace":
			pprof.Trace(ctx.Response(), ctx.Request())
		default:
			pprof.Handler(handler).ServeHTTP(ctx.Response(), ctx.Request())
		}
		return nil
	}
}
