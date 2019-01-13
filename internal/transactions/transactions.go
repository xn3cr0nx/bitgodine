package txs

import (
	"time"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

// Walk parses the btcutil.Tx object
func Walk(tx *btcutil.Tx, v *visitor.BlockchainVisitor, timestamp time.Time, height uint64, blockItem *visitor.BlockItem, outputItems *map[chainhash.Hash][]visitor.OutputItem) btcutil.Tx {
	transactionItem := (*v).VisitTransactionBegin(blockItem)

	for _, i := range tx.MsgTx().TxIn {
		var outputItem visitor.OutputItem
		if occupied, ok := (*outputItems)[(*i).PreviousOutPoint.Hash]; ok {
			outputItem = occupied[(*i).PreviousOutPoint.Index]
			occupied = append(occupied[:(*i).PreviousOutPoint.Index], occupied[(*i).PreviousOutPoint.Index+1:]...)
			// delete(*outputItems, (*i).PreviousOutPoint.Hash)
			// 	if len(occupied) == 0 {
			// 		// occupied.remove()
			// 	}
		}
		(*v).VisitTransactionInput(*i, blockItem, &transactionItem, outputItem)
	}

	curOutputItems := make([]visitor.OutputItem, len(tx.MsgTx().TxOut))
	for n, o := range tx.MsgTx().TxOut {
		outputItem, err := (*v).VisitTransactionOutput(*o, blockItem, &transactionItem)
		if err != nil {
			logger.Error("Transactions", err, logger.Params{"output value": string((*o).Value), "output pkscript": string((*o).PkScript)})
			return btcutil.Tx{}
		}
		curOutputItems[n] = outputItem
	}

	if len(curOutputItems) > 0 {
		(*outputItems)[*tx.Hash()] = curOutputItems
	}

	(*v).VisitTransactionEnd(*tx, blockItem, transactionItem)
	return *tx
}
