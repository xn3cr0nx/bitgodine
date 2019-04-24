package txs

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// Tx transaction type
type Tx struct {
	btcutil.Tx
}

// GenerateTransaction converts the Transaction node struct to a btcsuite Transaction struct
func GenerateTransaction(tx *dgraph.Transaction) (Tx, error) {
	msgTx := wire.NewMsgTx(tx.Version)
	for _, input := range tx.Inputs {
		hash, err := chainhash.NewHashFromStr(input.Hash)
		if err != nil {
			return Tx{}, err
		}
		prev := wire.NewOutPoint(hash, input.Vout)
		ti := wire.NewTxIn(prev, input.SignatureScript, wire.TxWitness(input.Witness))
		msgTx.AddTxIn(ti)
	}
	for _, output := range tx.Outputs {
		to := wire.NewTxOut(output.Value, output.PkScript)
		msgTx.AddTxOut(to)
	}
	transaction := btcutil.NewTx(msgTx)

	return Tx{Tx: *transaction}, nil
}

// Get retrieves and returnes the tx object
func Get(hash *chainhash.Hash) (Tx, error) {
	hashString := hash.String()
	tx, err := dgraph.GetTx(&hashString)
	if err != nil {
		return Tx{}, err
	}
	transaction, err := GenerateTransaction(&tx)
	if err != nil {
		return Tx{}, err
	}
	// blockHash, err := chainhash.NewHashFromStr(node.Block)
	// if err != nil {
	// 	return Tx{}, err
	// }
	// block, err := db.GetBlock(blockHash)
	// if err != nil {
	// 	return Tx{}, err
	// }
	// var transaction *btcutil.Tx
	// for _, t := range block.Transactions() {
	// 	if t.Hash().IsEqual(hash) {
	// 		transaction = t
	// 	}
	// }
	// return Tx{Tx: *transaction}, nil
	return transaction, nil
}

// IsCoinbase returnes true if the transaction is a coinbase transaction
func (tx *Tx) IsCoinbase() bool {
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	return tx.MsgTx().TxIn[0].PreviousOutPoint.Hash.IsEqual(zeroHash)
}

// BlockHeight returnes the height of the block that contains the transaction
func (tx *Tx) BlockHeight() (int32, error) {
	height, err := dgraph.GetTxBlockHeight(tx.Hash().String())
	if err != nil {
		return 0, err
	}
	// txHash := tx.Hash().String()
	// node, err := dgraph.GetTx(&txHash)
	// if err != nil {
	// 	return 0, nil
	// }
	// blockHash, err := chainhash.NewHashFromStr(node.Block)
	// if err != nil {
	// 	return 0, err
	// }
	// block, err := db.GetBlock(blockHash)
	// if err != nil {
	// 	return 0, err
	// }
	// return block.Height(), nil
	return height, nil
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
	// hash := tx.Hash()
	// hashString := hash.String()
	// transaction, err := dgraph.GetFollowingTx(&hashString, &index)
	hash := tx.Hash().String()
	transaction, err := dgraph.GetFollowingTx(&hash, &index)
	if err != nil {
		return Tx{}, err
	}
	// blockHash, err := chainhash.NewHashFromStr(node.Block)
	// if err != nil {
	// 	return Tx{}, err
	// }
	// block, err := db.GetBlock(blockHash)
	// if err != nil {
	// 	return Tx{}, err
	// }
	// var transaction *btcutil.Tx
	// for _, t := range block.Transactions() {
	// 	for _, i := range t.MsgTx().TxIn {
	// 		if i.PreviousOutPoint.Hash.IsEqual(hash) {
	// 			transaction = t
	// 		}
	// 	}
	// }
	// if transaction == nil {
	// 	return Tx{}, errors.New("something went wrong extracting the transaction")
	// }
	genTx, err := GenerateTransaction(&transaction)
	if err != nil {
		return Tx{}, err
	}
	return genTx, nil
}
