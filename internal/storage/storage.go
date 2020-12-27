package storage

import (
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/storage/redis"
	"github.com/xn3cr0nx/bitgodine/internal/storage/tikv"
)

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

// NewStorage returns a new DB instance based on the environment
func NewStorage() (db DB, err error) {
	switch viper.GetString("db") {
	case "tikv":
		db, err = tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
		if err != nil {
			return
		}
	case "badger":
		db, err = badger.NewBadger(badger.Conf(viper.GetString("badger")), false)
		if err != nil {
			return
		}
	case "redis":
		db, err = redis.NewRedis(redis.Conf(viper.GetString("redis")))
		if err != nil {
			return
		}
	}
	return
}
