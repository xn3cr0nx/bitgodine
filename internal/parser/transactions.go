package parser

import (
	"time"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
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

// TxWalk parses the txs.Tx object
func TxWalk(tx *txs.Tx, b *blocks.Block, v *visitor.BlockchainVisitor, timestamp time.Time, height uint64, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo) txs.Tx {
	transactionItem := (*v).VisitTransactionBegin(blockItem)
	// level, _ := db.LevelDB(nil)
	dg := dgraph.Instance(&dgraph.Config{"localhost", 9080})
	err := dgraph.StoreTx(dg, tx.Hash().String(), b.Hash().String(), tx.MsgTx().LockTime, tx.MsgTx().TxIn)
	if err != nil {
		logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
		return txs.Tx{}
	}

	parseTxIn(tx, v, blockItem, utxoSet, &transactionItem)
	err = parseTxOut(tx, v, blockItem, utxoSet, &transactionItem)
	if err != nil {
		logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
		return txs.Tx{}
	}
	(*v).VisitTransactionEnd(*tx, blockItem, &transactionItem)
	return *tx
}

// Read the tx inputs removing them from related utxo set. The tx is deleted from utxo set when all outputs are spent
func parseTxIn(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem) {
	for _, i := range tx.MsgTx().TxIn {
		var utxo visitor.Utxo
		if occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]; ok {
			utxo = occupied[(*i).PreviousOutPoint.Index]
			occupied[(*i).PreviousOutPoint.Index] = visitor.Utxo("")
			if emptySlice(&occupied) {
				delete(*utxoSet, (*i).PreviousOutPoint.Hash)
			}
		}
		(*v).VisitTransactionInput(*i, blockItem, transactionItem, utxo)
	}
}

// Creates a new set of utxo to append to the global utxo set (utxoSet)
func parseTxOut(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem) error {
	curUtxoSet := make([]visitor.Utxo, len(tx.MsgTx().TxOut))
	for n, o := range tx.MsgTx().TxOut {
		utxo, err := (*v).VisitTransactionOutput(*o, blockItem, transactionItem)
		if err != nil {
			return err
		}
		curUtxoSet[n] = utxo
	}
	if len(curUtxoSet) > 0 {
		(*utxoSet)[*tx.Hash()] = curUtxoSet
	}
	return nil
}
