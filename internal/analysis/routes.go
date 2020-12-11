package analysis

import (
	"net/http"

	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts /analysis based routes on the main group
func Routes(g *echo.Group) {
	r := g.Group("/analysis")
	r.GET("/:txid", analysisID)
	r.GET("/blocks", analysisBlocks)
}

// analysisID godoc
// @ID analysis-id
//
// @Router /analysis/{txid} [get]
// @Summary Analysis by id
// @Description get analysis for transaction by id
// @Tags analysis
//
// @Accept  json
// @Produce  json
//
// @Param txid path string true "Transaction ID"
// @Param heuristics query []string false "Heuristics list" Enums(locktime, peeling, power, optimal, exact, type, reuse, shadow, client, forward, backward)
// @Param type query string false "Analysis type" Enums(applicability, reliability)
//
// @Success 200 {object} object
// @Success 500 {string} string
func analysisID(c echo.Context) error {
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
}

// analysisBlocks godoc
// @ID analysis-blocks
//
// @Router /analysis/blocks [get]
// @Summary Analysis blocks
// @Description get analysis for a block range
// @Tags analysis
//
// @Accept  json
// @Produce  json
//
// @Param from query int false "From block" minimum(0)
// @Param to query int false "To block"
// @Param heuristics query []string false "Heuristics" Enums(locktime, peeling, power, optimal, exact, type, reuse, shadow, client, forward, backward)
// @Param plot query string false "Plot type" Enums(timeline, percentage, combination)
// @Param force query bool false "Rewrite previous stored results"
// @Param analysis query string false "Analysis output" Enums(offbyone, securebasis, fullmajorityvoting, majorityvoting, strictmajorityvoting, fullmajorityanalysis, reducingmajorityanalysis, overlapping)
// @Param type query string false "Analysis tpye" Enums(applicability, reliability)
//
// @Success 200 {string} ok
// @Success 500 {string} string
func analysisBlocks(c echo.Context) error {
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
			q.To = 600000
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
}
