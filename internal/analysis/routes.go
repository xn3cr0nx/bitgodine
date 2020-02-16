package analysis

import (
	"net/http"
	"strings"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/analysis")

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
			From int32    `query:"from" validate:"omitempty,gte=0"`
			To   int32    `query:"to" validate:"omitempty,gtfield=From"`
			Plot bool     `query:"plot" validate:"omitempty"`
			List []string `query:"heuristics" validate:"dive,oneof=locktime peeling power optimal exact type reuse shadow client forward backward"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		var list []string
		for _, h := range q.List {
			list = append(list, heuristics.Abbreviation(h))
		}

		vuln, err := AnalyzeBlocks(&c, q.From, q.To, list, q.Plot)
		if err != nil {
			return err
		}

		if q.Plot {
			return c.File("plot.png")
		}
		return c.JSON(http.StatusOK, vuln)
	})

	return r
}
