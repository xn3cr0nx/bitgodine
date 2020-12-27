// Package peeling chain heuristic
// Since in this case we are looking for a set of transactions
// that have different length, can be really difficult to be sure that a series
// of transactions that have one input and two outputs is a peeling chains
// transaction, also because a chain can be connected in the middle point of
// another chain. In this case, we count all the transactions that belong to a
// series that have almost length 3 as peeling chains transactions.
package peeling

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// PeelingLikeCondition check the basic condition of peeling chain (2 txout and 1 txin)
func PeelingLikeCondition(transaction *tx.Tx) bool {
	return len(transaction.Vout) == 2 && len(transaction.Vin) == 1
}

// IsPeelingChain returns true id the transaction is part of a peeling chain
func IsPeelingChain(db storage.DB, c *cache.Cache, transaction *tx.Tx) (is bool, err error) {
	if !PeelingLikeCondition(transaction) {
		return
	}

	// Check if past transaction is peeling chain
	spentTx, err := tx.GetFromHash(db, c, transaction.Vin[0].TxID)
	if err != nil {
		return
	}
	if PeelingLikeCondition(&spentTx) {
		return true, nil
	}
	// Check if future transaction is peeling chain
	for _, out := range transaction.Vout {
		spendingTx, e := tx.GetSpendingFromHash(db, c, transaction.TxID, out.Index)
		if e != nil {
			err = e
			return
		}
		if PeelingLikeCondition(&spendingTx) {
			return true, nil
		}
	}
	return
}

// ChangeOutput returns the vout of the change address output based on peeling chain heuristic
func ChangeOutput(db storage.DB, ca *cache.Cache, transaction *tx.Tx) (c []uint32, err error) {
	is, err := IsPeelingChain(db, ca, transaction)
	if err != nil {
		return
	}
	if is {
		if transaction.Vout[0].Value <= transaction.Vout[1].Value {
			c = append(c, 1)
		} else {
			c = append(c, 0)
		}
		return
	}
	err = fmt.Errorf("%w: transaction is not peeling chain like", errorx.ErrInvalidArgument)
	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, c *cache.Cache, transaction *tx.Tx) bool {
	_, err := ChangeOutput(db, c, transaction)
	return err == nil
}
