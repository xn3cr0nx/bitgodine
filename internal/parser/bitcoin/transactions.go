package bitcoin

import (
	"fmt"
	"runtime"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/task"
)

// Tx transaction type
type Tx struct {
	btcutil.Tx
}

// PrepareTransactions parses the btcutil.TX array of structs and convert them in Transaction object compatible with dgraph schema
// TODO: here I have to provide a solution in case the parsed block contains transactions which spend each other, e.g a transaction
// has inputs spending output from a tx in the same block. In this case utxo are not found and txin is not prepared. To fix this
// I have to define the id of the transaction with interested output and link the culprit inputs through that it. The approach are two:
// 1) my current solution starts from the assumption that this situation is uncommon, so is better to handle it just in those uncommon cases
// 2) if this situation is more common than I though, well is better to check this condition before to start parsing the tx, so I'll refactor
func PrepareTransactions(db storage.DB, txs []*btcutil.Tx) (transactions []tx.Tx, err error) {
	transactions = make([]tx.Tx, len(txs))

	pool := task.New(runtime.NumCPU() * 2)
	// block 170 seems to be the first with more than one tx
	for t := range txs {
		pool.Do(&TransactionsParser{t, txs[t], transactions})
	}
	err = pool.Shutdown()

	return
}

// TransactionsParser worker wrapper for parsing transactions in sync pool
type TransactionsParser struct {
	Index        int
	Tx           *btcutil.Tx
	Transactions []tx.Tx
}

// InputParser worker wrapper for parsing inputs in sync pool
type InputParser struct {
	Index  int
	Input  *wire.TxIn
	Inputs []tx.Input
}

// OutputParser worker wrapper for parsing inputs in sync pool
type OutputParser struct {
	Index   int
	Output  *wire.TxOut
	TxHash  string
	Outputs []tx.Output
}

// Work interface to execute transactions parser worker operations
func (w *TransactionsParser) Work() (err error) {
	pool := task.New(runtime.NumCPU() * 2)
	inputs := make([]tx.Input, len(w.Tx.MsgTx().TxIn))
	outputs := make([]tx.Output, len(w.Tx.MsgTx().TxOut))

	for i, in := range w.Tx.MsgTx().TxIn {
		pool.Do(&InputParser{
			i,
			in,
			inputs,
		})
	}
	for o, out := range w.Tx.MsgTx().TxOut {
		pool.Do(&OutputParser{
			o,
			out,
			w.Tx.Hash().String(),
			outputs,
		})
	}
	if err = pool.Shutdown(); err != nil {
		logger.Debug("Transactions", err.Error(), logger.Params{})
		return
	}

	w.Transactions[w.Index] = tx.Tx{
		TxID:     w.Tx.Hash().String(),
		Version:  w.Tx.MsgTx().Version,
		Locktime: w.Tx.MsgTx().LockTime,
		// Size
		// Weight
		// Fee
		Vin:  inputs,
		Vout: outputs,
	}
	return
}

// Work interface to execute input parser worker operations
func (w *InputParser) Work() (err error) {
	h := w.Input.PreviousOutPoint.Hash.String()
	var wtn []string
	for _, w := range [][]byte(w.Input.Witness) {
		wtn = append(wtn, string(w))
	}
	asm, e := txscript.DisasmString(w.Input.SignatureScript)
	if e != nil {
		logger.Debug("Transactions", e.Error(), logger.Params{})
		asm = e.Error()
	}
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	input := tx.Input{
		TxID:         h,
		Vout:         w.Input.PreviousOutPoint.Index,
		IsCoinbase:   w.Input.PreviousOutPoint.Hash.IsEqual(zeroHash),
		Scriptsig:    fmt.Sprintf("%X", w.Input.SignatureScript),
		ScriptsigAsm: asm,
		// InnerRedeemscriptAsm:  fmt.Sprintf("%X", w.Input.SignatureScript),
		// InnerWitnessscriptAsm:  fmt.Sprintf("%X", w.Input.SignatureScript),
		Sequence: w.Input.Sequence,
		Witness:  wtn,
		// Prevout
	}
	w.Inputs[w.Index] = input
	return
}

// Work interface to execute output parser worker operations
func (w *OutputParser) Work() (err error) {
	if w.Output.PkScript == nil { // there are invalid scripts
		w.Outputs[w.Index] = tx.Output{Value: w.Output.Value, Index: uint32(w.Index)}
		return
	}
	class, addr, _, e := txscript.ExtractPkScriptAddrs(w.Output.PkScript, &chaincfg.MainNetParams)
	if e != nil {
		logger.Debug("Transactions", e.Error(), logger.Params{"class": class, "addr": addr})
		// return
	}
	asm, e := txscript.DisasmString(w.Output.PkScript)
	if e != nil {
		logger.Debug("Transactions", e.Error(), logger.Params{})
		asm = e.Error()
	}
	// TODO: here should be managemed the multisig (just take all the addr, not just the first)
	output := tx.Output{
		Scriptpubkey:     fmt.Sprintf("%X", w.Output.PkScript),
		ScriptpubkeyAsm:  asm,
		ScriptpubkeyType: class.String(),
		Value:            w.Output.Value,
		Index:            uint32(w.Index),
	}
	if len(addr) > 0 {
		output.ScriptpubkeyAddress = addr[0].EncodeAddress()
	}
	w.Outputs[w.Index] = output
	return
}

// IsCoinbase returns true if the transaction is a coinbase transaction
func (tx *Tx) IsCoinbase() bool {
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	return tx.MsgTx().TxIn[0].PreviousOutPoint.Hash.IsEqual(zeroHash)
}
