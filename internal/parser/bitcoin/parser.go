package bitcoin

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/utxoset"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
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
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *Blockchain, client *rpcclient.Client, db storage.DB, skipped *Skipped, utxoset *utxoset.UtxoSet, c *cache.Cache, interrupt chan int) Parser {
	return Parser{
		blockchain: blockchain,
		client:     client,
		db:         db,
		skipped:    skipped,
		utxoset:    utxoset,
		cache:      c,
		interrupt:  interrupt,
	}
}

func handleInterrupt(c chan os.Signal, interrupt chan int) {
	for sig := range c {
		logger.Info("Sync", "Killing the application", logger.Params{"signal": sig})
		interrupt <- 1
	}
}

// InfinitelyParse parses the blockchain starting from scratch when it reaches the end in order to implement a real time mechanism
func (p *Parser) InfinitelyParse() (err error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go handleInterrupt(ch, p.interrupt)

	for {
		if err = p.blockchain.Read(""); err != nil {
			return
		}

		if err := p.Parse(); err != nil {
			if errors.Is(err, ErrInterrupt) {
				break
			}
			// if errors.Is(err, ErrExceededSize) {
			// 	p.skipped.Empty()
			// 	continue
			// }
			return err
		}
	}
	return
}

// Parse goes through the blockchain block by block
func (p *Parser) Parse() (err error) {
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
		fmt.Println("Restored", p.skipped.Len(), last.ID, lastBlock.MsgBlock().Header.PrevBlock.String())

		height++
		goalPrevHash = lastBlock.Hash()

		logger.Info("Blockchain", "Start syncing from block "+Itoa(height), logger.Params{"hash": lastBlock.Hash().String()})
	}

	for k, ref := range rawChain {
		if len(ref) == 0 {
			continue
		}
		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": Itoa(int32(k)) + "/" + Itoa(int32(len(p.blockchain.Maps)-1)), "height": Itoa(height), "lastBlock": goalPrevHash.String()})
		if goalPrevHash, lastBlock, err = ParseSlice(p, &ref, goalPrevHash, lastBlock, &height); err != nil {
			return
		}
	}

	if viper.GetBool("realtime") {
		logger.Info("Blockchain", "Waiting for new blocks", logger.Params{"height": height})
		time.Sleep(2 * time.Second)
	}

	return
}

// ParseSlice goes through a slice (block) of the chain
func ParseSlice(p *Parser, slice *[]uint8, g *chainhash.Hash, l *Block, height *int32) (goalPrevHash *chainhash.Hash, lastBlock *Block, err error) {
	goalPrevHash = g
	lastBlock = l
	for len(*slice) > 0 {
		select {
		case x, ok := <-p.interrupt:
			if !ok {
				err = ErrInterruptUnknown
			}
			logger.Info("Blockchain", "Received interrupt signal", logger.Params{"signal": x})
			err = ErrInterrupt
			return

		default:
			if _, e := p.skipped.GetBlock(goalPrevHash); e == nil {
				logger.Debug("Blockchain", "(rewind - pre-step) Block "+Itoa(*height)+" - "+lastBlock.MsgBlock().Header.PrevBlock.String()+" -> "+lastBlock.Hash().String(), logger.Params{})
				if err = lastBlock.Store(p.db, height); err != nil {
					return
				}
				(*height)++
				for {
					if block, e := p.skipped.GetBlock(goalPrevHash); e == nil {
						p.skipped.DeleteBlock(goalPrevHash)
						logger.Debug("Blockchain", "(rewind) Block "+Itoa(*height)+" - "+block.MsgBlock().Header.PrevBlock.String()+" -> "+block.Hash().String(), logger.Params{})
						if err = block.Store(p.db, height); err != nil {
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

			block, e := ExtractBlockFromSlice(slice)
			if e != nil {
				if !errors.Is(e, ErrEmptySliceParse) {
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
				// if p.skipped.Len() > skipped {
				// 	err = fmt.Errorf("%w: %d", ErrExceededSize, p.skipped.Len())
				// 	return
				// }

				// check if last_block.is_some() condition is correctly replaced with checkBlock()
				if lastBlock.CheckBlock() && block.MsgBlock().Header.PrevBlock.String() == lastBlock.MsgBlock().Header.PrevBlock.String() {
					logger.Debug("Blockchain", "Chain split detected: "+lastBlock.Hash().String()+"% <-> "+block.Hash().String()+". Detecting main chain and orphan.", logger.Params{})

					firstOrphan := lastBlock
					secondOrphan := block

					for {
						block, e := ExtractBlockFromSlice(slice)
						if err != nil {
							if !errors.Is(e, ErrEmptySliceParse) {
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
				if err = lastBlock.Store(p.db, height); err != nil {
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

func (p *Parser) findCheckPointByHash(chain [][]uint8, last *block.Block) (b *Block, err error) {
	step := int32(viper.GetInt("restoredBlocks"))
	from := last.Height - step
	if from < 0 {
		from = 0
	}
	list, err := block.GetStoredList(p.db, from)
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
			b, err = ExtractBlockFromSlice(&slice)
			if err != nil {
				return
			}

			if b.MsgBlock().Header.PrevBlock.String() == last.ID {
				chain[k] = slice
				return
			}
			if _, stored := list[b.Hash().String()]; !stored {
				p.skipped.StoreBlockPrevHash(b)
			}
		}
		chain[k] = slice
	}
	err = ErrCheckpointNotFound
	return
}

// Itoa utility function to convert height (int32) to string to easily print it
func Itoa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}
