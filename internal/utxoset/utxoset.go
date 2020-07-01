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

// Conf exports the Config object to initialize indexing dgraph
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

// Instance implements the singleton pattern to return the UtxoSet instance
func Instance(conf *Config) *UtxoSet {
	if instance == nil {
		if conf == nil {
			logger.Panic("DGraph", errors.New("missing configuration"), logger.Params{})
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
func (u *UtxoSet) GetUtxo(hash string, vout uint32) (uid string, err error) {
	u.lock.RLock()
	uid, ok := u.set[hash][vout]
	u.lock.RUnlock()
	if !ok {
		err = errors.New("index out of range")
		return
	}
	return
}

// GetStoredUtxo returnes the bolt stored output
func (u *UtxoSet) GetStoredUtxo(hash string, vout uint32) (uid string, err error) {
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

		uid = string(v)
		return nil
	})
	return
}

// GetUtxoSet returnes the required output set (outputs of a transaction)
func (u *UtxoSet) GetUtxoSet(hash string) (uids map[uint32]string, err error) {
	u.lock.RLock()
	uids, ok := u.set[hash]
	u.lock.RUnlock()
	if !ok {
		err = errors.New("index out of range")
		return
	}
	return
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

// DeleteUtxo deletes the uid in the tx bucket
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
func (u *UtxoSet) Restore(last string) (err error) {
	if !u.disk {
		return errors.New("cannot retrieve, disk not initialized")
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

		c := tx.Cursor()
		if k, _ := c.Seek([]byte(last)); k == nil {
			err = errors.New("utxo set incomplete")
		}
		return err
	})
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

// // UpdateUtxoSet update spent transaction outputs after block is correctly stored.
// // Consistency risk if software carshes between this operations. to avoid consistency errors using dgraph call
// // Operations are all concurrently executed since utxoset is lock safe
// func UpdateUtxoSet(utxoset *utxoset.UtxoSet, transactions []models.Tx, uids *map[string]map[uint32]string) (err error) {
// 	errAlarm := make(chan error, 1)
// 	var wg sync.WaitGroup
// 	wg.Add(len(transactions))
// 	for k := range transactions {
// 		go func(ind int, transaction models.Tx) {
// 			defer wg.Done()

// 			alarm := make(chan error, 1)
// 			var wgio sync.WaitGroup
// 			wgio.Add(2)

// 			go func(index int, transaction models.Tx) {
// 				defer wgio.Done()
// 				var wgi sync.WaitGroup
// 				wgi.Add(len(transaction.Vin))
// 				for k := range transaction.Vin {
// 					index := k
// 					go func(index int, i models.Input) {
// 						defer wgi.Done()
// 						if i.Vout != 4294967295 {
// 							utxoset.DeleteUtxo(i.TxID, i.Vout)
// 						}
// 					}(index, transaction.Vin[index])
// 				}
// 				wgi.Wait()
// 			}(ind, transactions[ind])

// 			go func(index int, transaction models.Tx) {
// 				defer wgio.Done()
// 				if err := utxoset.StoreUtxoSet(transaction.TxID, (*uids)[transaction.TxID]); err != nil {
// 					alarm <- err
// 					return
// 				}
// 			}(ind, transactions[ind])
// 			wgio.Wait()
// 			select {
// 			case e := <-alarm:
// 				errAlarm <- e
// 				return
// 			default:
// 			}
// 		}(k, transactions[k])
// 	}
// 	wg.Wait()
// 	select {
// 	case e := <-errAlarm:
// 		err = e
// 		return
// 	default:
// 	}
// 	return
// }

// // RestoreUtxoSet is called when the running session should recover data previously stored
// func (p *Parser) RestoreUtxoSet(rawChain [][]uint8, last *models.Block) (err error) {
// 	logger.Debug("Blockchain", "restoring utxo set", logger.Params{"last": last.ID})
// 	if err = p.utxoset.Restore(last.ID); err != nil {
// 		if err.Error() == "utxo set incomplete" {
// 			newChain := make([][]uint8, len(rawChain))
// 			copy(newChain, rawChain)
// 			if err = p.UtxoSetRecovery(last, newChain, true); err != nil {
// 				return
// 			}
// 		} else {
// 			return
// 		}
// 	}
// 	return
// }

// // UtxoSetRecovery parses data file and restore utxoset
// func (p *Parser) UtxoSetRecovery(last *models.Block, chain [][]uint8, progress bool) (err error) {
// 	logger.Debug("Blockchain", "recovering utxo set", logger.Params{"last": last.ID})

// 	var pb *mpb.Progress
// 	var bar *mpb.Bar
// 	if progress {
// 		pb = mpb.New(mpb.WithWidth(64))
// 		name := "Retrieving blocks:"
// 		bar = pb.AddBar(int64(last.Height),
// 			// set custom bar style, default one is "[=>-]"
// 			mpb.BarStyle("╢=>-╟"),
// 			mpb.PrependDecorators(
// 				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
// 				decor.OnComplete(
// 					decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 4}), "done",
// 				),
// 			),
// 			mpb.AppendDecorators(decor.Percentage()),
// 		)
// 	}

// 	logger.Debug("Blockchain", "retrieving list of uids to populate utxo set, this will be painful...", logger.Params{})
// 	step := 100000
// 	uids := make(map[string]map[string][]string)
// 	for h := 0; h < int(last.Height); h += step {
// 		start := time.Now()
// 		slice, e := p.db.GetBlockTxOutputsFromRange(int32(h), step)
// 		if err != nil {
// 			if progress {
// 				bar.Abort(true)
// 			}
// 			return e
// 		}
// 		for k, v := range slice {
// 			uids[k] = v
// 			if progress {
// 				bar.IncrBy(1, time.Since(start))
// 			}
// 		}
// 		if h >= step && step >= 10000 {
// 			step = step / 2
// 		}
// 	}
// 	if progress {
// 		if !bar.Completed() {
// 			bar.SetTotal(100, true)
// 		}
// 		pb.Wait()
// 	}
// 	logger.Info("Blockchain", "Blocks fetched", logger.Params{"size": len(uids)})

// 	var bar2 *mpb.Bar
// 	if progress {
// 		pb = mpb.New(mpb.WithWidth(64))
// 		name := "Creating Utxo Set:"
// 		bar2 = pb.AddBar(int64(len(uids)),
// 			// set custom bar style, default one is "[=>-]"
// 			mpb.BarStyle("╢=>-╟"),
// 			mpb.PrependDecorators(
// 				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
// 				decor.OnComplete(
// 					decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 4}), "done",
// 				),
// 			),
// 			mpb.AppendDecorators(decor.Percentage()),
// 		)
// 	}

// 	for _, slice := range chain {
// 		start := time.Now()

// 		for len(slice) > 0 {
// 			block, e := blocks.Parse(&slice)
// 			if e != nil {
// 				return e
// 			}

// 			// if empty the block is skipped, hence is not still stored
// 			if len(uids[block.Hash().String()]) == 0 {
// 				continue
// 			}
// 			if stored := p.utxoset.CheckUtxoSet(block.Hash().String()); !stored {
// 				errAlarm := make(chan error, 1)
// 				var wg sync.WaitGroup
// 				wg.Add(len(block.Transactions()))
// 				for k := range block.Transactions() {
// 					go func(ind int, transaction *btcutil.Tx) {
// 						defer wg.Done()

// 						alarm := make(chan error, 1)
// 						var wgio sync.WaitGroup
// 						wgio.Add(2)

// 						go func(index int, transaction *btcutil.Tx) {
// 							defer wgio.Done()
// 							var wgi sync.WaitGroup
// 							wgi.Add(len(transaction.MsgTx().TxIn))
// 							for j := range transaction.MsgTx().TxIn {
// 								go func(index int, i *wire.TxIn) {
// 									defer wgi.Done()
// 									if i.PreviousOutPoint.Index != 4294967295 {
// 										p.utxoset.DeleteUtxo(i.PreviousOutPoint.Hash.String(), i.PreviousOutPoint.Index)
// 									}
// 								}(j, transaction.MsgTx().TxIn[j])
// 							}
// 							wgi.Wait()
// 						}(ind, block.Transactions()[ind])

// 						go func(index int, transaction *btcutil.Tx) {
// 							defer wgio.Done()
// 							uidset := make(map[uint32]string)
// 							for o := range transaction.MsgTx().TxOut {
// 								if len(uids[block.Hash().String()][transaction.Hash().String()]) != len(transaction.MsgTx().TxOut) {
// 									logger.Error("Blockchain", errors.New("inconsistency in block"), logger.Params{"block": block.Hash().String()})
// 									if progress {
// 										bar2.Abort(true)
// 									}
// 									alarm <- errors.New("inconsistent length of transaction outputs")
// 									return
// 								}
// 								uidset[uint32(o)] = uids[block.Hash().String()][transaction.Hash().String()][uint32(o)]
// 							}
// 							if err := p.utxoset.StoreUtxoSet(transaction.Hash().String(), uidset); err != nil {
// 								alarm <- err
// 								return
// 							}
// 						}(ind, block.Transactions()[ind])
// 						wgio.Wait()
// 						select {
// 						case e := <-alarm:
// 							errAlarm <- e
// 							return
// 						default:
// 						}
// 					}(k, block.Transactions()[k])
// 				}
// 				wg.Wait()
// 				select {
// 				case e := <-errAlarm:
// 					return e
// 				default:
// 				}

// 				// delete block to reduce memory consumption
// 				delete(uids, block.Hash().String())

// 				if progress {
// 					bar2.IncrBy(1, time.Since(start))
// 				}
// 			}

// 			if block.Hash().String() == last.ID {
// 				if progress {
// 					if !bar2.Completed() {
// 						bar2.SetTotal(100, true)
// 					}
// 					pb.Wait()
// 				}
// 				return nil
// 			}
// 		}
// 	}
// 	if progress {
// 		bar2.Abort(true)
// 	}
// 	err = errors.New("Parsed the entire chain, head not found")
// 	return
// }
