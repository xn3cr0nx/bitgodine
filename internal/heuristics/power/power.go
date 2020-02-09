// Package power of ten heuristic
// It checks if in the outputs set there are addresses
// that have an amount in Satoshi that is a power of ten. We count the trans-
// action in which this type of output is found.
package power

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

// ChangeOutput returnes the index of the output which value is power of ten, if there is any and only one
func ChangeOutput(tx *models.Tx) (c []uint32, err error) {
	for k, out := range tx.Vout {
		if (out.Value % 10) == 0 {
			c = append(c, uint32(k))
		}
	}
	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to power heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	c, err := ChangeOutput(tx)
	return err == nil && len(c) > 0
}
