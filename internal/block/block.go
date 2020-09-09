package block

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// BlockOut enhanced model block with full transactions
type BlockOut struct {
	models.Block
	Transactions []models.Tx `json:"transactions"`
}

func fetchBlockTxs(db storage.DB, txs []string) (transactions []models.Tx, err error) {
	for _, hash := range txs {
		transaction, e := tx.GetTxFromHash(db, hash)
		if e != nil {
			return nil, e
		}
		transactions = append(transactions, transaction)
	}
	return
}

// GetBlockFromHeight return block structure based on block height
func GetBlockFromHeight(db storage.DB, height int32) (*BlockOut, error) {
	b, err := db.GetBlockFromHeight(height)
	if err != nil {
		if err.Error() == "Block not found" {
			return nil, echo.NewHTTPError(http.StatusNotFound)
		}
		return nil, err
	}

	txs, err := fetchBlockTxs(db, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetBlockFromHashWithTxs return block structure based on block hash
func GetBlockFromHashWithTxs(db storage.DB, hash string) (*BlockOut, error) {
	b, err := db.GetBlockFromHash(hash)
	if err != nil {
		if err.Error() == "Block not found" {
			return nil, echo.NewHTTPError(http.StatusNotFound)
		}
		return nil, err
	}

	txs, err := fetchBlockTxs(db, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetBlockFromHash return block structure based on block hash
func GetBlockFromHash(db storage.DB, hash string) (models.Block, error) {
	b, err := db.GetBlockFromHash(hash)
	if err != nil {
		if err.Error() == "Block not found" {
			return models.Block{}, echo.NewHTTPError(http.StatusNotFound)
		}
		return models.Block{}, err
	}

	return b, nil
}

// GetLastBlock return last synced block
func GetLastBlock(db storage.DB) (*BlockOut, error) {
	b, err := db.LastBlock()
	if err != nil {
		return nil, err
	}

	txs, err := fetchBlockTxs(db, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}
