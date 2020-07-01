package dgraph

import (
	"context"
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// UtxoSet tracks utxos
type UtxoSet struct {
	UID     string `json:"uid,omitempty"`
	UtxoSet []Utxo `json:"utxoset,omitempty"`
}

// Utxo tracks utxos
type Utxo struct {
	UID  string          `json:"uid,omitempty"`
	TxID string          `json:"txid,omitempty"`
	Utxo []models.Output `json:"utxo"` // TODO: this could just be a refernce, e.g. the UID of the parsed output node
}

// UtxoResp basic structure to unmarshall utxo query
type UtxoResp struct {
	U []struct{ UtxoSet }
}

// NewUtxoSet stores the basic struct to manage the utxo sets
func (d *Dgraph) NewUtxoSet() error {
	c := UtxoSet{
		UtxoSet: []Utxo{
			{
				UID: "_:init",
			},
		},
	}
	if err := d.Store(c); err != nil {
		return err
	}
	return nil
}

// GetUtxoSet returnes the set of all utxos stored in dgraph
func (d *Dgraph) GetUtxoSet() (UtxoSet, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		u(func: has(utxoset)) {
			uid
			utxoset {
				uid
				txid
				utxo (orderasc: index) (first: 1000000000) {
					uid
					scriptpubkey_type
					scriptpubkey_address
					index
				}
			}
		}
	}`)
	if err != nil {
		return UtxoSet{}, err
	}
	var r UtxoResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return UtxoSet{}, err
	}
	if len(r.U) == 0 {
		return UtxoSet{}, errors.New("UtxoSet not found")
	}

	logger.Debug("Dgraph Cluster", "Retrieving Utxo", logger.Params{})
	return r.U[0].UtxoSet, nil
}

// GetUtxoSetUID returns the UID of the cluster
func (d *Dgraph) GetUtxoSetUID() (string, error) {
	if cached, ok := d.cache.Get("utxoUID"); ok {
		return cached.(string), nil
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		u(func: has(utxoset)) {
			uid
		}
	}`)
	if err != nil {
		return "", err
	}
	var r UtxoResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.U) == 0 {
		return "", errors.New("UtxoSet not found")
	}

	if err == nil {
		if !d.cache.Set("clusterUID", r.U[0].UtxoSet.UID, 1) {
			logger.Error("Cache", errors.New("error caching"), logger.Params{})
		}
	}
	return r.U[0].UtxoSet.UID, nil
}

// GetUtxoSetByHash returnes the UID of the specified set of addresses
func (d *Dgraph) GetUtxoSetByHash(hash string) (Utxo, error) {
	if cached, ok := d.cache.Get(fmt.Sprintf("utxo_%s", hash)); ok {
		return cached.(Utxo), nil
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		u(func: has(utxoset)) {
			uid
			utxoset @filter(eq(txid, %s)) {
				uid
				txid
				utxo (orderasc: index) (first: 1000000000) {
					uid
					scriptpubkey_type
					scriptpubkey_address
					index
				}
			}
		}
	}`, hash))
	if err != nil {
		return Utxo{}, err
	}
	var r UtxoResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Utxo{}, err
	}
	if len(r.U) == 0 {
		return Utxo{}, errors.New("Cluster not found")
	}
	if len(r.U[0].UtxoSet.UtxoSet[0].Utxo) == 0 {
		return Utxo{}, errors.New("Set not found")
	}

	if err == nil {
		if !d.cache.Set(fmt.Sprintf("utxo_%s", hash), r.U[0].UtxoSet.UtxoSet[0], 1) {
			logger.Error("Cache", errors.New("error caching"), logger.Params{})
		}
	}

	return r.U[0].UtxoSet.UtxoSet[0], nil
}

// NewUtxo add a new group of unspent output associated with the tx hash
func (d *Dgraph) NewUtxo(txid string, outputs []models.Output) error {
	uid, err := d.GetUtxoSetUID()
	if err != nil {
		return err
	}
	u := []Utxo{
		{
			TxID: txid,
			Utxo: outputs,
		},
	}
	set := UtxoSet{
		UID:     uid,
		UtxoSet: u,
	}
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// RemoveStxo remove a spent transaction output from the utxos set
func (d *Dgraph) RemoveStxo(txid string, index uint32) error {
	// uid, err := d.GetUtxoSetUID()
	// if err != nil {
	// 	return err
	// }
	// TODO: this could be implemented as upsert. query utxosetbyhash and then delete mutaiton
	// set, err := d.GetUtxoSetByHash(txid)
	// if err != nil {
	// 	return err
	// }
	// var newSet []models.Output
	// for _, u := range set.Utxo {
	// 	if u.Index != index {
	// 		newSet = append(newSet, u)
	// 	}
	// }
	// var uid string
	// for _, u := range set.Utxo {
	// 	if u.Index == index {
	// 		uid = u.UID
	// 	}
	// }
	// utxo := []Utxo{
	// 	{
	// 		UID:  set.UID,
	// 		Utxo: newSet,
	// 	},
	// }
	// u := UtxoSet{
	// 	UID:     uid,
	// 	UtxoSet: utxo,
	// }

	// if _, err := d.Store(u); err != nil {
	// 	return err
	// }
	// if uid == "" {
	// 	return errors.New("output not found")
	// }
	// if err := d.Delete(uid); err != nil {
	// 	return err
	// }
	return nil
}
