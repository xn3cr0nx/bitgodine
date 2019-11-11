package peeling

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
)

// LikePeelingChain check the basic condition of peeling chain (2 txout and 1 txin)
func LikePeelingChain(tx *models.Tx) bool {
	return len(tx.Vout) == 2 && len(tx.Vin) == 1
}

// IsPeelingChain returnes true id the transaction is part of a peeling chain
func IsPeelingChain(db *dgraph.Dgraph, tx *models.Tx) bool {
	if !LikePeelingChain(tx) {
		return false
	}

	// Check if past transaction is peeling chain
	spentTx, err := db.GetTx(tx.Vin[0].TxID)
	if err != nil {
		return false
	}
	if LikePeelingChain(&spentTx) {
		return true
	}
	// Check if future transaction is peeling chain
	for _, out := range tx.Vout {
		if db.IsSpent(tx.TxID, out.Index) == false {
			return false
		}
		spendingTx, err := db.GetFollowingTx(tx.TxID, out.Index)
		if err != nil {
			return false
		}
		if LikePeelingChain(&spendingTx) {
			return true
		}
	}
	return true
}

// ChangeOutput returnes the vout of the change address output based on peeling chain heuristic
func ChangeOutput(db *dgraph.Dgraph, tx *models.Tx) (uint32, error) {
	if LikePeelingChain(tx) {
		if tx.Vout[0].Value > tx.Vout[1].Value {
			return 0, nil
		}
		return 1, nil
	}
	return 0, errors.New("transaction is not like peeling chain")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(dg *dgraph.Dgraph, tx *models.Tx) bool {
	_, err := ChangeOutput(dg, tx)
	return err == nil
}
