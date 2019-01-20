package blockchain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	mmap "github.com/edsrzf/mmap-go"
	dir "github.com/mitchellh/go-homedir"
)

type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
}

var blockchain *Blockchain

func Instance(network chaincfg.Params) *Blockchain {
	if blockchain == nil {
		blockchain = new(Blockchain)
		blockchain.Network = network
	}
	return blockchain
}

func (b *Blockchain) Read() error {
	var Maps []mmap.MMap
	netPath := b.Network.Name
	n := 0
	blocksDir, _ := dir.Dir()
	if b.Network.Name == "mainnet" {
		netPath = ""
	}
	blocksDir = filepath.Join(blocksDir, ".bitcoin", netPath, "blocks")
	fmt.Printf("%v\n", blocksDir)

	for {
		f, err := os.OpenFile(filepath.Join(blocksDir, fmt.Sprintf("blk%05d.dat", n)), os.O_RDWR, 0644)
		defer f.Close()
		if err != nil {
			break
		}

		n = n + 1
		m, err := mmap.Map(f, 2, 0)
		if err != nil {
			logger.Panic("Mapping file to memory", err, logger.Params{})
			return err
		}
		Maps = append(Maps, m)
	}

	b.Maps = Maps
	return nil
}

func (b *Blockchain) Walk(v visitor.BlockchainVisitor) (uint64, *chainhash.Hash, map[chainhash.Hash][]visitor.OutputItem) {
	skipped := make(map[chainhash.Hash]btcutil.Block)
	outputItems := make(map[chainhash.Hash][]visitor.OutputItem)
	goalPrevHash, _ := chainhash.NewHash(make([]byte, 32))
	var lastBlock btcutil.Block
	height := uint64(0)

	for k, value := range b.Maps {
		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": fmt.Sprintf("%v/%v", k, len(b.Maps)-1)})
		val := []uint8(value) // I need to cast mmap.Map initializing a variable to be able to take its address (&)
		b.WalkSlice(&val, goalPrevHash, &lastBlock, &height, &skipped, &outputItems, &v)
	}

	return height, goalPrevHash, outputItems
}

func (b *Blockchain) WalkSlice(slice *[]uint8, goalPrevHash *chainhash.Hash, lastBlock *btcutil.Block, height *uint64, skipped *map[chainhash.Hash]btcutil.Block, outputItems *map[chainhash.Hash][]visitor.OutputItem, v *visitor.BlockchainVisitor) {
	for len(*slice) > 0 {
		if _, ok := (*skipped)[*goalPrevHash]; ok {
			blocks.Walk(lastBlock, v, height, outputItems)
			logger.Info("Blockchain", fmt.Sprintf("(rewind - pre-step) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
			*height++
			// Here I should do the for loop removing every goal_prev_hash and
			// walking the block obtained at the index of goal_prev_hash
			for {
				if block, ok := (*skipped)[*goalPrevHash]; ok {
					delete(*skipped, *goalPrevHash)
					blocks.Walk(&block, v, height, outputItems)
					logger.Info("Blockchain", fmt.Sprintf("(rewind) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
					*height++
					*goalPrevHash = *block.Hash()
					// possible bug initialization to null (None in rust)
					*lastBlock = btcutil.Block{}
					continue
				}
				break
			}
		}

		block, err := blocks.Read(slice)
		if err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			if len(*slice) != 0 {
				logger.Panic("Blockchain", errors.New("Block not found but the slice of blocks is not empty"), logger.Params{})
			}
			break
		}

		logger.Info("Blockchain", fmt.Sprintf("Block candidate for height %d - goal_prev_hash = %v, prev_hash = %v, cur_hash = %v", *height, goalPrevHash.String(), block.MsgBlock().Header.PrevBlock.String(), block.Hash().String()), logger.Params{})

		if !block.MsgBlock().Header.PrevBlock.IsEqual(goalPrevHash) {
			(*skipped)[block.MsgBlock().Header.PrevBlock] = *block

			// check if last_block.is_some() condition is correctly replaced with checkBlock()
			if blocks.CheckBlock(lastBlock) && block.MsgBlock().Header.PrevBlock == lastBlock.MsgBlock().Header.PrevBlock {
				logger.Info("Blockchain", fmt.Sprintf("Chain split detected: %v <-> %v. Detecting main chain and orphan.", lastBlock.Hash().String(), block.Hash().String()), logger.Params{})

				firstOrphan := lastBlock
				secondOrphan := block

				for {
					block, err := blocks.Read(slice)
					if err != nil {
						logger.Error("Blockchain", err, logger.Params{})
						if len(*slice) != 0 {
							logger.Panic("Blockchain", errors.New("Block not found but the slice of blocks is not empty"), logger.Params{})
						}
						break
					}

					(*skipped)[block.MsgBlock().Header.PrevBlock] = *block
					if block.MsgBlock().Header.PrevBlock == *firstOrphan.Hash() {
						// First wins
						logger.Info("Blockchain", fmt.Sprintf("Chain split: %v is on the main chain!", firstOrphan.Hash().String()), logger.Params{})
						break
					}
					if block.MsgBlock().Header.PrevBlock == *secondOrphan.Hash() {
						// Second wins
						logger.Info("Blockchain", fmt.Sprintf("Chain split: %v is on the main chain!", secondOrphan.Hash().String()), logger.Params{})
						goalPrevHash = secondOrphan.Hash()
						*lastBlock = *secondOrphan
						break
					}
				}
			}
			continue
		}

		// TODO: Here add too the check to be sure the block is correct
		if blocks.CheckBlock(lastBlock) {
			blocks.Walk(lastBlock, v, height, outputItems)
			logger.Info("Blockchain", fmt.Sprintf("(last_block) Block %v - %v -> %v", *height, lastBlock.MsgBlock().Header.PrevBlock.String(), lastBlock.Hash().String()), logger.Params{})
			*height++
		}

		goalPrevHash = block.Hash()
		*lastBlock = *block
	}
}
