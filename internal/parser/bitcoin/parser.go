package bitcoin

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/utxoset"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// Parser defines the objects involved in the parsing of Bitcoin blockchain
// The involved objects include the parsed structure, the kind of parser, storage instances
// and some channel to manage the state of the parsing session
type Parser struct {
	blockchain *Blockchain
	client     *rpcclient.Client
	db         storage.DB
	skipped    *Skipped
	utxoset    *utxoset.UtxoSet
	cache      *cache.Cache
	interrupt  chan int
	done       chan int
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *Blockchain, client *rpcclient.Client, db storage.DB, skipped *Skipped, utxoset *utxoset.UtxoSet, c *cache.Cache, interrupt chan int, done chan int) Parser {
	return Parser{
		blockchain: blockchain,
		client:     client,
		db:         db,
		skipped:    skipped,
		utxoset:    utxoset,
		cache:      c,
		interrupt:  interrupt,
		done:       done,
	}
}

// Walk goes through the blockchain block by block
func (p *Parser) Walk(skipped int) (s int, err error) {
	goalPrevHash, _ := chainhash.NewHash(make([]byte, 32))
	var lastBlock *Block
	height, err := p.blockchain.Height()
	if err != nil {
		return
	}

	var rawChain [][]uint8
	for _, ref := range p.blockchain.Maps {
		rawChain = append(rawChain, []uint8(ref))
	}
	logger.Debug("Blockchain", "Files converted to be parsed: "+fmt.Sprintf("%v", len(rawChain)), logger.Params{})

	if height > 0 {
		last, e := p.blockchain.Head()
		if e != nil {
			err = e
			return
		}

		logger.Debug("Blockchain", "reaching endpoint to start from "+Itoa(height)+" - "+last.ID, logger.Params{})
		lastBlock, err = p.findCheckPointByHash(rawChain, &last)
		if err != nil {
			return
		}
		fmt.Println("RESTORED", p.skipped.Len(), last.ID, lastBlock.MsgBlock().Header.PrevBlock.String())

		height++
		goalPrevHash = lastBlock.Hash()

		logger.Info("Blockchain", "Start syncing from block "+Itoa(height), logger.Params{"hash": lastBlock.Hash().String()})
	}

	for k, ref := range rawChain {
		if len(ref) == 0 {
			continue
		}
		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": Itoa(int32(k)) + "/" + Itoa(int32(len(p.blockchain.Maps)-1)), "height": Itoa(height), "lastBlock": goalPrevHash.String()})
		if goalPrevHash, lastBlock, s, err = WalkSlice(p, &ref, goalPrevHash, lastBlock, &height, skipped); err != nil {
			return
		}
	}

	if viper.GetBool("realtime") {
		logger.Info("Blockchain", "Waiting for new blocks", logger.Params{"height": height})
		for {
			time.Sleep(2 * time.Second)
		}
	}

	return
}

