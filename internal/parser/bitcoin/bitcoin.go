package bitcoin

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/xn3cr0nx/bitgodine/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine/internal/utxoset"

	"github.com/xn3cr0nx/bitgodine/internal/skipped"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// Parser defines the objects involved in the parsing of Bitcoin blockchain
// The involved objects include the parsed structure, the kind of parser, storage instances
// and some channel to manage the state of the parsing session
type Parser struct {
	blockchain *blockchain.Blockchain
	client     *rpcclient.Client
	db         storage.DB
	skipped    *skipped.Skipped
	utxoset    *utxoset.UtxoSet
	cache      *cache.Cache
	interrupt  chan int
	done       chan int
}

// NewParser return a new instance to Bitcoin blockchai parser
func NewParser(blockchain *blockchain.Blockchain, client *rpcclient.Client, db storage.DB, skipped *skipped.Skipped, utxoset *utxoset.UtxoSet, c *cache.Cache, interrupt chan int, done chan int) Parser {
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
