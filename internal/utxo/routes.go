package utxo

import (
	"net/http"

	"github.com/xn3cr0nx/bitgodine_clusterizer/pkg/utxoset"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /utxo based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/utxo")

	r.GET("", func(c echo.Context) error {
		utxostorage := c.Get("utxoset")
		set, err := utxostorage.(*utxoset.UtxoSet).GetFullUtxoSet()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, set)
	})

	r.GET("/:txid", func(c echo.Context) error {
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,testing"); err != nil {
			return err
		}
		db := c.Get("db")
		t, err := db.(storage.DB).GetTx(txid)
		if err != nil {
			if err.Error() == "transaction not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		return c.JSON(http.StatusOK, t)
	})

	return r
}
