// Package class heuristic
// This heuristic is the address type heuristic and it checks if the all the inputs
// are of the same type and then try to locate only one output
// that is of the same type. Again, we just need to check a simple condition.
package class

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// ChangeOutput returns the index of the output which address type corresponds to input addresses type
func ChangeOutput(db storage.DB, ca *cache.Cache, transaction *tx.Tx) (c []uint32, err error) {
	inputTypes := make([]string, len(transaction.Vin))
	outputTypes := make([]string, len(transaction.Vout))

	var g errgroup.Group
	g.Go(func() error {
		for o, out := range transaction.Vout {
			outputTypes[o] = out.ScriptpubkeyType
			if o > 0 && outputTypes[o] == outputTypes[0] {
				return fmt.Errorf("%w: Two or more output of the same type, cannot determine change output", errorx.ErrUnknown)
			}
		}
		return nil
	})
	g.Go(func() error {
		for i, in := range transaction.Vin {
			if in.IsCoinbase {
				continue
			}
			spentTx, err := tx.GetFromHash(db, ca, in.TxID)
			if err != nil {
				return err
			}
			inputTypes[i] = spentTx.Vout[in.Vout].ScriptpubkeyType
			if inputTypes[i] != inputTypes[0] {
				return fmt.Errorf("%w: different kind of addresses between inputs", errorx.ErrUnknown)
			}
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		return
	}

	for _, input := range inputTypes {
		for vout, output := range outputTypes {
			if input == output {
				c = append(c, uint32(vout))
			}
		}
	}

	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, ca *cache.Cache, transaction *tx.Tx) bool {
	c, err := ChangeOutput(db, ca, transaction)
	return err == nil && len(c) > 0
}
