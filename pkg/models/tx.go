package models

import "time"

// Tx model defined by standard
type Tx struct {
	TxID     string   `json:"txid,omitempty"`
	Version  int32    `json:"version"`
	Locktime uint32   `json:"locktime"`
	Size     float32  `json:"size"`
	Weight   float32  `json:"weight"`
	Fee      float32  `json:"fee"`
	Vin      []Input  `json:"input,omitempty"`
	Vout     []Output `json:"output,omitempty"`
	Status   []Status `json:"status,omitempty"` // I don't get why this should be an array, dgraph set it to array by default
}

// Input model part of Tx
type Input struct {
	TxID                  string   `json:"txid,omitempty"`
	Vout                  uint32   `json:"vout"`
	IsCoinbase            bool     `json:"is_coinbase"`
	Scriptsig             string   `json:"scriptsig"`
	ScriptsigAsm          string   `json:"scriptsig_asm"`
	InnerRedeemscriptAsm  string   `json:"inner_redeemscript_asm"`
	InnerWitnessscriptAsm string   `json:"inner_witnessscript_asm"`
	Sequence              uint32   `json:"sequence"`
	Witness               []string `json:"witness"`
	Prevout               uint32   `json:"prevout"`
	// IsPegin               bool
	// Issuance              Issuance
}

// Output model part of Tx
type Output struct {
	Scriptpubkey        string `json:"scriptpubkey"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm"`
	ScriptpubkeyType    string `json:"scriptpubkey_type"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address"`
	Value               int64  `json:"value"`
	Index               uint32 `json:"index"` // this shoudln't be here, useful for dgraph
	// Valuecommitment     uint64 `json:"valuecommitment,omitempty"`
	// Asset               string `json:"asset,omitempty"`
	// Pegout              Pegout `json:"pegout,omitempty"`
}

// // (Elements only) Issuance model part of Input
// type Issuance struct {
// 	AssetID            string  `json:"asset_id,omitempty"`
// 	IsReissuance       bool    `json:"is_reissuance,omitempty"`
// 	AssetBlindingNonce string  `json:"asset_blinding_nonce,omitempty"`
// 	AssetEntropy       string  `json:"asset_entropy,omitempty"`
// 	ContractHash       string  `json:"contract_hash,omitempty"`
// 	Assetamount        float32 `json:"assetamount,omitempty"`
// 	Tokenamount        float32 `json:"tokenamount,omitempty"`
// }

// // (Elements only) Pegout model part of Output
// type Pegout struct {
// 	GenesisHash         string `json:"genesis_hash,omitempty"`
// 	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
// 	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
// 	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
// }

// Status model part of Tx
type Status struct {
	Confirmed   bool      `json:"confirmed"`
	BlockHeight int32     `json:"block_height"`
	BlockHash   string    `json:"block_hash"`
	BlockTime   time.Time `json:"block_time"`
}
