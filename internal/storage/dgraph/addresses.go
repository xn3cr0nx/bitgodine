package dgraph

import (
	"context"
	"errors"
)

// GetOccurences returns an array containing the transactions where the address appears in the blockchain
func (d *Dgraph) GetOccurences(address string) (occurences []string, err error) {
	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			txs(func: has(output)) @cascade {
				uid
				txid
				output @filter(allofterms(scriptpubkey_address, "$s")) {
					uid
					scriptpubkey
					scriptpubkey_asm
					scriptpubkey_type
					scriptpubkey_address
					value
					index
				}
			}
		}`, map[string]string{"$s": address})
	if err != nil {
		return
	}
	var r TxResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Txs) == 0 {
		err = errors.New("no address occurences")
		return
	}
	for _, tx := range r.Txs {
		occurences = append(occurences, tx.Tx.TxID)
	}
	return
}

// GetFirstOccurenceHeight returns the height of the block in which the address appeared for the first time
func (d *Dgraph) GetFirstOccurenceHeight(address string) (height int32, err error) {
	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			bl as var(func: has(prev_block)) @cascade {
				uid
				transactions {
					output @filter(allofterms(scriptpubkey_address, "$s")) 
					}		
			}   
			var(func: uid(bl)) {
				uid
				H as height
			}
			first() {
				min: min(val(H))
			}
		}`, map[string]string{"$s": address})
	if err != nil {
		return
	}
	var r struct {
		First []struct{ Min int32 }
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.First) == 0 {
		err = errors.New("No address occurences")
		return
	}
	if len(r.First) > 1 {
		err = errors.New("the min returns more than an occurence. Strange behaviour")
		return
	}
	height = r.First[0].Min
	return
}
