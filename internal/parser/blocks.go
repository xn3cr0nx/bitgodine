package parser

import (
	"strings"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
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
	if !db.IsStored(b.Hash()) {
		logger.Debug("Parser Blocks", "storing block", logger.Params{"hash": b.Hash().String(), "height": b.Height()})
		err := db.StoreBlock(b)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			logger.Error("Block Parser", err, logger.Params{})
		}
		err = db.StoreLast(b.Hash())
		if err != nil {
			logger.Error("Block Parser", err, logger.Params{})
		}
	} else {
		logger.Debug("Block Parser", "skippin already stored block", logger.Params{"hash": b.Hash().String()})
	}
	for _, tx := range b.Transactions() {
		TxWalk(&txs.Tx{Tx: *tx}, b, v, timestamp, &blockItem, utxoSet)
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}
