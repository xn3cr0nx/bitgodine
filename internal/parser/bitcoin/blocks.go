package bitcoin

import (
	"fmt"
	"sync"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	// "github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BlockWalk parses the block and iterates over block's transaction to parse them
func BlockWalk(b *blocks.Block, v *visitor.BlockchainVisitor, height *int32, utxoSet *map[chainhash.Hash][]visitor.Utxo) {
	b.SetHeight(*height)
	blockItem := (*v).VisitBlockBegin(b, *height)
	if *height%100 == 0 {
		logger.Info("Parser Blocks", fmt.Sprintf("Block %d", b.Height()), logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	}
	logger.Debug("Parser Blocks", "storing block", logger.Params{"hash": b.Hash().String(), "height": b.Height()})

	var wg sync.WaitGroup
	var lock = sync.RWMutex{}
	wg.Add(len(b.Transactions()))
	logger.Debug("Blocks", fmt.Sprintf("Dispatching %d threads to parse transactions", len(b.Transactions())), logger.Params{})
	for _, tx := range b.Transactions() {
		go TxWalk(&txs.Tx{Tx: *tx}, b, v, &blockItem, utxoSet, &wg, &lock)
	}
	wg.Wait()

	if err := b.Store(); err != nil {
		logger.Panic("Block Parser", err, logger.Params{})
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}