// WalkSlice goes through a slice (block) of the chain
func WalkSlice(p *Parser, slice *[]uint8, g *chainhash.Hash, l *Block, height *int32, skipped int) (goalPrevHash *chainhash.Hash, lastBlock *Block, s int, err error) {
	goalPrevHash = g
	lastBlock = l
	for len(*slice) > 0 {
		select {
		case x, ok := <-p.interrupt:
			if !ok {
				err = errors.New("Something wrong in interrupt signal")
			}
			logger.Info("Blockchain", "Received interrupt signal", logger.Params{"signal": x})
			return

		default:
			if _, e := p.skipped.GetBlock(goalPrevHash); e == nil {
				logger.Debug("Blockchain", "(rewind - pre-step) Block "+Itoa(*height)+" - "+lastBlock.MsgBlock().Header.PrevBlock.String()+" -> "+lastBlock.Hash().String(), logger.Params{})
				if err = BlockWalk(p, lastBlock, height); err != nil {
					return
				}
				(*height)++
				for {
					if block, e := p.skipped.GetBlock(goalPrevHash); e == nil {
						p.skipped.DeleteBlock(goalPrevHash)
						logger.Debug("Blockchain", "(rewind) Block "+Itoa(*height)+" - "+block.MsgBlock().Header.PrevBlock.String()+" -> "+block.Hash().String(), logger.Params{})
						if err = BlockWalk(p, &block, height); err != nil {
							return
						}
						(*height)++
						goalPrevHash = block.Hash()
						lastBlock = nil
						continue
					}
					break
				}
			}

			block, e := Parse(slice)
			if e != nil {
				if len(*slice) != 0 {
					err = e
					return
				}
				break
			}

			logger.Debug("Blockchain", "Block candidate for height "+Itoa(*height)+" - goal_prev_hash = "+goalPrevHash.String()+", prev_hash = "+block.MsgBlock().Header.PrevBlock.String()+", cur_hash = "+block.Hash().String(), logger.Params{})

			// Explanation: parsing the dat files means find a not ordinate sequence of  In most cases parsing the next block means
			// find a block that it's been added to blockchain many blocks after, so at a higher height. This means that, that block, will
			// be necessaire later when the parsing will reach the preceding block. At that point you will need to have the already parsed
			// block. This is why the skipped slice is built, is where we keep the unordinate blocks already parsed. If we stop the parsing
			// process and restart it, we need to restore the skipped block slice too, because otherwire we wouldn't have all the blocks
			// needed to complete the chain.
			if !block.MsgBlock().Header.PrevBlock.IsEqual(goalPrevHash) {
				logger.Debug("Blockchain", "Skipped block", logger.Params{"prev": block.MsgBlock().Header.PrevBlock.String()})
				p.skipped.StoreBlockPrevHash(block)
				if p.skipped.Len() > skipped {
					err = errors.New("too many skipped blocks, stopping process")
					s = p.skipped.Len()
					return
				}

				// check if last_block.is_some() condition is correctly replaced with checkBlock()
				if lastBlock.CheckBlock() && block.MsgBlock().Header.PrevBlock.String() == lastBlock.MsgBlock().Header.PrevBlock.String() {
					logger.Debug("Blockchain", "Chain split detected: "+lastBlock.Hash().String()+"% <-> "+block.Hash().String()+". Detecting main chain and orphan.", logger.Params{})

					firstOrphan := lastBlock
					secondOrphan := block

					for {
						block, e := Parse(slice)
						if err != nil {
							if len(*slice) != 0 {
								err = e
								return
							}
							break
						}

						p.skipped.StoreBlockPrevHash(block)
						if block.MsgBlock().Header.PrevBlock.IsEqual(firstOrphan.Hash()) {
							// First wins
							logger.Debug("Blockchain", "Chain split: "+firstOrphan.Hash().String()+" is on the main chain!", logger.Params{})
							break
						}
						if block.MsgBlock().Header.PrevBlock.IsEqual(secondOrphan.Hash()) {
							// Second wins
							logger.Debug("Blockchain", "Chain split: "+secondOrphan.Hash().String()+" is on the main chain!", logger.Params{})
							goalPrevHash = secondOrphan.Hash()
							*lastBlock = *secondOrphan
							break
						}
					}
				}
				continue
			}

			if lastBlock.CheckBlock() {
				logger.Debug("Blockchain", "(last_block) Parsing block "+Itoa(*height)+" - "+lastBlock.MsgBlock().Header.PrevBlock.String()+" -> "+lastBlock.Hash().String(), logger.Params{})
				if err = BlockWalk(p, lastBlock, height); err != nil {
					return
				}
				(*height)++
			}

			logger.Debug("Blockchain", "(next_block) Updating block "+Itoa(*height)+": "+block.Hash().String(), logger.Params{})
			goalPrevHash = block.Hash()
			lastBlock = block
		}
	}

	return
}

func (p *Parser) findCheckPointByHash(chain [][]uint8, last *models.Block) (block *Block, err error) {
	step := int32(viper.GetInt("restoredBlocks"))
	from := last.Height - step
	if from < 0 {
		from = 0
	}
	list, err := p.db.GetStoredBlocksList(from)
	if err != nil {
		return
	}
	logger.Info("Blockchain", "Blocks list fetched", logger.Params{"length": len(list)})

	file := viper.GetInt("file")
	logger.Info("Blockchain", "Files already parsed", logger.Params{"number": file})

	for k, slice := range chain {
		if k < file {
			chain[k] = []uint8{}
			continue
		}

		for len(slice) > 0 {
			block, err = Parse(&slice)
			if err != nil {
				return
			}

			if block.MsgBlock().Header.PrevBlock.String() == last.ID {
				chain[k] = slice
				return
			}
			if _, stored := list[block.Hash().String()]; !stored {
				p.skipped.StoreBlockPrevHash(block)
			}
		}
		chain[k] = slice
	}
	err = errors.New("Checkpoint not found")
	return
}

// Itoa utility function to convert height (int32) to string to easily print it
func Itoa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}
