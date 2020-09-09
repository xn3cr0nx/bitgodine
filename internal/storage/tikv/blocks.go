package tikv

import (
	"strconv"

	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// IsStored returns true if the block corresponding to passed hash is stored in db
func (db *KV) IsStored(hash string) bool {
	_, err := db.Read(hash)
	return err == nil
}

// StoreBlock inserts in the db the block as []byte passed
// for fast research purpose blocks have _ prefix, tx_ for txs prefix and h_ for height prefix
func (db *KV) StoreBlock(v interface{}, t interface{}) (err error) {
	b := v.(*models.Block)
	txs := t.([]models.Tx)
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

	err = db.StoreBatch(batch)
	return
}

// GetBlockFromHash retrieves block by hash
func (db *KV) GetBlockFromHash(hash string) (block models.Block, err error) {
	r, err := db.Read(hash)
	if err != nil {
		return
	}
	if err = encoding.Unmarshal(r, &block); err != nil {
		return
	}
	return
}

// GetBlockFromHeight retrieves block by height
func (db *KV) GetBlockFromHeight(height int32) (block models.Block, err error) {
	hash, err := db.Read(strconv.Itoa(int(height)))
	if err != nil {
		return
	}
	block, err = db.GetBlockFromHash(string(hash))
	if err != nil {
		return
	}
	return
}

// GetBlockFromHeightRange returns the hash of the block retrieving it based on its height
func (db *KV) GetBlockFromHeightRange(height int32, first int) (blocks []models.Block, err error) {
	for i := height; i < (height + int32(first)); i++ {
		b, e := db.GetBlockFromHeight(i)
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, b)
	}
	return
}

// GetLastBlockHeight returns the hash of the block retrieving it based on its height
func (db *KV) GetLastBlockHeight() (height int32, err error) {
	h, err := db.Read("last")
	if err != nil {
		if err.Error() == "Key not found" {
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

// LastBlock returns the hash of the block retrieving it based on its height
func (db *KV) LastBlock() (block models.Block, err error) {
	h, err := db.GetLastBlockHeight()
	if err != nil {
		return
	}
	block, err = db.GetBlockFromHeight(h)
	return
}

// GetStoredBlocks returns list of all stored blocks
func (db *KV) GetStoredBlocks() (blocks []models.Block, err error) {
	h, err := db.GetLastBlockHeight()
	if err != nil {
		return
	}
	for i := int32(0); i <= h; i++ {
		hash, e := db.Read(strconv.Itoa(int(i)))
		if e != nil {
			err = e
			return
		}
		blocks = append(blocks, models.Block{ID: string(hash), Height: h})
	}
	return
}

// GetStoredBlocksList returns all the stored block ids
func (db *KV) GetStoredBlocksList(from int32) (blocks map[string]interface{}, err error) {
	// blocks, err = db.ReadKeysWithPrefix("_0000")
	h, err := db.GetLastBlockHeight()
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

// RemoveBlock removes the block specified by its height
func (db *KV) RemoveBlock(block *models.Block) error {
	return db.Delete(block.ID)
}

// RemoveLastBlock removes the last block stored
func (db *KV) RemoveLastBlock() error {
	h, err := db.GetLastBlockHeight()
	if err != nil {
		return err
	}
	block, err := db.GetBlockFromHeight(h)
	if err != nil {
		return err
	}
	return db.Delete(block.ID)
}

// StoreFileParsed set file stored so far
func (db *KV) StoreFileParsed(file int) (err error) {
	f := strconv.Itoa(file)
	err = db.Store("file", []byte(f))
	return
}

// GetFileParsed returns the file parsed so far
func (db *KV) GetFileParsed() (file int, err error) {
	f, err := db.Read("file")
	if err != nil {
		if err.Error() == "Key not found" {
			return 0, nil
		}
		return
	}
	file, err = strconv.Atoi(string(f))
	return
}
