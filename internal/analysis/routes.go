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
		h := vuln.ToList()
		logger.Info("Tx analysis", strings.Join(h, ","), logger.Params{})
		return c.JSON(http.StatusOK, vuln)
	})

	r.GET("/blocks", func(c echo.Context) error {
		type Query struct {
			From  int32    `query:"from" validate:"omitempty,gte=0"`
			To    int32    `query:"to" validate:"omitempty,gtefield=From"`
			List  []string `query:"heuristics" validate:"dive,oneof=locktime peeling power optimal exact type reuse shadow client forward backward"`
			Plot  string   `query:"plot" validate:"omitempty,oneof=timeline percentage"`
			Force bool     `query:"force" validate:"omitempty"`
			Type  string   `query:"type" validate:"omitempty,oneof=offbyone"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		var list []heuristics.Heuristic
		for _, h := range q.List {
			list = append(list, heuristics.Abbreviation(h))
		}
		if len(list) == 0 {
			list = heuristics.List()
		}

		if q.Type == "offbyone" {
			err := offByOneAnalysis(&c, 0, 220250, heuristics.FromListToMask(list), q.Plot)
			if err != nil {
				return err
			}
			if q.Plot != "" {
				return c.File(q.Plot + ".png")
			}
		} else {
			vuln, err := AnalyzeBlocks(&c, q.From, q.To, heuristics.FromListToMask(list), q.Force, q.Plot)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, vuln)
		}

		return c.JSON(http.StatusOK, "ok")
	})

	return r
}
