package heuristics

import (
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// Condition function signature for condition definition
type Condition func(*models.Tx) bool

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

func offByOneBugCondition(tx *models.Tx) (output bool) {
	if len(tx.Vout) != 2 {
		output = true
	}
	return
}

func coinbaseCondition(tx *models.Tx) (output bool) {
	if len(tx.Vin) == 1 && tx.Vin[0].IsCoinbase {
		output = true
	}
	return
}

func selfTransferCondition(tx *models.Tx) (output bool) {
	if len(tx.Vin) > 1 && len(tx.Vout) == 1 {
		output = true
	}
	return
}
