package heuristics

import (
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
)

// Condition function signature for condition definition
type Condition func(*tx.Tx) bool

// ConditionsSet defines the list of transacions applicable conditions
type ConditionsSet []Condition

func newConditionsSet() ConditionsSet {
	var set ConditionsSet
	set = append(set, coinbaseCondition)
	set = append(set, selfTransferCondition)
	set = append(set, offByOneBugCondition)
	set = append(set, peeling.PeelingLikeCondition)
	return set
}

func conditionsList() []Heuristic {
	return []Heuristic{Coinbase, SelfTransfer, OffByOne, PeelingLike}
}

func (set *ConditionsSet) fillConditionsSet(criteria string) {
	switch criteria {
	case "offbyone":
		*set = append(*set, offByOneBugCondition)
	}
}

func offByOneBugCondition(transaction *tx.Tx) (output bool) {
	if len(transaction.Vout) != 2 {
		output = true
	}
	return
}

func coinbaseCondition(transaction *tx.Tx) (output bool) {
	if len(transaction.Vin) == 1 && transaction.Vin[0].IsCoinbase {
		output = true
	}
	return
}

func selfTransferCondition(transaction *tx.Tx) (output bool) {
	if len(transaction.Vin) > 1 && len(transaction.Vout) == 1 {
		output = true
	}
	return
}
