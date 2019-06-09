package bitcoin

import (
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/dbblocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

// Parser defines the objects involved in the parsing of Bitcoin blockchain
// The involved objects include the parsed structure, the kind of parser, storage instances
// and some channel to manage the state of the parsing session
type Parser struct {
	blockchain *blockchain.Blockchain
	visitor    visitor.BlockchainVisitor
	dbblocks   *dbblocks.DbBlocks
	interrupt  chan int
	done       chan int
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *blockchain.Blockchain, visitor visitor.BlockchainVisitor, dbblocks *dbblocks.DbBlocks, interrupt chan int, done chan int) *Parser {
	return &Parser{
		blockchain: blockchain,
		visitor:    visitor,
		dbblocks:   dbblocks,
		interrupt:  interrupt,
		done:       done,
	}
}
