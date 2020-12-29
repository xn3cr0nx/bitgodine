// Package optimal change heuristic
// It tries to locate in the output set of a transaction an
// address that receives an amount which is smaller or equal than all inputs values.
// We count the transactions in which this condition is satisfied.
package optimal

import (
	"runtime"

	task "github.com/xn3cr0nx/bitgodine/internal/errtask"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// Worker struct implementing workers pool
type Worker struct {
	db     kv.DB
	ca     *cache.Cache
	txid   string
	vout   uint32
	index  int
	values []int64
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	spentTx, err := tx.GetFromHash(w.db, w.ca, w.txid)
	if err != nil {
		return
	}
	w.values[w.index] = spentTx.Vout[int(w.vout)].Value
	return
}

// ChangeOutput returns the index of the output which value is less than any inputs value, if there is any
func ChangeOutput(db kv.DB, ca *cache.Cache, transaction *tx.Tx) (c []uint32, err error) {
	values := make([]int64, len(transaction.Vin))
	pool := task.New(runtime.NumCPU() / 2)
	for i, in := range transaction.Vin {
		if in.IsCoinbase {
			continue
		}
		pool.Do(&Worker{db, ca, in.TxID, in.Vout, i, values})
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

	for _, out := range transaction.Vout {
		if out.Value <= minInput {
			c = append(c, out.Index)
		}
	}

	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db kv.DB, ca *cache.Cache, transaction *tx.Tx) bool {
	c, err := ChangeOutput(db, ca, transaction)
	return err == nil && len(c) > 0
}
