// Package power of ten heuristic
// It checks if in the outputs set there are addresses
// that have an amount in Satoshi that is a power of ten. We count the trans-
// action in which this type of output is found.
package power

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

// ChangeOutput returnes the index of the output which value is power of ten, if there is any and only one
func ChangeOutput(tx *models.Tx) (uint32, error) {
	var powerOutputs []uint32
	for k, out := range tx.Vout {
		if (out.Value % 10) == 0 {
			powerOutputs = append(powerOutputs, uint32(k))
		}
	}
	if len(powerOutputs) > 1 {
		return 0, errors.New("More than an output which value is power of ten, heuristic ineffective")
	}
	if len(powerOutputs) == 1 {
		return powerOutputs[0], nil
	}
	return 0, errors.New("No output value power of ten found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to power heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(tx)
	return err == nil
}
