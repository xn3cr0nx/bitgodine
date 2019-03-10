package visitor

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
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
	VisitBlockBegin(*blocks.Block, int32) BlockItem
	VisitBlockEnd(*blocks.Block, int32, BlockItem)

	VisitTransactionBegin(*BlockItem) TransactionItem
	VisitTransactionInput(wire.TxIn, *BlockItem, *TransactionItem, Utxo)

	VisitTransactionOutput(wire.TxOut, *BlockItem, *TransactionItem) (Utxo, error)
	VisitTransactionEnd(txs.Tx, *BlockItem, *TransactionItem)

	Done() (DoneItem, error)
}
