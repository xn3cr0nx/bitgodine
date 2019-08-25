package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil"
)

// GetAddressOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetAddressOccurences(address *btcutil.Address) ([]string, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: has(output)) @cascade {
			uid
			txid
			output @filter(allofterms(scriptpubkey_address, "%s")) {
				uid
				scriptpubkey
				scriptpubkey_asm
				scriptpubkey_type
				scriptpubkey_address
				value
				index
			}
		}
		}`, (*address).String()))
	if err != nil {
		return nil, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Txs) == 0 {
		return nil, errors.New("No address occurences")
	}
	var occurences []string
	for _, tx := range r.Txs {
		occurences = append(occurences, tx.Tx.TxID)
	}
	return occurences, nil
}

// GetAddressFirstOccurenceHeight returnes the height of the block in which the address appeared for the first time
func GetAddressFirstOccurenceHeight(address *btcutil.Address) (int32, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(fmt.Sprintf(`{    
		bl as var(func: has(prev_block)) @cascade {
			uid
			transactions {
				output @filter(allofterms(scriptpubkey_address, "%s")) 
				}		
		}   
		var(func: uid(bl)) {
			uid
			H as height
		}
		first() {
			min: min(val(H))
		}
	}`, (*address).EncodeAddress())))
	if err != nil {
		return 0, err
	}
	var r struct {
		First []struct{ Min int32 }
	}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return 0, err
	}
	if len(r.First) == 0 {
		return 0, errors.New("No address occurences")
	}
	if len(r.First) > 1 {
		return 0, errors.New("The min returnes more than an occurence. Strange behaviour")
	}
	return r.First[0].Min, nil
}
