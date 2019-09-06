package address

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"
)

// Routes mounts all /address based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/address")

	r.GET("/:address", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "OK")
	})

	r.GET("/:address/txs", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "OK")
	})

	r.GET("/:address/txs/chain/:last_seen_txid", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "OK")
	})

	r.GET("/:address/txs/mempool", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "OK")
	})

	r.GET("/:address/utxo", func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "OK")
	})

	return r
}
