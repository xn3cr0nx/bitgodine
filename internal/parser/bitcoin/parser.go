package bitcoin

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
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
	db         kv.DB
	skipped    *Skipped
	utxoset    *utxoset.UtxoSet
	cache      *cache.Cache
	interrupt  chan int
}

// CheckPoint represents the last parse state
type CheckPoint struct {
	height       int32
	goalPrevHash *chainhash.Hash
	lastBlock    *Block
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *Blockchain, client *rpcclient.Client, db kv.DB, skipped *Skipped, utxoset *utxoset.UtxoSet, c *cache.Cache, interrupt chan int) Parser {
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
		file, e := GetFileParsed(p.db)
		if e != nil {
			return e
		}

		if err = p.blockchain.Read("", file); err != nil {
			return
		}

		if err := p.Parse(); err != nil {
			if errors.Is(err, ErrInterrupt) {
				break
			}
			return err
		}
	}
	return
}

// Parse goes through the blockchain block by block
func (p *Parser) Parse() (err error) {
	var rawChain [][]uint8
	for _, ref := range p.blockchain.Maps {
		rawChain = append(rawChain, []uint8(ref))
	}
	logger.Debug("Blockchain", "Files converted to be parsed: "+fmt.Sprintf("%v", len(rawChain)), logger.Params{})

	check, err := p.FindCheckPoint(rawChain)
	if err != nil {
		return
	}
	logger.Info("Blockchain", "Start syncing from block "+Itoa(check.height), logger.Params{})

	for k, file := range rawChain {
		if len(file) == 0 {
			continue
		}
		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": Itoa(int32(k)) + "/" + Itoa(int32(len(p.blockchain.Maps)-1)), "height": Itoa(check.height), "lastBlock": check.goalPrevHash.String()})
		// if check, err = ParseFile(p, &file, check); err != nil {
		if check, err = ParseFile(p, check, &file); err != nil {
			return
		}

		if err = StoreFileParsed(p.db, k); err != nil {
			return
		}

		if err = p.blockchain.Maps[k].Unmap(); err != nil {
			return
		}
	}

	return
}

