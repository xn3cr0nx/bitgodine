package analysis

import (
	"net/http"

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

		type Query struct {
			List []string `query:"heuristics" validate:"dive,oneof=locktime peeling power optimal exact type reuse shadow client forward backward"`
			Type string   `query:"type" validate:"omitempty,oneof=applicability reliability"`
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

		if q.Type == "" {
			q.Type = "applicability"
		}

		vuln, err := AnalyzeTx(&c, txid, heuristics.FromListToMask(list), q.Type)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, vuln)
	})

	r.GET("/blocks", func(c echo.Context) error {
		type Query struct {
			From     int32    `query:"from" validate:"omitempty,gte=0"`
			To       int32    `query:"to" validate:"omitempty,gtefield=From"`
			List     []string `query:"heuristics" validate:"dive,oneof=locktime peeling power optimal exact type reuse shadow client forward backward"`
			Plot     string   `query:"plot" validate:"omitempty,oneof=timeline percentage combination"`
			Force    bool     `query:"force" validate:"omitempty"`
			Analysis string   `query:"analysis" validate:"omitempty,oneof=offbyone securebasis fullmajorityvoting majorityvoting strictmajorityvoting fullmajorityanalysis reducingmajorityanalysis overlapping"`
			Type     string   `query:"type" validate:"omitempty,oneof=applicability reliability"`
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

		if q.Analysis == "offbyone" {
			if q.From > 220250 || q.To > 220250 {
				return echo.NewHTTPError(http.StatusBadRequest, "out of off by one bug range")
			}
			if q.From == 0 && q.To == 0 {
				q.To = 220250
			}
		}

		if q.Type == "" {
			q.Type = "applicability"
		}

		if err := AnalyzeBlocks(&c, q.From, q.To, heuristics.FromListToMask(list), q.Type, q.Analysis, q.Plot, q.Force); err != nil {
			return err
		}

		if q.Plot != "" {
			return c.File(q.Plot + ".png")
		}

		return c.JSON(http.StatusOK, "ok")
	})

	return r
}
