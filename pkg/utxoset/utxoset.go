package utxoset

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// UtxoSet unspent transaction output in memory and persistant storage
type UtxoSet struct {
	set  map[string]map[uint32]string
	bolt *bolt.DB
	lock sync.RWMutex
	disk bool
}

// Config struct containing initialization fields
type Config struct {
	Path string
	Disk bool
}

// Conf exports the Config object to initialize indexing Utxoset
func Conf(file string, disk bool) *Config {
	if file == "" {
		file = viper.GetString("utxo")
	}

	return &Config{
		Path: file,
		Disk: disk,
	}
}

var instance *UtxoSet

// NewUtxoSet implements the singleton pattern to return the UtxoSet instance
func NewUtxoSet(conf *Config) *UtxoSet {
	if instance == nil {
		if conf == nil {
			logger.Panic("Utxoset", errors.New("missing configuration"), logger.Params{})
		}

		instance = &UtxoSet{
			set:  make(map[string]map[uint32]string),
			lock: sync.RWMutex{},
			disk: conf.Disk,
		}
		if conf.Disk {
			path := strings.Split(conf.Path, "/")
			if len(path) > 1 {
				dir := filepath.Join(path[:len(path)-1]...)
				err := os.Mkdir("/"+dir, os.ModeDir)
				if err != nil {
					logger.Error("Utxoset", err, logger.Params{})
				}
			}
			db, err := bolt.Open(conf.Path, 0600, nil)
			if err != nil {
				log.Fatal(err)
			}
			instance.bolt = db
		}
	}
	return instance
}

// StoreUtxoSet write a new tx outputs set to the struct
func (u *UtxoSet) StoreUtxoSet(hash string, outputs map[uint32]string) (err error) {
	u.lock.Lock()
	u.set[hash] = outputs
	u.lock.Unlock()

	if u.disk {
		err = u.bolt.Batch(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(hash))
			if err != nil {
				return err
			}

			for i := range outputs {
				if err := b.Put(Itob(int(i)), []byte(outputs[i])); err != nil {
					fmt.Println("Error writing bucket", err)
					return err
				}
			}
			return nil
		})
	}
	return
}

// GetUtxo returnes the required output
func (u *UtxoSet) GetUtxo(hash string, vout uint32) (set string, err error) {
	u.lock.RLock()
	set, ok := u.set[hash][vout]
	u.lock.RUnlock()
	if !ok {
		err = errors.New("index out of range")
		return
	}
	return
}

// GetStoredUtxo returnes the bolt stored output
func (u *UtxoSet) GetStoredUtxo(hash string, vout uint32) (set string, err error) {
	if !u.disk {
		return "", errors.New("cannot retrieve, disk not initialized")
	}
	err = u.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hash))
		if b == nil {
			return errors.New("index out of range")
		}
		v := b.Get(Itob(int(vout)))
		if v == nil {
			return errors.New("index out of range")
		}

		set = string(v)
		return nil
	})
	return
}

// GetUtxoSet returnes the required output set (outputs of a transaction)
func (u *UtxoSet) GetUtxoSet(hash string) (sets map[uint32]string, err error) {
	u.lock.RLock()
	sets, ok := u.set[hash]
	u.lock.RUnlock()
	if !ok {
		err = errors.New("index out of range")
		return
	}
	return
}

// GetStoredUtxoSet returnes the bolt stored output
func (u *UtxoSet) GetStoredUtxoSet(hash string) (sets map[uint32]string, err error) {
	if !u.disk {
		return nil, errors.New("cannot retrieve, disk not initialized")
	}
	err = u.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hash))
		if b == nil {
			return errors.New("index out of range")
		}
		sets = make(map[uint32]string)
		return b.ForEach(func(k, v []byte) error {
			sets[uint32(Btoi(k))] = string(v)
			return nil
		})
	})
	return
}

// GetFullUtxoSet returnes the full stored utxoset
func (u *UtxoSet) GetFullUtxoSet() (set map[string]map[uint32]string, err error) {
	if err = u.Restore(); err != nil {
		return
	}
	return u.set, nil
}

// CheckUtxoSet returnes true if the utxo set is stored
func (u *UtxoSet) CheckUtxoSet(hash string) bool {
	if !u.disk {
		return false
	}
	err := u.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hash))
		if b == nil {
			return errors.New("index out of range")
		}
		return nil
	})
	return err == nil
}

// DeleteUtxo deletes the set in the tx bucket
func (u *UtxoSet) DeleteUtxo(hash string, vout uint32) (err error) {
	u.lock.RLock()
	if _, ok := u.set[hash][vout]; !ok {
		u.lock.RUnlock()
		err = errors.New("utxo not found")
		return
	}
	u.lock.RUnlock()

	if u.disk {
		err = u.bolt.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(hash))
			if b == nil {
				return errors.New("bucket not found")
			}
			if err := b.Delete(Itob(int(vout))); err != nil {
				return err
			}

			if k, _ := b.Cursor().First(); len(k) == 0 {
				if err := tx.DeleteBucket([]byte(hash)); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return
}

// DeleteUtxoSet deletes the tx bucket
func (u *UtxoSet) DeleteUtxoSet(hash string) (err error) {
	if u.disk {
		err = u.bolt.Update(func(tx *bolt.Tx) error {
			return tx.DeleteBucket([]byte(hash))
		})
	}
	return
}

// Restore initialize memory set mapping the stored set
// func (u *UtxoSet) Restore(last string) (err error) {
func (u *UtxoSet) Restore() (err error) {
	if !u.disk {
		return
	}

	err = u.bolt.View(func(tx *bolt.Tx) error {
		if err := tx.ForEach(func(hash []byte, b *bolt.Bucket) error {
			u.set[string(hash)] = make(map[uint32]string)
			return b.ForEach(func(k, v []byte) error {
				u.lock.Lock()
				u.set[string(hash)][uint32(Btoi(k))] = string(v)
				u.lock.Unlock()
				return nil
			})
		}); err != nil {
			return err
		}

		// c := tx.Cursor()
		// if k, _ := c.Seek([]byte(last)); k == nil {
		// 	err = errors.New("utxo set incomplete")
		// }
		return err
	})
	logger.Debug("Blockchain", "restored utxo set", logger.Params{"txs": len(u.set)})
	return
}

// Itob returns an 8-byte big endian representation of v
func Itob(v int) (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return
}

// Btoi returns a decimal representation of a 8-byte big endian
func Btoi(b []byte) (v int) {
	v = int(binary.BigEndian.Uint64(b))
	return
}

// Close closed bolt connection
func (u *UtxoSet) Close() error {
	return u.bolt.Close()
}
