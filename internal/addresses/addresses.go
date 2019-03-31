package addresses

import (
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// FirstAppearence returnes true if the passed address appears for the first time in blockchain at the given block height
func FirstAppearence(address string) (int32, error) {
	blocks, err := dgraph.GetAddressBlocksOccurences(&address)
	if err != nil {
		return 0, err
	}
	var minHeight int32
	for i, block := range blocks {
		blockHash, err := chainhash.NewHashFromStr(block)
		if err != nil {
			return 0, err
		}
		block, err := db.GetBlock(blockHash)
		if err != nil {
			return 0, err
		}
		if i == 0 {
			minHeight = block.Height()
		} else {
			minHeight = int32(math.Min(float64(minHeight), float64(block.Height())))
		}
	}
	return minHeight, nil
}
