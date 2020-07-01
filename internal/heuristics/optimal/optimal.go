// Package optimal change heuristic
// It tries to locate in the output set of a transaction an
// address that receives an amount which is smaller or equal than all inputs values.
// We count the transactions in which this condition is satisfied.
package optimal

import (
	"runtime"

	task "github.com/xn3cr0nx/bitgodine/internal/errtask"

	"github.com/xn3cr0nx/bitgodine/pkg/models"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"
)

// Worker struct implementing workers pool
type Worker struct {
	db     storage.DB
	txid   string
	vout   uint32
	index  int
	values []int64
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	spentTx, err := w.db.GetTx(w.txid)
	if err != nil {
		return
	}
	w.values[w.index] = spentTx.Vout[int(w.vout)].Value
	return
}

// ChangeOutput returnes the index of the output which value is less than any inputs value, if there is any
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	values := make([]int64, len(tx.Vin))
	pool := task.New(runtime.NumCPU() / 2)
	for i, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		pool.Do(&Worker{db, in.TxID, in.Vout, i, values})
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	var minInput int64
	for i, e := range values {
		if i == 0 || e < minInput {
			minInput = e
		}
	}

	for _, out := range tx.Vout {
		if out.Value <= minInput {
			c = append(c, out.Index)
		}
	}

	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	c, err := ChangeOutput(db, tx)
	return err == nil && len(c) > 0
}
