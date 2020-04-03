package abuse

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"
)

// Routes mounts all /abuse based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/abuses")

	r.GET("", func(c echo.Context) error {
		type Query struct {
			Output bool `query:"output" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		abuses, err := GetAbuses(&c, q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, abuses)
	})

	r.POST("", func(c echo.Context) error {
		b := new(Model)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		if err := CreateAbuse(&c, b); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, "")
	})

	r.GET("/:address", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		type Query struct {
			Output bool `query:"output" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		abuses, err := GetAbuse(&c, address, q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, abuses)
	})

	r.GET("/cluster/:address", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		clusters, err := GetAbusedCluster(&c, address)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, clusters)
	})

	r.GET("/cluster/:address/set", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		clusters, err := GetAbusedClusterSet(&c, address)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, clusters)
	})

	return r
}
