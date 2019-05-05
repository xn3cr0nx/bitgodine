package txs

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Tx transaction type
type Tx struct {
	btcutil.Tx
}

// GenerateTransaction converts the Transaction node struct to a btcsuite Transaction struct
func GenerateTransaction(tx *dgraph.Transaction) (Tx, error) {
	msgTx := wire.NewMsgTx(tx.Version)
	msgTx.LockTime = tx.Locktime
	for _, input := range tx.Inputs {
		hash, err := chainhash.NewHashFromStr(input.Hash)
		if err != nil {
			return Tx{}, err
		}
		prev := wire.NewOutPoint(hash, input.Vout)

		var witness [][]byte
		for _, w := range input.Witness {
			witness = append(witness, []byte(w))
		}

		ti := wire.NewTxIn(prev, []byte(input.SignatureScript), wire.TxWitness(witness))
		msgTx.AddTxIn(ti)
	}
	for _, output := range tx.Outputs {
		to := wire.NewTxOut(output.Value, []byte(output.PkScript))
		msgTx.AddTxOut(to)
	}
	transaction := btcutil.NewTx(msgTx)

	return Tx{Tx: *transaction}, nil
}

// Get retrieves and returnes the tx object
func Get(hash *chainhash.Hash) (Tx, error) {
	tx, err := dgraph.GetTx(hash.String())
	if err != nil {
		return Tx{}, err
	}
	transaction, err := GenerateTransaction(&tx)
	if err != nil {
		return Tx{}, err
	}
	fmt.Println("transaction", transaction.Hash().String(), len(transaction.MsgTx().TxOut), transaction.MsgTx().TxIn[0].PreviousOutPoint.Hash.String())
	return transaction, nil
}

// Store prepares the dgraph transaction struct and and call StoreTx to store it in dgraph
func (tx *Tx) Store() error {
	// check if tx is already stored
	hash := tx.Hash().String()
	if _, err := dgraph.GetTxUID(&hash); err == nil {
		logger.Debug("Dgraph", "already stored transaction", logger.Params{"hash": hash})
		return nil
	}

	txIns, err := prepareInputs(tx.MsgTx().TxIn, nil)
	if err != nil {
		return err
	}
	txOuts, err := prepareOutputs(tx.MsgTx().TxOut)
	if err != nil {
		return err
	}

	transaction := dgraph.Transaction{
		Hash:     hash,
		Locktime: tx.MsgTx().LockTime,
		Version:  tx.MsgTx().Version,
		Inputs:   txIns,
		Outputs:  txOuts,
	}
	// if err := dgraph.StoreTx(&transaction); err != nil {
	if err := dgraph.Store(&transaction); err != nil {
		return err
	}
	return nil
}

// PrepareTransactions parses the btcutil.TX array of structs and convert them in Transaction object compatible with dgraph schema
// TODO: here I have to provide a solution in case the parsed block contains transactions which spend each other, e.g a transaction
// has inputs spending output from a tx in the same block. In this case utxo are not found and txin is not prepared. To fix this
// I have to define the id of the transaction with interested output and link the culprit inputs through that it. The approach are two:
// 1) my current solution starts from the assumption that this situation is uncommon, so is better to handle it just in those uncommon cases
// 2) if this situation is more common than I though, well is better to check this condition before to start parsing the tx, so I'll refactor
func PrepareTransactions(txs []*btcutil.Tx) ([]dgraph.Transaction, error) {
	var transactions []dgraph.Transaction
	for _, tx := range txs {
		// fmt.Println("Parsing tx", tx.Hash().String())
		inputs, err := prepareInputs(tx.MsgTx().TxIn, &transactions)
		if err != nil {
			return nil, err
		}
		outputs, err := prepareOutputs(tx.MsgTx().TxOut)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, dgraph.Transaction{
			Hash:     tx.Hash().String(),
			Locktime: tx.MsgTx().LockTime,
			Version:  tx.MsgTx().Version,
			Inputs:   inputs,
			Outputs:  outputs,
		})
	}
	return transactions, nil
}

func prepareInputs(inputs []*wire.TxIn, transactions *[]dgraph.Transaction) ([]dgraph.Input, error) {
	var txIns []dgraph.Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		// fmt.Println("input", h)
		stxo, err := dgraph.GetSpentTxOutput(&h, &in.PreviousOutPoint.Index)
		if err != nil {
			if err.Error() != "output not found" {
				return nil, err
			}
		}
		var wtn []dgraph.TxWitness
		for _, w := range [][]byte(in.Witness) {
			wtn = append(wtn, dgraph.TxWitness(w))
		}
		input := dgraph.Input{UID: stxo.UID, Hash: h, Vout: in.PreviousOutPoint.Index, SignatureScript: fmt.Sprintf("%X", in.SignatureScript), Witness: wtn}
		// This is for managing the TODO specified above
		if input.UID == "" && in.PreviousOutPoint.Index != uint32(4294967295) {
			for i, tx := range *transactions {
				if tx.Hash == in.PreviousOutPoint.Hash.String() {
					UID := fmt.Sprintf("_:utxo%dvout%d", i, in.PreviousOutPoint.Index)
					input.UID = UID
					tx.Outputs[in.PreviousOutPoint.Index].UID = UID
				}
			}
		}
		txIns = append(txIns, input)
	}
	return txIns, nil
}

func prepareOutputs(outputs []*wire.TxOut) ([]dgraph.Output, error) {
	var txOuts []dgraph.Output
	for k, out := range outputs {
		if out.PkScript == nil {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value})
			txOuts = append(txOuts, dgraph.Output{Value: out.Value})
		} else {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value, Vout: uint32(k)})
			_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
			if err != nil {
				return nil, err
			}
			txOuts = append(txOuts, dgraph.Output{Value: out.Value, Vout: uint32(k), Address: addr[0].EncodeAddress(), PkScript: fmt.Sprintf("%X", out.PkScript)})
		}
	}

	return txOuts, nil
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
