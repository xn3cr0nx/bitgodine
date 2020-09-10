package block

import (
	"time"

	"github.com/xn3cr0nx/bitgodine/internal/tx"
)

// Block model defined by standard
type Block struct {
	ID                string    `json:"id,omitempty"`
	Height            int32     `json:"height,omitempty"`
	Version           int32     `json:"version,omitempty"`
	Timestamp         time.Time `json:"timestamp,omitempty"`
	Bits              uint32    `json:"bits,omitempty"`
	Nonce             uint32    `json:"nonce,omitempty"`
	MerkleRoot        string    `json:"merkle_root,omitempty"`
	Transactions      []string  `json:"transactions,omitempty"`
	TxCount           int       `json:"tx_count,omitempty"`
	Size              int       `json:"size,omitempty"`
	Weight            int       `json:"weight,omitempty"`
	Previousblockhash string    `json:"previousblockhash,omitempty"`
	// Proof             Proof     `json:"proof,omitempty"`
}

// // (Elements models) Proof model part of Block
// type Proof struct {
// 	Challenge    string `json:"challenge,omitempty"`
// 	ChallengeAsm string `json:"challenge_asm,omitempty"`
// 	Solution     string `json:"solution,omitempty"`
// 	SolutionAsm  string `json:"solution_asm,omitempty"`
// }

// BlockOut enhanced model block with full transactions
type BlockOut struct {
	Block
	Transactions []tx.Tx `json:"transactions"`
}
