package badger

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/dgraph-io/badger/v2"
	"github.com/imdario/mergo"
	"github.com/spf13/viper"
)

// Badger client wrapper
type Badger struct {
	*badger.DB
}

// Config strcut containing initialization fields
type Config struct {
	Dir string
}

// Conf returnes default config struct
func Conf(path string) *Config {
	dir := viper.GetString("badger")
	if path != "" {
		dir = path
	}

	return &Config{
		Dir: dir,
	}
}

// NewBadger creates a new instance of the db
func NewBadger(conf *Config, readonly bool) (*Badger, error) {
	opts := badger.DefaultOptions(conf.Dir)
	opts.Logger = nil
	// opts.NumVersionsToKeep = 1
	opts.Truncate = true
	opts.ReadOnly = readonly
	opts.ValueLogFileSize = 2000000000
	// opts.EventLogging = false
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Badger{db}, nil
}

// Store insert new key-value in badger
func (b *Badger) Store(key string, value []byte) error {
	return b.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// StoreBatch insert new key-value in badger
func (b *Badger) StoreBatch(batch interface{}) (err error) {
	series := batch.(map[string][]byte)
	wb := b.NewWriteBatch()
	defer wb.Cancel()

	for k, v := range series {
		if err = wb.Set([]byte(k), v); err != nil {
			return
		}
	}
	if err = wb.Flush(); err != nil {
		return
	}
	return
}

var queue map[string][]byte
var counter int

// StoreQueueBatch loads a queue until a threshold to perform a bulk insertion
func (b *Badger) StoreQueueBatch(v interface{}) (err error) {
	series := v.(map[string][]byte)
	if queue == nil {
		queue = make(map[string][]byte, 0)
	}
	if err = mergo.Merge(&queue, series, mergo.WithOverride); err != nil {
		return
	}
	if counter >= 100 {
		if err = b.StoreBatch(queue); err != nil {
			return
		}
		queue = make(map[string][]byte, 0)
		counter = 0
	}
	counter++
	return
}

// Read extract required value by key
func (b *Badger) Read(key string) (value []byte, err error) {
	err = b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		value = make([]byte, len(val))
		copy(value, val)
		return nil
	})
	if err != nil {
		return
	}
	return
}

// ReadKeys concurrent read values based on a prefix
func (b *Badger) ReadKeys() (value []string, err error) {
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			value = append(value, string(k))
		}
		return nil
	})
	return
}

// ReadKeyValues concurrent read values based on a prefix
func (b *Badger) ReadKeyValues() (value map[string][]byte, err error) {
	value = make(map[string][]byte)
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				value[string(item.Key())] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

// ReadKeysWithPrefix concurrent read values based on a prefix
func (b *Badger) ReadKeysWithPrefix(prefix string) (keys []string, err error) {
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))
		}
		return nil
	})
	return
}

// ReadPrefix concurrent read values based on a prefix
func (b *Badger) ReadPrefix(prefix string) (value [][]byte, err error) {
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			// k := item.Key()
			err := item.Value(func(v []byte) error {
				// fmt.Printf("key=%s, value=%s, %v\n", k, v, v)
				value = append(value, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

// ReadFirstValueByPrefix returns the first value matched by prefix
func (b *Badger) ReadFirstValueByPrefix(prefix string) (value []byte, err error) {
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			// k := item.Key()
			err := item.Value(func(v []byte) error {
				// fmt.Printf("key=%s, value=%s, %v\n", k, v, v)
				value = v
				return nil
			})
			if err != nil {
				return err
			}
			break
		}
		return nil
	})
	return
}

// ReadPrefixWithKey concurrent read values based on a prefix
func (b *Badger) ReadPrefixWithKey(prefix string) (value map[string][]byte, err error) {
	value = make(map[string][]byte)
	err = b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				// fmt.Printf("key=%s, value=%s, %v\n", k, v, v)
				value[string(k)] = v
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

// IsStored returns true if the block corresponding to passed hash is stored in db
func (b *Badger) IsStored(key string) bool {
	_, err := b.Read(key)
	return err == nil
}

// Delete inserts in the db the block as []byte passed
func (b *Badger) Delete(key string) (err error) {
	err = b.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return
}

// Empty empties the badger store
func (b *Badger) Empty() (err error) {
	dir, err := ioutil.ReadDir(viper.GetString("dbDir"))
	if err != nil {
		return
	}
	for _, d := range dir {
		if err = os.RemoveAll(path.Join([]string{viper.GetString("dbDir"), d.Name()}...)); err != nil {
			return
		}
	}
	return
}
