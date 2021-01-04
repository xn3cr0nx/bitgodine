package trace

import (
	"net/http"

	"github.com/xn3cr0nx/bitgodine/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts /trace based routes on the main group
func Routes(g *echo.Group, s Service) {
	r := g.Group("/trace", validator.JWT())
	r.GET("/address/:address", trace(s))
}

// trace godoc
// @ID trace
//
// @Router /trace/address/{address} [get]
// @Summary Trace address
// @Description get address tracing
// @Tags trace
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "Address"
// @Param limit query int false "Limit"
// @Param skip query int false "Skip"
//
// @Success 200 {object} Flow
// @Success 500 {string} string
func trace(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
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

		res, err := s.TraceAddress(address, q.Limit, q.Skip)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}
