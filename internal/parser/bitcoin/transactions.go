package bitcoin

import (
	"sync"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
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
func TxWalk(tx *txs.Tx, b *blocks.Block, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, wg *sync.WaitGroup, lock *sync.RWMutex, clusterLock *sync.RWMutex) txs.Tx {
	defer wg.Done()
	transactionItem := (*v).VisitTransactionBegin(blockItem)
	parseTxIn(tx, v, blockItem, utxoSet, &transactionItem, lock)
	err := parseTxOut(tx, v, blockItem, utxoSet, &transactionItem, lock, clusterLock)
	if err != nil {
		logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
		return txs.Tx{}
	}
	(*v).VisitTransactionEnd(*tx, blockItem, &transactionItem)
	return *tx
}

// Read the tx inputs removing them from related utxo set. The tx is deleted from utxo set when all outputs are spent
func parseTxIn(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem, lock *sync.RWMutex) {
	for _, i := range tx.MsgTx().TxIn {
		var utxo visitor.Utxo
		lock.RLock()
		occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]
		lock.RUnlock()
		if ok {
			// if occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]; ok {
			utxo = occupied[(*i).PreviousOutPoint.Index]
			occupied[(*i).PreviousOutPoint.Index] = visitor.Utxo("")
			if emptySlice(&occupied) {
				lock.Lock()
				delete(*utxoSet, (*i).PreviousOutPoint.Hash)
				lock.Unlock()
			}
		}
		(*v).VisitTransactionInput(*i, blockItem, transactionItem, utxo)
	}
}

// Creates a new set of utxo to append to the global utxo set (utxoSet)
func parseTxOut(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem, lock *sync.RWMutex, clusterLock *sync.RWMutex) error {
	curUtxoSet := make([]visitor.Utxo, len(tx.MsgTx().TxOut))
	for n, o := range tx.MsgTx().TxOut {
		clusterLock.Lock()
		utxo, err := (*v).VisitTransactionOutput(*o, blockItem, transactionItem)
		if err != nil {
			return err
		}
		clusterLock.Unlock()
		curUtxoSet[n] = utxo
	}
	if len(curUtxoSet) > 0 {
		lock.Lock()
		(*utxoSet)[*tx.Hash()] = curUtxoSet
		lock.Unlock()
	}
	return nil
}
