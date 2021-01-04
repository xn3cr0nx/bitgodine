// Package shadow heuristic
// This heuristic checks if there are
// output addresses that are the first time they appear in the Blockchain. We
// count the transactions in which appear at least one fresh address in the
// output set.
package shadow

import (
	"runtime"

	"github.com/xn3cr0nx/bitgodine/internal/address"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	task "github.com/xn3cr0nx/bitgodine/internal/errtask"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// ShadowAddress heuristic
type ShadowAddress struct {
	Kv    kv.DB
	Cache *cache.Cache
}

// Worker struct implementing workers pool
type Worker struct {
	db          kv.DB
	ca          *cache.Cache
	output      tx.Output
	vout        int
	candidates  []uint32
	blockHeight int32
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	firstOccurence, err := address.GetFirstOccurenceHeight(w.db, w.ca, w.output.ScriptpubkeyAddress)
	if err != nil {
		return
	}
	if firstOccurence < w.blockHeight {
		w.candidates[w.vout] = 1
	}
	return
}

// ChangeOutput returns the index of the output which appears for the first time in the chain based on client behaviour heuristic
// TODO: violates DRY, just different evaluation in output change, but same operations
func (h *ShadowAddress) ChangeOutput(transaction *tx.Tx) (c []uint32, err error) {
	candidates := make([]uint32, len(transaction.Vout))
	blockHeight, err := block.NewService(h.Kv, h.Cache).GetTxBlockHeight(transaction.TxID)
	if err != nil {
		return
	}

	pool := task.New(runtime.NumCPU())
	for vout, out := range transaction.Vout {
		if out.ScriptpubkeyAddress == "" {
			continue
		}
		pool.Do(&Worker{h.Kv, h.Cache, out, vout, candidates, blockHeight})
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	for i, v := range candidates {
		if v == 1 {
			c = append(c, uint32(i))
		}
	}
	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func (h *ShadowAddress) Vulnerable(transaction *tx.Tx) bool {
	c, err := h.ChangeOutput(transaction)
	return err == nil && len(c) > 0
}
