package parser

import (
	"fmt"
	"strings"

	bdg "github.com/xn3cr0nx/bitgodine_code/internal/bdg"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BlockWalk parses the block and iterates over block's transaction to parse them
func BlockWalk(b *blocks.Block, v *visitor.BlockchainVisitor, height *uint64, utxoSet *map[chainhash.Hash][]visitor.Utxo) {
	timestamp := b.MsgBlock().Header.Timestamp
	blockItem := (*v).VisitBlockBegin(b, *height)
	fmt.Println("block", b.Hash().String(), "already stored?", bdg.IsStored(b.Hash()))
	if !bdg.IsStored(b.Hash()) {
		logger.Info("Parser Blocks", "storing block", logger.Params{"hash": b.Hash().String()})
		err := bdg.StoreBlock(b)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			logger.Error("Block Parser", err, logger.Params{})
		}
	} else {
		logger.Info("Block Parser", "skippin already stored block", logger.Params{"hash": b.Hash().String()})
	}
	for _, tx := range b.Transactions() {
		TxWalk(&txs.Tx{Tx: *tx}, b, v, timestamp, *height, &blockItem, utxoSet)
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}
