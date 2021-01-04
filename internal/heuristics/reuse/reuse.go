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

// AddressReuse heuristic
type AddressReuse struct {
	Kv    kv.DB
	Cache *cache.Cache
}

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
	service        tx.Service
	txid           string
	vout           uint32
	index          int
	inputAddresses []string
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	spentTx, err := w.service.GetFromHash(w.txid)
	if err != nil {
		return
	}
	w.inputAddresses[w.index] = spentTx.Vout[w.vout].ScriptpubkeyAddress
	return
}

// ChangeOutput returns the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func (h *AddressReuse) ChangeOutput(transaction *tx.Tx) (c []uint32, err error) {
	inputAddresses := make([]string, len(transaction.Vin))
	pool := task.New(runtime.NumCPU() / 2)
	txService := tx.NewService(h.Kv, h.Cache)
	for i, in := range transaction.Vin {
		if in.IsCoinbase {
			continue
		}
		pool.Do(&Worker{txService, in.TxID, in.Vout, i, inputAddresses})
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
func (h *AddressReuse) Vulnerable(transaction *tx.Tx) bool {
	c, err := h.ChangeOutput(transaction)
	return err == nil && len(c) > 0
}
