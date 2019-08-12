package bitcoin

import (
	"github.com/allegro/bigcache"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/badger/skipped"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

// Parser defines the objects involved in the parsing of Bitcoin blockchain
// The involved objects include the parsed structure, the kind of parser, storage instances
// and some channel to manage the state of the parsing session
type Parser struct {
	blockchain *blockchain.Blockchain
	visitor    visitor.BlockchainVisitor
	skipped    *skipped.Skipped
	cache      *bigcache.BigCache
	interrupt  chan int
	done       chan int
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *blockchain.Blockchain, visitor visitor.BlockchainVisitor, skipped *skipped.Skipped, cache *bigcache.BigCache, interrupt chan int, done chan int) *Parser {
	return &Parser{
		blockchain: blockchain,
		visitor:    visitor,
		skipped:    skipped,
		cache:      cache,
		interrupt:  interrupt,
		done:       done,
	}
}
