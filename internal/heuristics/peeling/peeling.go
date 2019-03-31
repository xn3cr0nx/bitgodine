package peeling

import (
	"errors"

	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

// LikePeelingChain check the basic condition of peeling chain (2 txout and 1 txin)
func LikePeelingChain(tx *txs.Tx) bool {
	return len((*tx).MsgTx().TxOut) == 2 && len((*tx).MsgTx().TxIn) == 1
}

// IsPeelingChain returnes true id the transaction is part of a peeling chain
func IsPeelingChain(tx *txs.Tx) bool {
	if !LikePeelingChain(tx) {
		return false
	}

	// Check if past transaction is peeling chain
	spentTx, err := tx.GetSpentTx(0)
	if err != nil {
		return false
	}
	if LikePeelingChain(&spentTx) {
		return true
	}

	// Check if future transaction is peeling chain
	for i := range tx.MsgTx().TxOut {
		index := uint32(i)
		if tx.IsSpent(index) == false {
			return false
		}
		spendingTx, err := tx.GetSpendingTx(index)
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
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	if LikePeelingChain(tx) {
		if (*tx).MsgTx().TxOut[0].Value > (*tx).MsgTx().TxOut[1].Value {
			return 0, nil
		}
		return 1, nil
	}
	return 0, errors.New("transaction is not like peeling chain")
}
