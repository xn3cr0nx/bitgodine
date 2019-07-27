package bitcoin

import (
	"sync"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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
func TxWalk(tx *txs.Tx, b *blocks.Block, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo) txs.Tx {
	transactionItem := (*v).VisitTransactionBegin(blockItem)
	var wg sync.WaitGroup
	var utxoLock = sync.RWMutex{}
	var txItemLock = sync.RWMutex{}
	wg.Add(2)
	alarm := make(chan error)
	go parseTxIn(tx, v, blockItem, utxoSet, &transactionItem, &wg, &utxoLock, &txItemLock)
	// err := parseTxOut(tx, v, blockItem, utxoSet, &transactionItem)
	go parseTxOut(tx, v, blockItem, utxoSet, &transactionItem, &wg, &utxoLock, alarm)
	// if err != nil {
	// 	logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
	// 	return txs.Tx{}
	// }
	wg.Wait()
	select {
	case err := <-alarm:
		{
			logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
			return txs.Tx{}
		}
	default:
	}
	(*v).VisitTransactionEnd(*tx, blockItem, &transactionItem)
	return *tx
}

// Read the tx inputs removing them from related utxo set. The tx is deleted from utxo set when all outputs are spent
func parseTxIn(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem, s *sync.WaitGroup, utxoLock *sync.RWMutex, txItemLock *sync.RWMutex) {
	defer s.Done()
	var wg sync.WaitGroup
	wg.Add(len(tx.MsgTx().TxIn))
	for n := range tx.MsgTx().TxIn {
		k := n
		go func(i *wire.TxIn) {
			defer wg.Done()
			var utxo visitor.Utxo
			utxoLock.RLock()
			occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]
			utxoLock.RUnlock()
			if ok {
				// if occupied, ok := (*utxoSet)[(*i).PreviousOutPoint.Hash]; ok {
				utxo = occupied[(*i).PreviousOutPoint.Index]
				occupied[(*i).PreviousOutPoint.Index] = visitor.Utxo("")
				if emptySlice(&occupied) {
					utxoLock.Lock()
					delete(*utxoSet, (*i).PreviousOutPoint.Hash)
					utxoLock.Unlock()
				}
			}
			txItemLock.Lock()
			(*v).VisitTransactionInput(*i, blockItem, transactionItem, utxo)
			txItemLock.Unlock()
		}(tx.MsgTx().TxIn[k])
	}
	wg.Wait()
}

// Creates a new set of utxo to append to the global utxo set (utxoSet)
func parseTxOut(tx *txs.Tx, v *visitor.BlockchainVisitor, blockItem *visitor.BlockItem, utxoSet *map[chainhash.Hash][]visitor.Utxo, transactionItem *visitor.TransactionItem, s *sync.WaitGroup, utxoLock *sync.RWMutex, alarm chan error) {
	defer s.Done()
	curUtxoSet := make([]visitor.Utxo, len(tx.MsgTx().TxOut))
	var wg sync.WaitGroup
	wg.Add(len(tx.MsgTx().TxOut))
	// alarm := make(chan error)
	for n := range tx.MsgTx().TxOut {
		k := n
		go func(o *wire.TxOut, index int) {
			defer wg.Done()
			utxo, err := (*v).VisitTransactionOutput(*o, blockItem, transactionItem)
			if err != nil {
				// return err
				alarm <- err
			}
			curUtxoSet[index] = utxo
		}(tx.MsgTx().TxOut[k], k)
	}
	wg.Wait()

	if len(curUtxoSet) > 0 {
		utxoLock.Lock()
		(*utxoSet)[*tx.Hash()] = curUtxoSet
		utxoLock.Unlock()
	}
}
