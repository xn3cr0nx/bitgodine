package models

import "time"

// Block model defined by standard
type Block struct {
	ID                string    `json:"id,omitempty"`
	Height            uint32    `json:"height,omitempty"`
	Version           uint8     `json:"version,omitempty"`
	Timestamp         time.Time `json:"timestamp,omitempty"`
	Bits              uint32    `json:"bits,omitempty"`
	Nonce             uint32    `json:"nonce,omitempty"`
	MerkleRoot        string    `json:"merkle_root,omitempty"`
	TxCount           int       `json:"tx_count,omitempty"`
	Size              uint32    `json:"size,omitempty"`
	Weight            uint32    `json:"weight,omitempty"`
	Previousblockhash string    `json:"previousblockhash,omitempty"`
	Proof             Proof     `json:"proof,omitempty"`
}

// Proof model part of Block
type Proof struct {
	Challenge    string `json:"challenge,omitempty"`
	ChallengeAsm string `json:"challenge_asm,omitempty"`
	Solution     string `json:"solution,omitempty"`
	SolutionAsm  string `json:"solution_asm,omitempty"`
}
