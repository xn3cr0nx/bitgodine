package bitcoin

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	// "github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BlockWalk parses the block and iterates over block's transaction to parse them
func BlockWalk(b *blocks.Block, v *visitor.BlockchainVisitor, height *int32, utxoSet *map[chainhash.Hash][]visitor.Utxo) {
	timestamp := b.MsgBlock().Header.Timestamp
	b.SetHeight(*height)
	blockItem := (*v).VisitBlockBegin(b, *height)
	if *height%100 == 0 {
		logger.Info("Parser Blocks", fmt.Sprintf("Block %d", b.Height()), logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	}
	logger.Debug("Parser Blocks", "storing block", logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	if err := b.Store(); err != nil {
		logger.Panic("Block Parser", err, logger.Params{})
	}
	for _, tx := range b.Transactions() {
		TxWalk(&txs.Tx{Tx: *tx}, b, v, timestamp, &blockItem, utxoSet)
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}
