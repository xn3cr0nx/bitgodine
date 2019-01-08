package visitor

import "github.com/btcsuite/btcutil"

type BlockItem interface{}

type TransactionItem interface {
	Add(...interface{})
	Size() int
	Values() []interface{}
}
type OutputItem btcutil.Address
type DoneItem uint

type BlockchainVisitor interface {

	// new() BlockchainVisitor
	visitBlockBegin() BlockItem
	visitBlockEnd()

	visitTransactionBegin() TransactionItem
	visitTransactionInput()

	visitTransactionOutput() (OutputItem, error)
	visitTransactionEnd()

	done() (DoneItem, error)
}
