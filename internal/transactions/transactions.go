package txs

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// Tx transaction type
type Tx struct {
	btcutil.Tx
}

// Get retrieves and returnes the tx object
func Get(hash *chainhash.Hash) (Tx, error) {
	hashString := hash.String()
	node, err := dgraph.GetTx("hash", &hashString)
	if err != nil {
		return Tx{}, err
	}
	blockHash, err := chainhash.NewHashFromStr(node.Block)
	if err != nil {
		return Tx{}, err
	}
	block, err := db.GetBlock(blockHash)
	if err != nil {
		return Tx{}, err
	}
	var transaction *btcutil.Tx
	for _, t := range block.Transactions() {
		if t.Hash().IsEqual(hash) {
			transaction = t
		}
	}
	return Tx{Tx: *transaction}, nil
}

// IsCoinbase returnes true if the transaction is a coinbase transaction
func (tx *Tx) IsCoinbase() bool {
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	return tx.MsgTx().TxIn[0].PreviousOutPoint.Hash.IsEqual(zeroHash)
}

// GetSpentTx returnes the spent transaction corresponding to the index
// passed between input transactions
func (tx *Tx) GetSpentTx(index uint32) (Tx, error) {
	if len(tx.MsgTx().TxIn)-1 < int(index) {
		return Tx{}, errors.New("Index out of range in transaction input")
	}
	hash := tx.MsgTx().TxIn[index].PreviousOutPoint.Hash
	coinbaseHash, err := chainhash.NewHash(make([]byte, 32))
	if err != nil {
		return Tx{}, err
	}
	if (&hash).IsEqual(coinbaseHash) {
		return Tx{}, errors.New("Coinbase transaction")
	}
	transaction, err := Get(&hash)
	if err != nil {
		return Tx{}, err
	}
	return transaction, nil
}

// IsSpent returnes true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func (tx *Tx) IsSpent(index uint32) bool {
	hashString := tx.Hash().String()
	_, err := dgraph.GetFollowingTx(&hashString, &index)
	if err != nil {
		fmt.Println("error", err)
		// just for sake of clarity, untill I'm going to refactor this piece to be more useful
		if err.Error() == "transaction not found" {
			return false
		}
		return false
	}
	return true
}

// GetSpendingTx returns the transaction spending the output tx of the transaction
// passed by its index as argument
func (tx *Tx) GetSpendingTx(index uint32) (Tx, error) {
	if len(tx.MsgTx().TxOut)-1 < int(index) {
		return Tx{}, errors.New("Index out of range in transaction input")
	}
	hash := tx.Hash()
	hashString := hash.String()
	node, err := dgraph.GetFollowingTx(&hashString, &index)
	if err != nil {
		return Tx{}, err
	}
	blockHash, err := chainhash.NewHashFromStr(node.Block)
	if err != nil {
		return Tx{}, err
	}
	block, err := db.GetBlock(blockHash)
	if err != nil {
		return Tx{}, err
	}
	var transaction *btcutil.Tx
	for _, t := range block.Transactions() {
		for _, i := range t.MsgTx().TxIn {
			if i.PreviousOutPoint.Hash.IsEqual(hash) {
				transaction = t
			}
		}
	}
	if transaction == nil {
		return Tx{}, errors.New("something went wrong extracting the transaction")
	}
	return Tx{Tx: *transaction}, nil
}
