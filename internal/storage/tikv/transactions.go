package tikv

import (
	"errors"
	"strconv"

	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// GetTx retrieves tx by hash
func (db *KV) GetTx(hash string) (tx models.Tx, err error) {
	if cached, ok := db.cache.Get(hash); ok {
		tx = cached.(models.Tx)
		return
	}
	r, err := db.Read(hash)
	if err != nil {
		return
	}
	if err = encoding.Unmarshal(r, &tx); err != nil {
		return
	}
	if !db.cache.Set(tx.TxID, tx, 1) {
		logger.Error("Cache", errors.New("error caching"), logger.Params{"hash": tx.TxID})
	}
	return
}

// GetTxOutputs retrieves tx's outputs by hash
func (db *KV) GetTxOutputs(hash string) (outputs []models.Output, err error) {
	tx, err := db.GetTx(hash)
	if err != nil {
		return
	}
	outputs = tx.Vout
	return
}

// GetSpentTxOutput retrieves spent tx output based on hash and index
func (db *KV) GetSpentTxOutput(hash string, vout uint32) (output models.Output, err error) {
	tx, err := db.GetTx(hash)
	if err != nil {
		return
	}
	output = tx.Vout[vout]
	return
}

// GetFollowingTx retrieves spending tx of the output based on hash and index
func (db *KV) GetFollowingTx(hash string, vout uint32) (tx models.Tx, err error) {
	txHash, err := db.Read(hash + "_" + string(vout))
	if err != nil {
		return
	}
	tx, err = db.GetTx(string(txHash))
	return
}

// GetStoredTxs returns all the stored transactions hashes
func (db *KV) GetStoredTxs() (transactions []string, err error) {
	blocks, err := db.GetStoredBlocks()
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			transactions = append(transactions, tx)
		}
	}
	return
}

// GetTxBlock returns the block containing the transaction
func (db *KV) GetTxBlock(hash string) (block models.Block, err error) {
	h, err := db.Read("_" + hash)
	if err != nil {
		return
	}
	inth, err := strconv.Atoi(string(h))
	if err != nil {
		return
	}
	height := int32(inth)
	block, err = db.GetBlockFromHeight(height)
	return
}

// GetTxBlockHeight returns the height of the block based on its hash
func (db *KV) GetTxBlockHeight(hash string) (height int32, err error) {
	if cached, ok := db.cache.Get("h_" + hash); ok {
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

	if !db.cache.Set("h_"+hash, height, 1) {
		logger.Error("Cache", errors.New("error caching"), logger.Params{"height": height})
	}
	return
}

// IsSpent returns true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func (db *KV) IsSpent(tx string, index uint32) bool {
	_, err := db.GetFollowingTx(tx, index)
	return err == nil
}
