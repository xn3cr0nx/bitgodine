// Package optimal change heuristic
// It tries to locate in the output set of a transaction an
// address that receives an amount which is smaller than all inputs values. We
// count the transactions in which this condition is satisfied.
package optimal

import (
	"errors"
	"golang.org/x/sync/errgroup"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

// ChangeOutput returnes the index of the output which value is less than any inputs value, if there is any
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	values := make([]int64, len(tx.Vin))
	var g errgroup.Group
	for i, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		k, input := i, in
		g.Go(func() error {
			spentTx, err := db.GetTx(input.TxID)
			if err != nil {
				return err
			}
			values[k] = spentTx.Vout[int(input.Vout)].Value
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return 0, err
	}

	var minInput int64
	for i, e := range values {
		if i == 0 || e < minInput {
			minInput = e
		}
	}

	var lowerOuts []uint32
	for _, out := range tx.Vout {
		if out.Value < minInput {
			lowerOuts = append(lowerOuts, out.Index)
		}
	}
	if len(lowerOuts) > 1 {
		return 0, errors.New("More than an out with lower value of an input, ineffective heuristic")
	}
	if len(lowerOuts) == 1 {
		return lowerOuts[0], nil
	}
	return 0, errors.New("No matching output with value inferior to every input")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
