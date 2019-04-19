package dgraph

import (
	"errors"

	"github.com/btcsuite/btcutil"
)

//TODO: fix this methods with new logic
// GetAddressOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetAddressOccurences(address *btcutil.Address) ([]string, error) {
	// 	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
	// 		q(func: has(outputs)) @cascade {
	// 			uid
	//     	block
	//     	hash
	// 			outputs @filter(allofterms(address, "%s")) {
	// 				address
	// 				value
	// 			}
	// 		}
	// 	}`, (*address).String()))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	var r Resp
	// 	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
	// 		return nil, err
	// 	}
	// 	if len(r.Q) == 0 {
	return nil, errors.New("No address occurences")
	// 	}
	// 	var occurences []string
	// 	for _, tx := range r.Q {
	// 		occurences = append(occurences, tx.Hash)
	// 	}
	// 	return occurences, nil
}

// // GetAddressBlocksOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetAddressBlocksOccurences(address *string) ([]string, error) {
	// 	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
	// 		q(func: has(outputs)) @cascade {
	// 			uid
	//     	block
	//     	hash
	// 			outputs @filter(allofterms(address, "%s")) {
	// 				address
	// 				value
	// 			}
	// 		}
	// 	}`, *address))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	var r Resp
	// 	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
	// 		return nil, err
	// 	}
	// 	if len(r.Q) == 0 {
	return nil, errors.New("No address occurences")
	// 	}
	// 	var occurences []string
	// 	for _, tx := range r.Q {
	// 		occurences = append(occurences, tx.Block)
	// 	}
	// 	return occurences, nil
}
