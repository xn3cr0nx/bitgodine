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
		type Query struct {
			Limit int `query:"limit" validate:"omitempty,gt=0"`
			Skip  int `query:"skip" validate:"omitempty,gte=0"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}
		if q.Limit == 0 {
			q.Limit = 5
		}

		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		res, err := traceAddress(&c, address, q.Limit, q.Skip)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	})

	return r
}
