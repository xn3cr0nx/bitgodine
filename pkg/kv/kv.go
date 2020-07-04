package kv

// KV interface implements methods for a generic key value storage db
type KV interface {
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

	Delete(string) error
	Empty() error
	Close() error
}
