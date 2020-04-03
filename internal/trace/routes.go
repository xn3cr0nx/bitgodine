package trace

import (
	"net/http"

	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/trace")

	r.GET("/address/:address", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		res, err := traceAddress(&c, address)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)

		// type Query struct {
		// 	From  int32    `query:"from" validate:"omitempty,gte=0"`
		// 	To    int32    `query:"to" validate:"omitempty,gtfield=From"`
		// 	List  []string `query:"heuristics" validate:"dive,oneof=locktime peeling power optimal exact type reuse shadow client forward backward"`
		// 	Plot  string   `query:"plot" validate:"omitempty,oneof=timeline percentage"`
		// 	Force bool     `query:"force" validate:"omitempty"`
		// }
		// q := new(Query)
		// if err := validator.Struct(&c, q); err != nil {
		// 	return err
		// }

		// var list []string
		// for _, h := range q.List {
		// 	list = append(list, heuristics.Abbreviation(h))
		// }
		// if len(list) == 0 {
		// 	list = heuristics.List()
		// }

		// vuln, err := AnalyzeBlocks(&c, q.From, q.To, list, q.Force, q.Plot)
		// if err != nil {
		// 	return err
		// }

		// if q.Plot != "" {
		// 	return c.File(q.Plot + ".png")
		// }
		// return c.JSON(http.StatusOK, vuln)
	})

	return r
}
