package visitor

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type BlockItem interface{}

type TransactionItem interface {
	Add(...interface{})
	Size() int
	Values() []interface{}
}

// Utxo is an address type, converted to a string. Represents and unspent transaction output
type Utxo string
type DoneItem int

type BlockchainVisitor interface {
	// New() BlockchainVisitor
	VisitBlockBegin(*btcutil.Block, uint64) BlockItem
	VisitBlockEnd(*btcutil.Block, uint64, BlockItem)

	VisitTransactionBegin(*BlockItem) TransactionItem
	VisitTransactionInput(wire.TxIn, *BlockItem, *TransactionItem, Utxo)

	VisitTransactionOutput(wire.TxOut, *BlockItem, *TransactionItem) (Utxo, error)
	VisitTransactionEnd(btcutil.Tx, *BlockItem, *TransactionItem)

	Done() (DoneItem, error)
}
