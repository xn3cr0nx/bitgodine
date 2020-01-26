// Package reuse heuristic
// This heuristic just checks if an address that appears
// in the input set, appears also in the output set, we just need to count the
// number of transactions in which this condition is satisfied. This happens
// when a user uses the same address to pay and to recollect the exceeding
// amount of a transaction.
package reuse

import (
	"errors"
	"golang.org/x/sync/errgroup"

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
	var g errgroup.Group
	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		g.Go(func() error {
			spentTx, err := db.GetTx(in.TxID)
			if err != nil {
				return err
			}
			inputAddresses = append(inputAddresses, spentTx.Vout[in.Vout].ScriptpubkeyAddress)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return 0, err
	}

	var candidates []uint32
	for _, out := range tx.Vout {
		if contains(inputAddresses, out.ScriptpubkeyAddress) {
			candidates = append(candidates, out.Index)
		}
	}

	if len(candidates) > 1 {
		return 0, errors.New("More than an output address between inputs, ineffective heuristic")
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	return 0, errors.New("No reuse address found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
