// Package reuse heuristic
// This heuristic just checks if an address that appears
// in the input set, appears also in the output set, we just need to count the
// number of transactions in which this condition is satisfied. This happens
// when a user uses the same address to pay and to recollect the exceeding
// amount of a transaction.
package reuse

import (
	"runtime"

	task "github.com/xn3cr0nx/bitgodine/internal/errtask"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

func contains(recipient []string, element string) bool {
	for _, v := range recipient {
		if v == element {
			return true
		}
	}
	return false
}

// Worker struct implementing workers pool
type Worker struct {
	db             kv.DB
	ca             *cache.Cache
	txid           string
	vout           uint32
	index          int
	inputAddresses []string
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	spentTx, err := tx.GetFromHash(w.db, w.ca, w.txid)
	if err != nil {
		return
	}
	w.inputAddresses[w.index] = spentTx.Vout[w.vout].ScriptpubkeyAddress
	return
}

// ChangeOutput returns the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db kv.DB, ca *cache.Cache, transaction *tx.Tx) (c []uint32, err error) {
	inputAddresses := make([]string, len(transaction.Vin))
	pool := task.New(runtime.NumCPU() / 2)
	for i, in := range transaction.Vin {
		if in.IsCoinbase {
			continue
		}
		pool.Do(&Worker{db, ca, in.TxID, in.Vout, i, inputAddresses})
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	for _, out := range transaction.Vout {
		if contains(inputAddresses, out.ScriptpubkeyAddress) {
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
