// Package reuse heuristic
// This heuristic just checks if an address that appears
// in the input set, appears also in the output set, we just need to count the
// number of transactions in which this condition is satisfied. This happens
// when a user uses the same address to pay and to recollect the exceeding
// amount of a transaction.
package reuse

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

func contains(recipient []string, element string) bool {
	for _, v := range recipient {
		if v == element {
			return true
		}
	}
	return false
}

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	var inputAddresses []string

	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, err := db.GetTx(in.TxID)
		if err != nil {
			return 0, err
		}
		inputAddresses = append(inputAddresses, spentTx.Vout[in.Vout].ScriptpubkeyAddress)
	}
	// Here on the first matching output, that output is returned as change, but could be a reuse on more outputs?
	for vout, out := range tx.Vout {
		if contains(inputAddresses, out.ScriptpubkeyAddress) {
			return uint32(vout), nil
		}
	}

	return 0, errors.New("No reuse address found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
