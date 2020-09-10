// Package backward heuristic
// It checks the transactions that
// come before the one in which we want to find the change address.
package backward

import (
	"errors"

	"golang.org/x/sync/errgroup"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// ChangeOutput returns the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db storage.DB, ca *cache.Cache, transaction *tx.Tx) (c []uint32, err error) {
	var outputAddresses,
		inputAddresses,
		inputTargets []string
	var spentTxs []tx.Tx
	var outputTargets []uint32

	var g errgroup.Group
	g.Go(func() error {
		for _, out := range transaction.Vout {
			outputAddresses = append(outputAddresses, out.ScriptpubkeyAddress)
		}
		return nil
	})
	g.Go(func() error {
		for _, in := range transaction.Vin {
			if in.IsCoinbase {
				continue
			}
			spentTx, err := tx.GetFromHash(db, ca, in.TxID)
			if err != nil {
				return err
			}
			addr := spentTx.Vout[in.Vout].ScriptpubkeyAddress
			inputAddresses = append(inputAddresses, addr)
			spentTxs = append(spentTxs, spentTx)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		return
	}

	for _, spent := range spentTxs {
		for _, in := range spent.Vin {
			spentTx, e := tx.GetFromHash(db, ca, in.TxID)
			if e != nil {
				return nil, e
			}
			if in.IsCoinbase {
				continue
			}
			addr := spentTx.Vout[in.Vout].ScriptpubkeyAddress
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					inputTargets = append(inputTargets, addr)
				}
			}
			if len(inputTargets) > 0 {
				for target, outputAddr := range outputAddresses {
					if addr == outputAddr {
						outputTargets = append(outputTargets, uint32(target))
					}
				}

				for _, target := range outputTargets {
					for _, input := range inputTargets {
						if outputAddresses[int(target)] != input {
							c = []uint32{target}
							return
						}
					}
				}
				inputTargets, outputTargets = []string{}, []uint32{}
			}
		}
	}

	err = errors.New("No output address matching backward heurisitic requirements found")
	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, ca *cache.Cache, transaction *tx.Tx) bool {
	_, err := ChangeOutput(db, ca, transaction)
	return err == nil
}
