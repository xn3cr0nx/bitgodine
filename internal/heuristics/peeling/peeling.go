package peeling

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// LikePeelingChain check the basic condition of peeling chain (2 txout and 1 txin)
func LikePeelingChain(tx *dgraph.Transaction) bool {
	return len(tx.Outputs) == 2 && len(tx.Inputs) == 1
}

// IsPeelingChain returnes true id the transaction is part of a peeling chain
func IsPeelingChain(tx *dgraph.Transaction) bool {
	if !LikePeelingChain(tx) {
		return false
	}

	// Check if past transaction is peeling chain
	spentTx, err := dgraph.GetTx(tx.Inputs[0].Hash)
	if err != nil {
		return false
	}
	if LikePeelingChain(&spentTx) {
		return true
	}
	// Check if future transaction is peeling chain
	for _, out := range tx.Outputs {
		if dgraph.IsSpent(tx.Hash, out.Vout) == false {
			return false
		}
		spendingTx, err := dgraph.GetFollowingTx(&tx.Hash, &out.Vout)
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
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	if LikePeelingChain(tx) {
		if tx.Outputs[0].Value > tx.Outputs[1].Value {
			return 0, nil
		}
		return 1, nil
	}
	return 0, errors.New("transaction is not like peeling chain")
}
