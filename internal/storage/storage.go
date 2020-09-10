package storage

// import (
// 	"github.com/xn3cr0nx/bitgodine/pkg/models"
// )

// // DB interface implements methods for a package to be used as data layer
// type DB interface {
// 	StoreBatch(interface{}) error
// 	Delete(string) error
// 	Empty() error

// 	// Address methods
// 	GetOccurences(string) ([]string, error)
// 	GetFirstOccurenceHeight(string) (int32, error)

// 	// Transaction methods
// 	GetTx(string) (models.Tx, error)
// 	GetTxOutputs(string) ([]models.Output, error)
// 	GetSpentTxOutput(string, uint32) (models.Output, error)
// 	GetFollowingTx(string, uint32) (models.Tx, error)
// 	GetStoredTxs() ([]string, error)
// 	GetTxBlockHeight(string) (int32, error)
// 	GetTxBlock(string) (models.Block, error)
// 	IsSpent(string, uint32) bool

// 	// Blocks methods
// 	IsStored(string) bool
// 	StoreBlock(interface{}, interface{}) error
// 	GetBlockFromHash(string) (models.Block, error)
// 	GetBlockFromHeight(int32) (models.Block, error)
// 	GetBlockFromHeightRange(int32, int) ([]models.Block, error)
// 	GetStoredBlocks() ([]models.Block, error)
// 	GetStoredBlocksList(int32) (map[string]interface{}, error)
// 	GetLastBlockHeight() (int32, error)
// 	LastBlock() (models.Block, error)
// 	RemoveBlock(*models.Block) error
// 	RemoveLastBlock() error

// 	StoreFileParsed(int) error
// 	GetFileParsed() (int, error)

// 	Close() error
// }

// DB interface implements methods for a generic key value storage db
type DB interface {
	Store(string, []byte) error
	StoreBatch(interface{}) error
	StoreQueueBatch(interface{}) error
	Read(string) ([]byte, error)
	ReadKeys() ([]string, error)
	ReadKeyValues() (map[string][]byte, error)
	ReadKeysWithPrefix(string) ([]string, error)
	ReadPrefix(string) ([][]byte, error)
	ReadFirstValueByPrefix(string) ([]byte, error)
	ReadPrefixWithKey(string) (map[string][]byte, error)
	IsStored(string) bool
	Delete(string) error
	Empty() error
	Close() error
}
