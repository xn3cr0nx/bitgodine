package txs

import (
	"time"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

func emptySlice(arr *[]visitor.Utxo) bool {
	for _, e := range *arr {
		if e != "" {
			return false
		}
	}
	return true
}

// Walk parses the btcutil.Tx object
func Walk(tx *btcutil.Tx, v *visitor.BlockchainVisitor, timestamp time.Time, height uint64, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo) btcutil.Tx {
	transactionItem := (*v).VisitTransactionBegin(blockItem)

	for _, i := range tx.MsgTx().TxIn {
		var utxo visitor.Utxo
		if occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]; ok {
			utxo = occupied[(*i).PreviousOutPoint.Index]
			occupied[(*i).PreviousOutPoint.Index] = visitor.Utxo("")
			if emptySlice(&occupied) {
				delete(*utxoSet, (*i).PreviousOutPoint.Hash)
			}
		}
		(*v).VisitTransactionInput(*i, blockItem, &transactionItem, utxo)
	}

	curUtxoSet := make([]visitor.Utxo, len(tx.MsgTx().TxOut))
	for n, o := range tx.MsgTx().TxOut {
		utxo, err := (*v).VisitTransactionOutput(*o, blockItem, &transactionItem)
		if err != nil {
			logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String(), "output value": string((*o).Value)})
			return btcutil.Tx{}
		}
		curUtxoSet[n] = utxo
	}

	if len(curUtxoSet) > 0 {
		(*utxoSet)[*tx.Hash()] = curUtxoSet
	}

	(*v).VisitTransactionEnd(*tx, blockItem, &transactionItem)
	return *tx
}
