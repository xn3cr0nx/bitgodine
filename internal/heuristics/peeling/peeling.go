// Package peeling chain heuristic
// Since in this case we are looking for a set of transactions
// that have different length, can be really difficult to be sure that a series
// of transactions that have one input and two outputs is a peeling chains
// transaction, also because a chain can be connected in the middle point of
// another chain. In this case, we count all the transactions that belong to a
// series that have almost length 3 as peeling chains transactions.
package peeling

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// PeelingLikeCondition check the basic condition of peeling chain (2 txout and 1 txin)
func PeelingLikeCondition(tx *models.Tx) bool {
	return len(tx.Vout) == 2 && len(tx.Vin) == 1
}

// IsPeelingChain returnes true id the transaction is part of a peeling chain
func IsPeelingChain(db storage.DB, tx *models.Tx) (is bool, err error) {
	if !PeelingLikeCondition(tx) {
		return
	}

	// Check if past transaction is peeling chain
	spentTx, err := db.GetTx(tx.Vin[0].TxID)
	if err != nil {
		return
	}
	if PeelingLikeCondition(&spentTx) {
		return true, nil
	}
	// Check if future transaction is peeling chain
	for _, out := range tx.Vout {
		spendingTx, e := db.GetFollowingTx(tx.TxID, out.Index)
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

// ChangeOutput returnes the vout of the change address output based on peeling chain heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	is, err := IsPeelingChain(db, tx)
	if err != nil {
		return
	}
	if is {
		if tx.Vout[0].Value <= tx.Vout[1].Value {
			c = append(c, 1)
		} else {
			c = append(c, 0)
		}
		return
	}
	err = errors.New("transaction is not like peeling chain")
	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
