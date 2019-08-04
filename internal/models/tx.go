package models

import "time"

// Tx model defined by standard
type Tx struct {
	TxID     string   `json:"txid,omitempty"`
	Version  uint8    `json:"version,omitempty"`
	Locktime uint32   `json:"locktime,omitempty"`
	Size     float32  `json:"size,omitempty"`
	Weight   float32  `json:"weight,omitempty"`
	Fee      float32  `json:"fee,omitempty"`
	Vin      []Input  `json:"vin,omitempty"`
	Vout     []Output `json:"vout,omitempty"`
	Status   Status   `json:"status,omitempty"`
}

// Input model part of Tx
type Input struct {
	TxID                  string
	Vout                  uint32
	IsCoinbase            bool
	Scriptsig             string
	ScriptsigAsm          string
	InnerRedeemscriptAsm  string
	InnerWitnessscriptAsm string
	Sequence              uint32
	Witness               []string
	Prevout               uint32
	IsPegin               bool
	Issuance              Issuance
}

// Issuance model part of Input
type Issuance struct {
	AssetID            string  `json:"asset_id,omitempty"`
	IsReissuance       bool    `json:"is_reissuance,omitempty"`
	AssetBlindingNonce string  `json:"asset_blinding_nonce,omitempty"`
	AssetEntropy       string  `json:"asset_entropy,omitempty"`
	ContractHash       string  `json:"contract_hash,omitempty"`
	Assetamount        float32 `json:"assetamount,omitempty"`
	Tokenamount        float32 `json:"tokenamount,omitempty"`
}

// Output model part of Tx
type Output struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               uint64 `json:"value,omitempty"`
	Valuecommitment     uint64 `json:"valuecommitment,omitempty"`
	Asset               string `json:"asset,omitempty"`
	Pegout              Pegout `json:"pegout,omitempty"`
}

// Pegout model part of Output
type Pegout struct {
	GenesisHash         string `json:"genesis_hash,omitempty"`
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
}

// Status model part of Tx
type Status struct {
	Confirmed   bool      `json:"confirmed,omitempty"`
	BlockHeight uint32    `json:"block_height,omitempty"`
	BlockHash   string    `json:"block_hash,omitempty"`
	BlockTime   time.Time `json:"block_time,omitempty"`
}
