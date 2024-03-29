// Package forward heuristic
// It checks the transactions that come
// after the one in which we want to find the change address.
package forward

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Forward heuristic
type Forward struct {
	Kv    kv.DB
	Cache *cache.Cache
}

// ChangeOutput returns the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func (h *Forward) ChangeOutput(transaction *tx.Tx) (c []uint32, err error) {
	var inputAddresses []string

	logger.Debug("Forward Heuristic", "transaction "+transaction.TxID, logger.Params{})

	txService := tx.NewService(h.Kv, h.Cache)

	for _, in := range transaction.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, e := txService.GetFromHash(in.TxID)
		if e != nil {
			return nil, e
		}
		inputAddresses = append(inputAddresses, spentTx.Vout[in.Vout].ScriptpubkeyAddress)
	}

	for _, out := range transaction.Vout {
		spendingTx, e := txService.GetSpendingFromHash(transaction.TxID, out.Index)
		if e != nil {
			// transaction not found => output not yet spent, but we can identify the change output anyway
			if errors.Is(err, errorx.ErrKeyNotFound) {
				continue
			}
			return nil, e
		}
		index := out.Index
		for _, spendingIn := range spendingTx.Vin {
			// check if the input is the one the spending transaction is reached from
			if spendingIn.Vout == index {
				continue
			}
			spentTx, e := txService.GetFromHash(spendingIn.TxID)
			if e != nil {
				return nil, e
			}
			addr := spentTx.Vout[spendingIn.Vout].ScriptpubkeyAddress
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					c = []uint32{index}
					return
				}
			}
		}
	}

	err = fmt.Errorf("%w: No output address matching forward heurisitic requirements", errorx.ErrNotFound)
	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func (h *Forward) Vulnerable(transaction *tx.Tx) bool {
	_, err := h.ChangeOutput(transaction)
	return err == nil
}
