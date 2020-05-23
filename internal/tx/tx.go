package tx

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

// GetTxFromHash return block structure based on block hash
func GetTxFromHash(db storage.DB, hash string) (models.Tx, error) {
	tx, err := db.GetTx(hash)
	if err != nil {
		if err.Error() == "transaction not found" {
			return models.Tx{}, echo.NewHTTPError(http.StatusNotFound)
		}
		return models.Tx{}, err
	}

	return tx, nil
}
