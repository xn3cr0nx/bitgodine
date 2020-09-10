package tx

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// read retrieves tx by hash
func read(db storage.DB, hash string) (transaction Tx, err error) {
	r, err := db.Read(hash)
	if err != nil {
		return
	}
	if err = encoding.Unmarshal(r, &transaction); err != nil {
		return
	}
	return
}

// readFollowing retrieves spending tx of the output based on hash and index
func readFollowing(db storage.DB, hash string, vout uint32) (transaction string, err error) {
	bytes, err := db.Read(hash + "_" + string(vout))
	if err != nil {
		return
	}
	transaction = string(bytes)
	return
}

// GetFromHash return block structure based on block hash
func GetFromHash(db storage.DB, c *cache.Cache, hash string) (transaction Tx, err error) {
	if cached, ok := c.Get(hash); ok {
		transaction = cached.(Tx)
		return
	}

	tx, err := read(db, hash)
	if err != nil {
		if err.Error() == "transaction not found" {
			return Tx{}, echo.NewHTTPError(http.StatusNotFound)
		}
		return Tx{}, err
	}

	if !c.Set(transaction.TxID, transaction, 1) {
		logger.Error("Cache", errors.New("error caching"), logger.Params{"hash": transaction.TxID})
	}
	return tx, nil
}

// GetOutputsFromHash retrieves tx's outputs by hash
func GetOutputsFromHash(db storage.DB, c *cache.Cache, hash string) (outputs []Output, err error) {
	tx, err := GetFromHash(db, c, hash)
	if err != nil {
		return
	}
	outputs = tx.Vout
	return
}

// GetSpentOutputFromHash retrieves spent tx output based on hash and index
func GetSpentOutputFromHash(db storage.DB, c *cache.Cache, hash string, vout uint32) (output Output, err error) {
	tx, err := GetFromHash(db, c, hash)
	if err != nil {
		return
	}
	output = tx.Vout[vout]
	return
}

// GetSpendingFromHash retrieves spending tx of the output based on hash and index
func GetSpendingFromHash(db storage.DB, c *cache.Cache, hash string, vout uint32) (transaction Tx, err error) {
	spendingHash, err := readFollowing(db, hash, vout)
	if err != nil {
		return
	}
	transaction, err = GetFromHash(db, c, spendingHash)
	return
}

// IsSpent returnes true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func IsSpent(db storage.DB, c *cache.Cache, tx string, index uint32) bool {
	_, err := GetSpendingFromHash(db, c, tx, index)
	return err == nil
}
