package parser

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
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
	// prevHeight := int32(0)
	height := bc.Height()

	// check if "coinbase output" is stored in dgraph
	hash := strings.Repeat("0", 64)
	if _, err := dgraph.GetTxUID(&hash); err != nil {
		logger.Debug("Blockchain", "missing coinbase outputs", logger.Params{"hash": hash})
		if err := dgraph.StoreCoinbase(); err != nil {
			logger.Panic("Blockchain", err, logger.Params{})
		}
	}

	var rawChain [][]uint8
	for _, ref := range bc.Maps {
		rawChain = append(rawChain, []uint8(ref))
	}
	logger.Debug("Blockchain", fmt.Sprintf("Files converted to be parsed: %d", len(rawChain)), logger.Params{})

	if height > 0 {
		logger.Debug("Blockchain", fmt.Sprintf("reaching endpoint to start from %d", height), logger.Params{})
		last, err := bc.Head()
		if err != nil {
			logger.Panic("Blockchain", err, logger.Params{})
		}
		if err := restoreSkipped(&skipped); err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			os.Exit(-1)
		}
		lastBlock, err := findCheckPointByHash(&rawChain, last.Hash())
		if err != nil {
			logger.Panic("Blockchain", err, logger.Params{})
		}
		height++
		goalPrevHash = lastBlock.Hash()
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
				logger.Debug("Blockchain", fmt.Sprintf("(rewind - pre-step) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
				BlockWalk(lastBlock, v, height, utxoSet)
				(*height)++
				// Here I should do the for loop removing every goal_prev_hash and
				// walking the block obtained at the index of goal_prev_hash
				for {
					if block, ok := (*skipped)[*goalPrevHash]; ok {
						delete(*skipped, *goalPrevHash)
						if err := db.DeleteBlock(goalPrevHash); err != nil {
							logger.Error("Blockchain", err, logger.Params{})
							return
						}
						logger.Debug("Blockchain", fmt.Sprintf("(rewind) Block %v - %v -> %v", *height, block.MsgBlock().Header.PrevBlock.String(), block.Hash().String()), logger.Params{})
						BlockWalk(&block, v, height, utxoSet)
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

			// Explanation: parsing the dat files means find a not ordinate sequence of blocks. In most cases parsing the next block means
			// find a block that it's been added to blockchain many blocks after, so at a higher height. This means that, that block, will
			// be necessaire later when the parsing will reach the preceding block. At that point you will need to have the already parsed
			// block. This is why the skipped slice is built, is where we keep the unordinate blocks already parsed. If we stop the parsing
			// process and restart it, we need to restore the skipped block slice too, because otherwire we wouldn't have all the blocks
			// needed to complete the chain.
			if !block.MsgBlock().Header.PrevBlock.IsEqual(goalPrevHash) {
				logger.Debug("Blockchain", "Skipped block", logger.Params{"prev": block.MsgBlock().Header.PrevBlock.String()})
				(*skipped)[block.MsgBlock().Header.PrevBlock] = *block
				if err := db.StoreBlockPrevHash(block); err != nil {
					logger.Panic("Blockchain", err, logger.Params{})
				}

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
						if err := db.StoreBlockPrevHash(block); err != nil {
							logger.Panic("Blockchain", err, logger.Params{})
						}
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
				logger.Debug("Blockchain", fmt.Sprintf("(last_block) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
				BlockWalk(lastBlock, v, height, utxoSet)
				(*height)++
			}

			logger.Debug("Blockchain", fmt.Sprintf("(next_block) Updating block %v: %v", *height, block.Hash().String()), logger.Params{})
			goalPrevHash = block.Hash()
			*lastBlock = *block
		}
	}
}

func findCheckPointByHash(chain *[][]uint8, hash *chainhash.Hash) (blocks.Block, error) {
	for k, slice := range *chain {
		for len(slice) > 0 {
			block, err := blocks.Parse(&slice)
			if err != nil {
				return blocks.Block{}, err
			}
			if block.Hash().IsEqual(hash) {
				(*chain)[k] = slice
				return *block, nil
			}
		}
		(*chain)[k] = slice
	}
	return blocks.Block{}, nil
}

func restoreSkipped(skipped *map[chainhash.Hash]blocks.Block) error {
	cachedSkipped, err := db.GetAll()
	if err != nil {
		return err
	}
	logger.Info("Blockchain", "Restoring skipped blocks", logger.Params{"n_blocks": len(cachedSkipped)})
	for _, skip := range cachedSkipped {
		(*skipped)[skip.MsgBlock().Header.PrevBlock] = skip
	}
	return nil
}
