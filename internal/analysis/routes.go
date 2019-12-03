package analysis

import (
	"net/http"
	"strings"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_server/internal/plot"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/analysis")

	r.GET("/test", func(c echo.Context) error {
		plot.HeuristicsTimeline()
		return c.JSON(http.StatusOK, "ok")
	})

	r.GET("/:txid", func(c echo.Context) error {
		// TODO: check id is correct and not of a block
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required"); err != nil {
			return err
		}
		vuln, err := AnalyzeTx(&c, txid)
		if err != nil {
			return err
		}
		h := heuristics.ToList(vuln)
		logger.Info("Tx analysis", strings.Join(h, ","), logger.Params{})
		return c.JSON(http.StatusOK, vuln)
	})

	r.GET("/blocks", func(c echo.Context) error {
		type Query struct {
			From int32 `query:"from" validate:"omitempty,gte=0"`
			To   int32 `query:"to" validate:"omitempty,gtefield=From"`
			Plot bool  `query:"plot" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		vuln, err := AnalyzeBlocks(&c, q.From, q.To, 5, q.Plot)
		if err != nil {
			return err
		}

		if q.Plot {
			return c.File("barchart.png")
		}
		return c.JSON(http.StatusOK, vuln)
	})

	return r
}
