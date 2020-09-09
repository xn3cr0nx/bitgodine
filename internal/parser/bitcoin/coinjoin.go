package bitcoin

import (
	"sort"

	mapset "github.com/deckarep/golang-set"
)

// IsCoinjoin returns true is the tx is a coinjoin transaction
func (tx *Tx) IsCoinjoin() bool {
	if len(tx.MsgTx().TxIn) < 2 || len(tx.MsgTx().TxOut) < 3 {
		return false
	}

	participantCount := (len(tx.MsgTx().TxOut) + 1) / 2
	if participantCount > len(tx.MsgTx().TxIn) {
		return false
	}

	inputAddresses := mapset.NewSet()
	for _, txin := range tx.MsgTx().TxIn {
		inputAddresses.Add(txin.PreviousOutPoint.Hash.String())
	}

	if participantCount > inputAddresses.Cardinality() {
		return false
	}

	outputValues := make(map[int64]uint16)
	for _, txout := range tx.MsgTx().TxOut {
		outputValues[txout.Value]++
	}

	keys := make([]int, 0)
	for k := range outputValues {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	if outputValues[int64(keys[len(keys)-2])] != uint16(participantCount) {
		return false
	}

	if outputValues[int64(keys[len(keys)-1])] == 546 || outputValues[int64(keys[len(keys)-1])] == 2730 {
		return false
	}

	return true
}
