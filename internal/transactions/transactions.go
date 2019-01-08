package tx

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

type Tx struct {
	btcutil.Tx
	Version    uint32
	TxID       chainhash.Hash
	TxInCount  uint64
	TxOutCount uint64
	Locktime   uint32
	Slice      *[]uint8
}

type TxInput struct {
	PrevHash   *chainhash.Hash
	PrevIndex  uint32
	Script     []byte
	SequenceNo uint32
	Slice      *[]uint8
}

type TxOutput struct {
	Value  uint64
	Script []byte
	Slice  *[]uint8
}
