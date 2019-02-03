package parser

import (
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BlockWalk parses the block and iterates over block's transaction to parse them
func BlockWalk(b *blocks.Block, v *visitor.BlockchainVisitor, height *uint64, utxoSet *map[chainhash.Hash][]visitor.Utxo) {
	timestamp := b.MsgBlock().Header.Timestamp
	blockItem := (*v).VisitBlockBegin(b, *height)
	for _, tx := range b.Transactions() {
		TxWalk(&txs.Tx{Tx: *tx}, v, timestamp, *height, &blockItem, utxoSet)
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}
