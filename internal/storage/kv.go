package storage

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
