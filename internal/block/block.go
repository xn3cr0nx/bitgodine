package block

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Service interface exports available methods for tx service
type Service interface {
	StoreBlock(b *Block, txs []tx.Tx) (err error)
	ReadFromHeight(height int32) (block Block, err error)
	ReadHeight() (height int32, err error)
	GetFromHash(hash string) (Block, error)
	GetFromHeight(height int32) (*BlockOut, error)
	GetFromHashWithTxs(hash string) (*BlockOut, error)
	GetLast() (*BlockOut, error)
	GetFromHeightRange(height int32, first int) (blocks []BlockOut, err error)
	GetStored() (blocks []Block, err error)
	GetStoredList(from int32) (blocks map[string]interface{}, err error)
	Remove(block *Block) error
	RemoveLast() error
	GetStoredTxs() (transactions []string, err error)
	GetTxBlock(hash string) (block *BlockOut, err error)
	GetTxBlockHeight(hash string) (height int32, err error)
}

type service struct {
	Kv    kv.DB
	Cache *cache.Cache
}

// NewService instantiates a new Service layer for customer
func NewService(k kv.DB, c *cache.Cache) *service {
	return &service{
		Kv:    k,
		Cache: c,
	}
}

func (s *service) fetchBlockTxs(txs []string) (transactions []tx.Tx, err error) {
	txService := tx.NewService(s.Kv, s.Cache)
	for _, hash := range txs {
		transaction, e := txService.GetFromHash(hash)
		if e != nil {
			return nil, e
		}
		transactions = append(transactions, transaction)
	}
	return
}

// StoreBlock inserts in the db the block as []byte passed
// for fast research purpose blocks have _ prefix, tx_ for txs prefix and h_ for height prefix
func (s *service) StoreBlock(b *Block, txs []tx.Tx) (err error) {
	// b := v.(*Block)
	// txs := t.([]tx.Tx)
	blockHash := []byte(b.ID)
	h := strconv.Itoa(int(b.Height))

	batch := make(map[string][]byte)
	serialized, err := encoding.Marshal(b)
	if err != nil {
		return
	}
	batch[b.ID] = serialized
	batch[h] = blockHash
	batch["last"] = []byte(h)

	for _, tx := range txs {
		serialized, e := encoding.Marshal(tx)
		if e != nil {
			err = e
			return
		}
		batch[tx.TxID] = serialized
		batch["_"+tx.TxID] = []byte(h)
		for _, o := range tx.Vout {
			batch[o.ScriptpubkeyAddress+"_"+tx.TxID] = []byte(h)
		}
		for _, i := range tx.Vin {
			batch[i.TxID+"_"+fmt.Sprint(i.Vout)] = []byte(tx.TxID)
		}
	}

	err = s.Kv.StoreQueueBatch(batch)
	return
}

// read retrieves tx by hash
func read(db kv.DB, hash string) (block Block, err error) {
	r, err := db.Read(hash)
	if err != nil {
		return
	}
	if err = encoding.Unmarshal(r, &block); err != nil {
		return
	}
	return
}

// ReadFromHeight retrieves block by height
func (s *service) ReadFromHeight(height int32) (block Block, err error) {
	hash, err := s.Kv.Read(strconv.Itoa(int(height)))
	if err != nil {
		return
	}
	block, err = s.GetFromHash(string(hash))
	if err != nil {
		return
	}
	return
}

// ReadHeight returnes the hash of the block retrieving it based on its height
func (s *service) ReadHeight() (height int32, err error) {
	h, err := s.Kv.Read("last")
	if err != nil {
		if errors.Is(err, errorx.ErrKeyNotFound) {
			return 0, nil
		}
		return
	}
	conv, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height = int32(conv)
	return
}

// GetFromHash return block structure based on block hash
func (s *service) GetFromHash(hash string) (Block, error) {
	b, err := read(s.Kv, hash)
	if err != nil {
		return Block{}, err
	}

	return b, nil
}

// GetFromHeight return block structure based on block height
func (s *service) GetFromHeight(height int32) (*BlockOut, error) {
	b, err := s.ReadFromHeight(height)
	if err != nil {
		return nil, err
	}

	txs, err := s.fetchBlockTxs(b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetFromHashWithTxs return block structure based on block hash
func (s *service) GetFromHashWithTxs(hash string) (*BlockOut, error) {
	b, err := s.GetFromHash(hash)
	if err != nil {
		return nil, err
	}

	txs, err := s.fetchBlockTxs(b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetLast return last synced block
func (s *service) GetLast() (*BlockOut, error) {
	h, err := s.ReadHeight()
	if err != nil {
		return nil, err
	}
	b, err := s.ReadFromHeight(h)
	if err != nil {
		return nil, err
	}

	txs, err := s.fetchBlockTxs(b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetFromHeightRange returnes the hash of the block retrieving it based on its height
func (s *service) GetFromHeightRange(height int32, first int) (blocks []BlockOut, err error) {
	for i := height; i < (height + int32(first)); i++ {
		b, e := s.GetFromHeight(i)
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, *b)
	}
	return
}

// GetStored returnes list of all stored blocks
func (s *service) GetStored() (blocks []Block, err error) {
	h, err := s.ReadHeight()
	if err != nil {
		return
	}
	for i := int32(0); i <= h; i++ {
		hash, e := s.Kv.Read(strconv.Itoa(int(i)))
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, Block{ID: string(hash), Height: h})
	}
	return
}

// GetStoredList returnes all the stored block ids
func (s *service) GetStoredList(from int32) (blocks map[string]interface{}, err error) {
	// blocks, err = db.ReadKeysWithPrefix("_0000")
	h, err := s.ReadHeight()
	if err != nil {
		return
	}
	blocks = make(map[string]interface{}, h)
	for i := from; i <= h; i++ {
		hash, e := s.Kv.Read(strconv.Itoa(int(i)))
		if e != nil {
			err = e
			return
		}
		blocks[string(hash)] = nil
	}
	return
}

// Remove removes the block specified by its height
func (s *service) Remove(block *Block) error {
	return s.Kv.Delete(block.ID)
}

// RemoveLast removes the last block stored
func (s *service) RemoveLast() error {
	h, err := s.ReadHeight()
	if err != nil {
		return err
	}
	block, err := s.GetFromHeight(h)
	if err != nil {
		return err
	}
	return s.Kv.Delete(block.ID)
}

// GetStoredTxs returnes all the stored transactions hashes
func (s *service) GetStoredTxs() (transactions []string, err error) {
	blocks, err := s.GetStored()
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			transactions = append(transactions, tx)
		}
	}
	return
}

// GetTxBlock returnes the block containing the transaction
func (s *service) GetTxBlock(hash string) (block *BlockOut, err error) {
	h, err := s.Kv.Read("_" + hash)
	if err != nil {
		return
	}
	inth, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height := int32(inth)
	block, err = s.GetFromHeight(height)
	return
}

// GetTxBlockHeight returnes the height of the block based on its hash
func (s *service) GetTxBlockHeight(hash string) (height int32, err error) {
	if cached, ok := s.Cache.Get("h_" + hash); ok {
		height = cached.(int32)
		return
	}

	h, err := s.Kv.Read("_" + hash)
	if err != nil {
		return
	}
	inth, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height = int32(inth)

	if !s.Cache.Set("h_"+hash, height, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"height": height})
	}
	return
}
