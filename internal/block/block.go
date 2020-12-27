package block

import (
	"errors"
	"strconv"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

func fetchBlockTxs(db storage.DB, c *cache.Cache, txs []string) (transactions []tx.Tx, err error) {
	for _, hash := range txs {
		transaction, e := tx.GetFromHash(db, c, hash)
		if e != nil {
			return nil, e
		}
		transactions = append(transactions, transaction)
	}
	return
}

// StoreBlock inserts in the db the block as []byte passed
// for fast research purpose blocks have _ prefix, tx_ for txs prefix and h_ for height prefix
func StoreBlock(db storage.DB, b *Block, txs []tx.Tx) (err error) {
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
			batch[i.TxID+"_"+string(i.Vout)] = []byte(tx.TxID)
		}
	}

	err = db.StoreQueueBatch(batch)
	return
}

// read retrieves tx by hash
func read(db storage.DB, hash string) (block Block, err error) {
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
func ReadFromHeight(db storage.DB, c *cache.Cache, height int32) (block Block, err error) {
	hash, err := db.Read(strconv.Itoa(int(height)))
	if err != nil {
		return
	}
	block, err = GetFromHash(db, c, string(hash))
	if err != nil {
		return
	}
	return
}

// ReadHeight returnes the hash of the block retrieving it based on its height
func ReadHeight(db storage.DB) (height int32, err error) {
	h, err := db.Read("last")
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
func GetFromHash(db storage.DB, c *cache.Cache, hash string) (Block, error) {
	b, err := read(db, hash)
	if err != nil {
		return Block{}, err
	}

	return b, nil
}

// GetFromHeight return block structure based on block height
func GetFromHeight(db storage.DB, c *cache.Cache, height int32) (*BlockOut, error) {
	b, err := ReadFromHeight(db, c, height)
	if err != nil {
		return nil, err
	}

	txs, err := fetchBlockTxs(db, c, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetFromHashWithTxs return block structure based on block hash
func GetFromHashWithTxs(db storage.DB, c *cache.Cache, hash string) (*BlockOut, error) {
	b, err := GetFromHash(db, c, hash)
	if err != nil {
		return nil, err
	}

	txs, err := fetchBlockTxs(db, c, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetLast return last synced block
func GetLast(db storage.DB, c *cache.Cache) (*BlockOut, error) {
	h, err := ReadHeight(db)
	if err != nil {
		return nil, err
	}
	b, err := ReadFromHeight(db, c, h)
	if err != nil {
		return nil, err
	}

	txs, err := fetchBlockTxs(db, c, b.Transactions)
	if err != nil {
		return nil, err
	}

	return &BlockOut{b, txs}, nil
}

// GetFromHeightRange returnes the hash of the block retrieving it based on its height
func GetFromHeightRange(db storage.DB, c *cache.Cache, height int32, first int) (blocks []BlockOut, err error) {
	for i := height; i < (height + int32(first)); i++ {
		b, e := GetFromHeight(db, c, i)
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, *b)
	}
	return
}

// GetStored returnes list of all stored blocks
func GetStored(db storage.DB) (blocks []Block, err error) {
	h, err := ReadHeight(db)
	if err != nil {
		return
	}
	for i := int32(0); i <= h; i++ {
		hash, e := db.Read(strconv.Itoa(int(i)))
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, Block{ID: string(hash), Height: h})
	}
	return
}

// GetStoredList returnes all the stored block ids
func GetStoredList(db storage.DB, from int32) (blocks map[string]interface{}, err error) {
	// blocks, err = db.ReadKeysWithPrefix("_0000")
	h, err := ReadHeight(db)
	if err != nil {
		return
	}
	blocks = make(map[string]interface{}, h)
	for i := from; i <= h; i++ {
		hash, e := db.Read(strconv.Itoa(int(i)))
		if e != nil {
			err = e
			return
		}
		blocks[string(hash)] = nil
	}
	return
}

// Remove removes the block specified by its height
func Remove(db storage.DB, block *Block) error {
	return db.Delete(block.ID)
}

// RemoveLast removes the last block stored
func RemoveLast(db storage.DB, c *cache.Cache) error {
	h, err := ReadHeight(db)
	if err != nil {
		return err
	}
	block, err := GetFromHeight(db, c, h)
	if err != nil {
		return err
	}
	return db.Delete(block.ID)
}

// StoreFileParsed set file stored so far
func StoreFileParsed(db storage.DB, file int) (err error) {
	f := strconv.Itoa(file)
	err = db.Store("file", []byte(f))
	return
}

// GetFileParsed returnes the file parsed so far
func GetFileParsed(db storage.DB) (file int, err error) {
	f, err := db.Read("file")
	if err != nil {
		if errors.Is(err, errorx.ErrKeyNotFound) {
			return 0, nil
		}
		return
	}
	file, err = strconv.Atoi(string(f))
	return
}

// GetStoredTxs returnes all the stored transactions hashes
func GetStoredTxs(db storage.DB) (transactions []string, err error) {
	blocks, err := GetStored(db)
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			transactions = append(transactions, tx)
		}
	}
	return
}

// GetTxBlock returnes the block containing the transaction
func GetTxBlock(db storage.DB, c *cache.Cache, hash string) (block *BlockOut, err error) {
	h, err := db.Read("_" + hash)
	if err != nil {
		return
	}
	inth, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height := int32(inth)
	block, err = GetFromHeight(db, c, height)
	return
}

// GetTxBlockHeight returnes the height of the block based on its hash
func GetTxBlockHeight(db storage.DB, c *cache.Cache, hash string) (height int32, err error) {
	if cached, ok := c.Get("h_" + hash); ok {
		height = cached.(int32)
		return
	}

	h, err := db.Read("_" + hash)
	if err != nil {
		return
	}
	inth, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height = int32(inth)

	if !c.Set("h_"+hash, height, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"height": height})
	}
	return
}
