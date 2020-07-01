package tag

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts all /tag based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/tags")

	r.GET("", func(c echo.Context) error {
		type Query struct {
			Output bool `query:"output" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		tags, err := GetTags(&c, q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, tags)
	})

	r.POST("", func(c echo.Context) error {
		b := new(Model)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		if err := CreateTag(&c, b); err != nil {
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

		tags, err := GetTag(&c, address, q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, tags)
	})

	r.GET("/cluster/:address", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		clusters, err := GetTaggedCluster(&c, address)
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

		clusters, err := GetTaggedClusterSet(&c, address)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, clusters)
	})

	return r
}
