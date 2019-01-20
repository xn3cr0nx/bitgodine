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

// OutputItem is an address type, converted to a string
type OutputItem string
type DoneItem uint

type BlockchainVisitor interface {
	// New() BlockchainVisitor
	VisitBlockBegin(*btcutil.Block, uint64) BlockItem
	VisitBlockEnd(*btcutil.Block, uint64, BlockItem)

	VisitTransactionBegin(*BlockItem) TransactionItem
	VisitTransactionInput(wire.TxIn, *BlockItem, *TransactionItem, OutputItem)

	VisitTransactionOutput(wire.TxOut, *BlockItem, *TransactionItem) (OutputItem, error)
	VisitTransactionEnd(btcutil.Tx, *BlockItem, TransactionItem)

	Done() (DoneItem, error)
}
