// Package class heuristic
// This heuristic is the address type heuristic and it checks if the all the inputs
// are of the same type and then try to locate only one output
// that is of the same type. Again, we just need to check a simple condition.
package class

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

// ChangeOutput returnes the index of the output which address type corresponds to input addresses type
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	inputTypes := make([]string, len(tx.Vin))
	outputTypes := make([]string, len(tx.Vout))

	var g errgroup.Group
	g.Go(func() error {
		for i, in := range tx.Vin {
			if in.IsCoinbase {
				continue
			}
			spentTx, err := db.GetTx(in.TxID)
			if err != nil {
				return err
			}
			inputTypes[i] = spentTx.Vout[in.Vout].ScriptpubkeyType
			if inputTypes[i] != inputTypes[0] {
				return errors.New("There are different kind of addresses between inputs")
			}
		}
		return nil
	})
	g.Go(func() error {
		for o, out := range tx.Vout {
			outputTypes[o] = out.ScriptpubkeyType
			if outputTypes[o] != outputTypes[0] {
				return errors.New("Two or more output of the same type, cannot determine change output")
			}
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return 0, err
	}

	for _, input := range inputTypes {
		for vout, output := range outputTypes {
			if input == output {
				return uint32(vout), nil
			}
		}
	}
	return 0, errors.New("No address type matching input addresses type found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
