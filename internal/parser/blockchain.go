package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Walk goes through the blockchain block by block
func Walk(bc *blockchain.Blockchain, v visitor.BlockchainVisitor, interrupt, done chan int) (int32, *chainhash.Hash, map[chainhash.Hash][]visitor.Utxo) {
	skipped := make(map[chainhash.Hash]blocks.Block)
	// Hashmap that represents the utxos set. For each transaction keeps track of utxo associated
	// with each transaction mapping the tx hash to the array of related utxo
	utxoSet := make(map[chainhash.Hash][]visitor.Utxo)
	goalPrevHash, _ := chainhash.NewHash(make([]byte, 32))
	var lastBlock blocks.Block
	prevHeight := int32(0)
	height := bc.Height()

	// check if "coinbase output" is stored in dgraph
	hash := strings.Repeat("0", 64)
	if _, err := dgraph.GetTxUID(&hash); err != nil {
		logger.Debug("Blockchain", "missing coinbase outputs", logger.Params{"hash": hash})
		dgraph.StoreTx(hash, "", 0, 0, nil, []*wire.TxOut{wire.NewTxOut(int64(5000000000), nil), wire.NewTxOut(int64(2500000000), nil), wire.NewTxOut(int64(1250000000), nil)})
	}

	var rawChain [][]uint8
	for _, ref := range bc.Maps {
		rawChain = append(rawChain, []uint8(ref))
	}
	logger.Debug("Blockchain", fmt.Sprintf("Files converted to be parsed: %d", len(rawChain)), logger.Params{})

	if height > 0 {
		logger.Debug("Blockchain", fmt.Sprintf("reaching endpoint to reach %d", height), logger.Params{})
		if err := findCheckPoint(&rawChain, &prevHeight, &height); err != nil {
		}
		last, err := bc.Head()
		if err != nil {
			logger.Panic("Blockchain", err, logger.Params{})
		}
		goalPrevHash = (*last).Hash()
		lastBlock = *last
		logger.Debug("Blockchain", "Last Block", logger.Params{"hash": lastBlock.Hash().String()})
	}
	logger.Info("Blockchain", fmt.Sprintf("Starting syncing from block %d", height), logger.Params{})

	for k, ref := range rawChain {
		if len(ref) == 0 {
			continue
		}
		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": fmt.Sprintf("%v/%v", k, len(bc.Maps)-1)})
		WalkSlice(&ref, goalPrevHash, &lastBlock, &height, &skipped, &utxoSet, &v, interrupt, done)
	}

	return height, goalPrevHash, utxoSet
}

// WalkSlice goes through a slice (block) of the chain
func WalkSlice(slice *[]uint8, goalPrevHash *chainhash.Hash, lastBlock *blocks.Block, height *int32, skipped *map[chainhash.Hash]blocks.Block, utxoSet *map[chainhash.Hash][]visitor.Utxo, v *visitor.BlockchainVisitor, interrupt, done chan int) {
	for len(*slice) > 0 {
		select {
		case x, ok := <-interrupt:
			if ok {
				logger.Info("Blockchain", "Received interrupt signal", logger.Params{"signal": x})
				select {
				case _, ok := <-done:
					fmt.Println("Received done")
					if ok {
						logger.Info("Blockchain", "Exporting done", logger.Params{})
						return
					}
				}
			}
			logger.Error("Blockchain", errors.New("interrupt received but something wrong"), logger.Params{"signal": x})
			return

		default:
			if _, ok := (*skipped)[*goalPrevHash]; ok {
				BlockWalk(lastBlock, v, height, utxoSet)
				logger.Debug("Blockchain", fmt.Sprintf("(rewind - pre-step) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
				(*height)++
				// Here I should do the for loop removing every goal_prev_hash and
				// walking the block obtained at the index of goal_prev_hash
				for {
					if block, ok := (*skipped)[*goalPrevHash]; ok {
						delete(*skipped, *goalPrevHash)
						BlockWalk(&block, v, height, utxoSet)
						logger.Debug("Blockchain", fmt.Sprintf("(rewind) Block %v - %v -> %v", *height, block.MsgBlock().Header.PrevBlock.String(), block.Hash().String()), logger.Params{})
						(*height)++
						*goalPrevHash = *block.Hash()
						// possible bug initialization to null (None in rust)
						*lastBlock = blocks.Block{}
						continue
					}
					break
				}
			}

			block, err := blocks.Parse(slice)
			if err != nil {
				if len(*slice) != 0 {
					logger.Panic("Blockchain", errors.New("Block not found but the slice of blocks is not empty"), logger.Params{})
				}
				break
			}

			logger.Debug("Blockchain", fmt.Sprintf("Block candidate for height %d - goal_prev_hash = %v, prev_hash = %v, cur_hash = %v", *height, goalPrevHash.String(), block.MsgBlock().Header.PrevBlock.String(), block.Hash().String()), logger.Params{})

			logger.Debug("Blockchain", "Checking Prev block equal prev goal hash", logger.Params{"prev": block.MsgBlock().Header.PrevBlock.String(), "prev_goal": goalPrevHash.String(), "cond": block.MsgBlock().Header.PrevBlock.IsEqual(goalPrevHash)})
			if !block.MsgBlock().Header.PrevBlock.IsEqual(goalPrevHash) {
				(*skipped)[block.MsgBlock().Header.PrevBlock] = *block

				// check if last_block.is_some() condition is correctly replaced with checkBlock()
				if lastBlock.CheckBlock() && block.MsgBlock().Header.PrevBlock == lastBlock.MsgBlock().Header.PrevBlock {
					logger.Debug("Blockchain", fmt.Sprintf("Chain split detected: %v <-> %v. Detecting main chain and orphan.", lastBlock.Hash().String(), block.Hash().String()), logger.Params{})

					firstOrphan := lastBlock
					secondOrphan := block

					for {
						block, err := blocks.Parse(slice)
						if err != nil {
							if len(*slice) != 0 {
								logger.Panic("Blockchain", errors.New("Block not found but the slice of blocks is not empty"), logger.Params{})
							}
							break
						}

						(*skipped)[block.MsgBlock().Header.PrevBlock] = *block
						if block.MsgBlock().Header.PrevBlock == *firstOrphan.Hash() {
							// First wins
							logger.Debug("Blockchain", fmt.Sprintf("Chain split: %v is on the main chain!", firstOrphan.Hash().String()), logger.Params{})
							break
						}
						if block.MsgBlock().Header.PrevBlock == *secondOrphan.Hash() {
							// Second wins
							logger.Debug("Blockchain", fmt.Sprintf("Chain split: %v is on the main chain!", secondOrphan.Hash().String()), logger.Params{})
							goalPrevHash = secondOrphan.Hash()
							*lastBlock = *secondOrphan
							break
						}
					}
				}
				continue
			}

			if lastBlock.CheckBlock() {
				BlockWalk(lastBlock, v, height, utxoSet)
				logger.Debug("Blockchain", fmt.Sprintf("(last_block) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
				(*height)++
			}

			goalPrevHash = block.Hash()
			*lastBlock = *block
		}
	}
}

func findCheckPoint(chain *[][]uint8, prevHeight, height *int32) error {
	for k, slice := range *chain {
		for len(slice) > 0 {
			_, err := blocks.Parse(&slice)
			if err != nil {
				return err
			}
			if *prevHeight == *height {
				(*height)++
				(*chain)[k] = slice
				return nil
			}
			(*prevHeight)++
		}
		(*chain)[k] = slice
	}
	return nil
}
