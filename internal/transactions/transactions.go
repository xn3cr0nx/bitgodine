package txs

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

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

		sigScript, _ := hex.DecodeString(input.SignatureScript)
		ti := wire.NewTxIn(prev, sigScript, wire.TxWitness([][]byte{}))
		msgTx.AddTxIn(ti)
	}
	for _, output := range tx.Outputs {
		pkScript, _ := hex.DecodeString(output.PkScript)
		to := wire.NewTxOut(output.Value, pkScript)
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
	return transaction, nil
}

// // Store prepares the dgraph transaction struct and and call StoreTx to store it in dgraph
// func (tx *Tx) Store() error {
// 	// check if tx is already stored
// 	hash := tx.Hash().String()
// 	if _, err := dgraph.GetTxUID(&hash); err == nil {
// 		logger.Debug("Dgraph", "already stored transaction", logger.Params{"hash": hash})
// 		return nil
// 	}

// 	txIns, err := prepareInputs(tx.MsgTx().TxIn, nil)
// 	if err != nil {
// 		return err
// 	}
// 	txOuts, err := prepareOutputs(tx.MsgTx().TxOut)
// 	if err != nil {
// 		return err
// 	}

// 	transaction := dgraph.Transaction{
// 		Hash:     hash,
// 		Locktime: tx.MsgTx().LockTime,
// 		Version:  tx.MsgTx().Version,
// 		Inputs:   txIns,
// 		Outputs:  txOuts,
// 	}
// 	if err := dgraph.Store(&transaction); err != nil {
// 		return err
// 	}
// 	return nil
// }

// PrepareTransactions parses the btcutil.TX array of structs and convert them in Transaction object compatible with dgraph schema
// TODO: here I have to provide a solution in case the parsed block contains transactions which spend each other, e.g a transaction
// has inputs spending output from a tx in the same block. In this case utxo are not found and txin is not prepared. To fix this
// I have to define the id of the transaction with interested output and link the culprit inputs through that it. The approach are two:
// 1) my current solution starts from the assumption that this situation is uncommon, so is better to handle it just in those uncommon cases
// 2) if this situation is more common than I though, well is better to check this condition before to start parsing the tx, so I'll refactor
func PrepareTransactions(txs []*btcutil.Tx) ([]dgraph.Transaction, error) {
	var transactions []dgraph.Transaction
	for _, tx := range txs {
		var wg sync.WaitGroup
		wg.Add(2)
		alarm := make(chan error, 1)
		defer close(alarm)
		inputs := make(chan dgraph.Input, len(tx.MsgTx().TxIn))
		defer close(inputs)
		outputs := make(chan dgraph.Output, len(tx.MsgTx().TxOut))
		defer close(outputs)

		go prepareInputs(tx.MsgTx().TxIn, &transactions, &wg, inputs, alarm)
		go prepareOutputs(tx.MsgTx().TxOut, &wg, outputs, alarm)

		wg.Wait()
		select {
		case err := <-alarm:
			{
				logger.Error("Transactions", err, logger.Params{"tx": tx.Hash().String()})
				return nil, err
			}
		default:
		}

		var inputsResult []dgraph.Input
		var outputsResult []dgraph.Output
		var wg2 sync.WaitGroup
		wg2.Add(2)
		go func() {
			defer wg2.Done()
			for i := range inputs {
				inputsResult = append(inputsResult, i)
				if len(inputs) == 0 {
					break
				}
			}
		}()
		go func() {
			defer wg2.Done()
			for o := range outputs {
				outputsResult = append(outputsResult, o)
				if len(outputs) == 0 {
					break
				}
			}
		}()
		wg2.Wait()

		transactions = append(transactions, dgraph.Transaction{
			Hash:     tx.Hash().String(),
			Locktime: tx.MsgTx().LockTime,
			Version:  tx.MsgTx().Version,
			Inputs:   inputsResult,
			Outputs:  outputsResult,
		})
	}

	return transactions, nil
}

func prepareInputs(inputs []*wire.TxIn, transactions *[]dgraph.Transaction, s *sync.WaitGroup, inputsChannel chan<- dgraph.Input, inputsAlarm chan<- error) {
	defer s.Done()
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	var lock = sync.RWMutex{}
	for k := range inputs {
		j := k
		go func(index int, in *wire.TxIn) {
			defer wg.Done()
			h := in.PreviousOutPoint.Hash.String()
			lock.Lock()
			stxo, err := dgraph.GetSpentTxOutput(&h, &in.PreviousOutPoint.Index)
			lock.Unlock()
			if err != nil {
				if err.Error() != "output not found" {
					if len(inputsAlarm) == 1 {
						return
					}
					inputsAlarm <- err
					return
				}
			}
			var wtn []dgraph.TxWitness
			for _, w := range [][]byte(in.Witness) {
				wtn = append(wtn, dgraph.TxWitness(w))
			}
			input := dgraph.Input{UID: stxo.UID, Hash: h, Vout: in.PreviousOutPoint.Index, SignatureScript: fmt.Sprintf("%X", in.SignatureScript), Witness: wtn}
			if input.UID == "" && in.PreviousOutPoint.Index != uint32(4294967295) {
				for i, tx := range *transactions {
					if tx.Hash == in.PreviousOutPoint.Hash.String() {
						UID := fmt.Sprintf("_:utxo%dvout%d", i, in.PreviousOutPoint.Index)
						input.UID = UID
						tx.Outputs[in.PreviousOutPoint.Index].UID = UID
					}
				}
			}
			inputsChannel <- input
		}(j, inputs[j])
	}
	wg.Wait()
}

func prepareOutputs(outputs []*wire.TxOut, s *sync.WaitGroup, outputsChannel chan<- dgraph.Output, outputsAlarm chan<- error) {
	defer s.Done()
	var wg sync.WaitGroup
	wg.Add(len(outputs))
	for k := range outputs {
		i := k
		go func(index int, out *wire.TxOut) {
			defer wg.Done()
			if out.PkScript == nil {
				// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value})
				outputsChannel <- dgraph.Output{Value: out.Value}
			} else {
				_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
				if err != nil {
					if len(outputsAlarm) == 1 {
						return
					}
					outputsAlarm <- err
					return
				}
				// TODO: here should be managemed the multisig (just take all the addr, not just the first)
				if len(addr) > 0 {
					outputsChannel <- dgraph.Output{Value: out.Value, Vout: uint32(i), Address: addr[0].EncodeAddress(), PkScript: fmt.Sprintf("%X", out.PkScript)}
				} else {
					outputsChannel <- dgraph.Output{Value: out.Value, Vout: uint32(i), PkScript: fmt.Sprintf("%X", out.PkScript)}
				}
			}
		}(i, outputs[k])
	}
	wg.Wait()
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
	hash := tx.Hash().String()
	transaction, err := dgraph.GetFollowingTx(&hash, &index)
	if err != nil {
		return Tx{}, err
	}
	genTx, err := GenerateTransaction(&transaction)
	if err != nil {
		return Tx{}, err
	}
	return genTx, nil
}

// GetHeightRange returnes an array of pointer to transactions in the height boundaries range passed as argument
func GetHeightRange(from, to *int32) ([]Tx, error) {
	txs, err := dgraph.GetTransactionsHeightRange(from, to)
	if err != nil {
		return nil, err
	}
	var transactions []Tx
	for _, tx := range txs {
		transaction, err := GenerateTransaction(&tx)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
