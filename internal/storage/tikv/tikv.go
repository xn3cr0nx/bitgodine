package tikv

import (
	"github.com/imdario/mergo"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"

	ctx "golang.org/x/net/context"
)

// TiKV client wrapper
type TiKV struct {
	kv.Storage
}

// Config strcut containing initialization fields
type Config struct {
	URL string
}

// Conf returns default config struct
func Conf(URL string) *Config {
	// host := viper.GetString("tikv")
	host := "tikv:2379"
	if URL != "" {
		host = URL
	}

	return &Config{
		URL: host,
	}
}

// NewTiKV creates a new instance of the db
func NewTiKV(conf *Config) (*TiKV, error) {
	driver := tikv.Driver{}
	storage, err := driver.Open(conf.URL)
	return &TiKV{storage}, err
}

// Store insert new key-value in tikv
func (t *TiKV) Store(key string, value []byte) (err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	err = tx.Set([]byte(key), value)
	if err != nil {
		return
	}
	return tx.Commit(ctx.Background())
}

// StoreBatch insert new key-value in tikv
func (t *TiKV) StoreBatch(batch interface{}) (err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	series := batch.(map[string][]byte)
	for key, value := range series {
		if err = tx.Set([]byte(key), value); err != nil {
			return
		}
	}
	return tx.Commit(ctx.Background())
}

var queue map[string][]byte
var counter int

// StoreQueueBatch loads a queue until a threshold to perform a bulk insertion
func (t *TiKV) StoreQueueBatch(v interface{}) (err error) {
	series := v.(map[string][]byte)
	if queue == nil {
		queue = make(map[string][]byte, 0)
	}
	if err = mergo.Merge(&queue, series, mergo.WithOverride); err != nil {
		return
	}
	if counter >= 100 {
		if err = t.StoreBatch(queue); err != nil {
			return
		}
		queue = make(map[string][]byte, 0)
		counter = 0
	}
	counter++
	return
}

func (t *TiKV) Read(key string) (value []byte, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	value, err = tx.Get(ctx.Background(), []byte(key))
	return
}

// ReadKeys concurrent read values based on a prefix
func (t *TiKV) ReadKeys() (value []string, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(nil, nil)
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		value = append(value, string(it.Key()[:]))
		it.Next()
	}
	return
}

// ReadKeyValues concurrent read values based on a prefix
func (t *TiKV) ReadKeyValues() (value map[string][]byte, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(nil, nil)
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		value[string(it.Key()[:])] = it.Value()[:]
		it.Next()
	}
	return
}

// ReadKeysWithPrefix concurrent read values based on a prefix
func (t *TiKV) ReadKeysWithPrefix(prefix string) (keys []string, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(kv.Key(prefix), nil)
	if err != nil {
		return
	}
	defer it.Close()
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		keys = append(keys, string(it.Key()[:]))
		it.Next()
	}
	return
}

// ReadPrefix concurrent read values based on a prefix
func (t *TiKV) ReadPrefix(prefix string) (value [][]byte, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(kv.Key(prefix), nil)
	if err != nil {
		return
	}
	defer it.Close()
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		value = append(value, it.Value()[:])
		it.Next()
	}
	return
}

// ReadFirstValueByPrefix returns the first value matched by prefix
func (t *TiKV) ReadFirstValueByPrefix(prefix string) (value []byte, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(kv.Key(prefix), nil)
	if err != nil {
		return
	}
	defer it.Close()
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		value = it.Value()[:]
		return
	}
	return
}

// ReadPrefixWithKey concurrent read values based on a prefix
func (t *TiKV) ReadPrefixWithKey(prefix string) (value map[string][]byte, err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	it, err := tx.Iter(kv.Key(prefix), nil)
	if err != nil {
		return
	}
	defer it.Close()
	if err != nil {
		return
	}
	defer it.Close()
	for it.Valid() {
		value[string(it.Key()[:])] = it.Value()[:]
		it.Next()
	}
	return
}

// IsStored returns true if the block corresponding to passed hash is stored in db
func (t *TiKV) IsStored(key string) bool {
	_, err := t.Read(key)
	return err == nil
}

// Delete inserts in the db the block as []byte passed
func (t *TiKV) Delete(key string) (err error) {
	tx, err := t.Begin()
	if err != nil {
		return
	}
	if err = tx.Delete([]byte(key)); err != nil {
		return
	}
	return tx.Commit(ctx.Background())
}

// Empty empties the tikv store
func (t *TiKV) Empty() (err error) {
	keys, err := t.ReadKeys()
	if err != nil {
		return
	}

	for _, key := range keys {
		if err = t.Delete(key); err != nil {
			return
		}
	}

	return
}