// ParseFile walks through the raw file and extract blocks
func ParseFile(p *Parser, c CheckPoint, file *[]uint8) (check CheckPoint, err error) {
	// performance improvement to reduce pointers allocation
	check = c
	for len(*file) > 0 {
		select {
		case x, ok := <-p.interrupt:
			if !ok {
				err = ErrInterruptUnknown
			}
			logger.Info("Blockchain", "Received interrupt signal", logger.Params{"signal": x})
			err = ErrInterrupt
			return

		default:
			if _, e := p.skipped.GetBlock(check.goalPrevHash); e == nil {
				logger.Debug("Blockchain", "(rewind - pre-step) Block "+Itoa(check.height)+" - "+check.lastBlock.MsgBlock().Header.PrevBlock.String()+" -> "+check.lastBlock.Hash().String(), logger.Params{})
				if err = check.lastBlock.Store(p.db, check.height); err != nil {
					return
				}
				check.height++
				for {
					if block, e := p.skipped.GetBlock(check.goalPrevHash); e == nil {
						p.skipped.DeleteBlock(check.goalPrevHash)
						logger.Debug("Blockchain", "(rewind) Block "+Itoa(check.height)+" - "+block.MsgBlock().Header.PrevBlock.String()+" -> "+block.Hash().String(), logger.Params{})
						if err = block.Store(p.db, check.height); err != nil {
							return
						}
						check.height++
						check.goalPrevHash = block.Hash()
						check.lastBlock = nil
						continue
					}
					break
				}
			}

			block, e := ExtractBlockFromFile(file)
			if e != nil {
				if !errors.Is(e, ErrEmptySliceParse) {
					err = e
					return
				}
				break
			}

			logger.Debug("Blockchain", "Block candidate for height "+Itoa(check.height)+" - goal_prev_hash = "+check.goalPrevHash.String()+", prev_hash = "+block.MsgBlock().Header.PrevBlock.String()+", cur_hash = "+block.Hash().String(), logger.Params{})

			// Explanation: parsing the dat files means find a not ordinate sequence of  In most cases parsing the next block means
			// find a block that it's been added to blockchain many blocks after, so at a higher height. This means that, that block, will
			// be necessaire later when the parsing will reach the preceding block. At that point you will need to have the already parsed
			// block. This is why the skipped slice is built, is where we keep the unordinate blocks already parsed. If we stop the parsing
			// process and restart it, we need to restore the skipped block slice too, because otherwire we wouldn't have all the blocks
			// needed to complete the chain.
			if !block.MsgBlock().Header.PrevBlock.IsEqual(check.goalPrevHash) {
				logger.Debug("Blockchain", "Skipped block", logger.Params{"prev": block.MsgBlock().Header.PrevBlock.String()})
				p.skipped.StoreBlockPrevHash(block)

				// check if last_block.is_some() condition is correctly replaced with checkBlock()
				if check.lastBlock.CheckBlock() && block.MsgBlock().Header.PrevBlock.String() == check.lastBlock.MsgBlock().Header.PrevBlock.String() {
					logger.Debug("Blockchain", "Chain split detected: "+check.lastBlock.Hash().String()+"% <-> "+block.Hash().String()+". Detecting main chain and orphan.", logger.Params{})

					firstOrphan := check.lastBlock
					secondOrphan := block

					for {
						block, e := ExtractBlockFromFile(file)
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
							check.goalPrevHash = secondOrphan.Hash()
							check.lastBlock = secondOrphan
							break
						}
					}
				}
				continue
			}

			if check.lastBlock.CheckBlock() {
				logger.Debug("Blockchain", "(last_block) Parsing block "+Itoa(check.height)+" - "+check.lastBlock.MsgBlock().Header.PrevBlock.String()+" -> "+check.lastBlock.Hash().String(), logger.Params{})
				if err = check.lastBlock.Store(p.db, check.height); err != nil {
					return
				}
				check.height++
			}

			logger.Debug("Blockchain", "(next_block) Updating block "+Itoa(check.height)+": "+block.Hash().String(), logger.Params{})
			check.goalPrevHash = block.Hash()
			check.lastBlock = block
		}
	}

	return
}

// FindCheckPoint restores the parsed files' state from last parsing and return a CheckPoint instance the keep parsing
func (p *Parser) FindCheckPoint(rawChain [][]uint8) (check CheckPoint, err error) {
	check = CheckPoint{}
	check.goalPrevHash, _ = chainhash.NewHash(make([]byte, 32))

	check.height, err = p.blockchain.Height()
	if err != nil {
		return
	}

	if check.height > 0 {
		last, e := p.blockchain.Head()
		if e != nil {
			err = e
			return
		}

		logger.Debug("Blockchain", "reaching endpoint to start from "+Itoa(check.height)+" - "+last.ID, logger.Params{})
		check.lastBlock, err = restoreFileState(p, rawChain, &last)
		if err != nil {
			return
		}
		fmt.Println("Restored", p.skipped.Len(), last.ID, check.lastBlock.MsgBlock().Header.PrevBlock.String())

		check.height++
		check.goalPrevHash = check.lastBlock.Hash()
	}

	return
}

func restoreFileState(p *Parser, chain [][]uint8, last *block.Block) (b *Block, err error) {
	step := int32(viper.GetInt("restoredBlocks"))
	from := last.Height - step
	if from < 0 {
		from = 0
	}
	list, err := block.NewService(p.db, nil).GetStoredList(from)
	if err != nil {
		return
	}
	logger.Info("Blockchain", "Blocks list fetched", logger.Params{"length": len(list)})

	for k, file := range chain {
		for len(file) > 0 {
			b, err = ExtractBlockFromFile(&file)
			if err != nil {
				return
			}

			if b.MsgBlock().Header.PrevBlock.String() == last.ID {
				chain[k] = file
				return
			}
			if _, stored := list[b.Hash().String()]; !stored {
				p.skipped.StoreBlockPrevHash(b)
			}
		}
		chain[k] = file
	}
	err = ErrCheckpointNotFound
	return
}

// Itoa utility function to convert height (int32) to string to easily print it
func Itoa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}
