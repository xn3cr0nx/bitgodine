package analysis

import "github.com/xn3cr0nx/bitgodine_parser/pkg/models"

func offByOneBugCondition(tx models.Tx) (output bool) {
	if len(tx.Vout) != 2 {
		output = true
	}
	return
}

func coinbaseCondition(tx models.Tx) (output bool) {
	if len(tx.Vin) == 1 && tx.Vin[0].IsCoinbase {
		output = true
	}
	return
}

func selfTransferCondition(tx models.Tx) (output bool) {
	if len(tx.Vin) > 1 && len(tx.Vout) == 1 {
		output = true
	}
	return
}
