package bitcoin

import (
	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// BlockWalk parses the block and iterates over block's transaction to parse them
func BlockWalk(p *Parser, b *blocks.Block, height *int32) (err error) {
	b.SetHeight(*height)
	if *height%100 == 0 {
		logger.Info("Parser Blocks", "Block "+string(b.Height()), logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	}
	logger.Debug("Parser Blocks", "Storing block", logger.Params{"hash": b.Hash().String(), "height": *height})
	if err = b.Store(p.db); err != nil {
		return
	}
	return
}
