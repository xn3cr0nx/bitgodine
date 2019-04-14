package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"google.golang.org/grpc"
)

// Config strcut containing initialization fields
type Config struct {
	Host string
	Port int
}

// Node represents the node structure in the dgraph
type Node struct {
	UID      string   `json:"uid,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Block    string   `json:"block,omitempty"`
	Height   int32    `json:"height,omitempty"`
	Locktime uint32   `json:"locktime,omitempty"`
	Inputs   []Input  `json:"inputs,omitempty"`
	Outputs  []Output `json:"outputs,omitempty"`
}

// Input represent input transaction, e.g. the link to a previous spent tx hash
type Input struct {
	UID  string `json:"uid,omitempty"`
	Hash string `json:"hash,omitempty"`
	Vout uint32 `json:"vout,omitempty"`
}

// Output represent output transaction, e.g. the value that can be spent as input
type Output struct {
	UID     string `json:"uid,omitempty"`
	Value   int64  `json:"value,omitempty"`
	Vout    uint32 `json:"vout"`
	Address string `json:"address,omitempty"`
}

// Resp represent the resp from a query function called q to dgraph
type Resp struct {
	Q []struct{ Node }
}

var instance *dgo.Dgraph

// Instance implements the singleton pattern to return the DGraph instance
func Instance(conf *Config) *dgo.Dgraph {
	if instance == nil {
		if conf == nil {
			logger.Panic("DGraph", errors.New("missing configuration"), logger.Params{})
		}
		// Dial a gRPC connection. The address to dial to can be configured when
		// setting up the dgraph cluster.
		d, err := grpc.Dial(fmt.Sprintf("%s:%d", conf.Host, conf.Port), grpc.WithInsecure())
		if err != nil {
			logger.Panic("DGraph", err, logger.Params{})
		}

		instance = dgo.NewDgraphClient(
			api.NewDgraphClient(d),
		)
		return instance
	}

	return instance
}

// Setup initializes the schema in dgraph
func Setup(c *dgo.Dgraph) error {
	err := c.Alter(context.Background(), &api.Operation{
		Schema: `
		hash: string @index(term) .
		block: string @index(term) .
		height: int @index(int) .
		vout: int @index(int) .
		value: int @index(int) .
		locktime: int @index(int) .
		address: string @index(term) .
		`,
	})
	return err
}

// Empty empties the dgraph instance removing all contained transactions
func Empty() error {
	resp, err := instance.NewTxn().Query(context.Background(), `{
		q(func: has(hash)) {
			uid
		}
	}`)
	if err != nil {
		return err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return err
	}
	if len(r.Q) > 0 {
		for _, e := range r.Q {
			e.Hash = ""
			e.Block = ""
			e.Locktime = 0
		}
		qx, err := json.Marshal(r.Q)
		if err != nil {
			return err
		}
		_, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{DeleteJson: qx, CommitNow: true})
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareInputs(inputs []*wire.TxIn, height int32) ([]Input, error) {
	var txIns []Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		txOutputs, err := GetTxOutputs(&h)
		if err != nil {
			return nil, err
		}
		var spentOutput Output
		for _, out := range txOutputs {
			if in.PreviousOutPoint.Index == uint32(4294967295) {
				spendingCoinbase := blocks.CoinbaseValue(height)
				if out.Value == spendingCoinbase {
					spentOutput = out
					break
				}
			}
			if out.Vout == in.PreviousOutPoint.Index {
				spentOutput = out
				break
			}
		}
		if spentOutput.UID == "" {
			return nil, errors.New("something not working")
		}
		txIns = append(txIns, Input{UID: spentOutput.UID, Hash: h, Vout: in.PreviousOutPoint.Index})
	}
	return txIns, nil
}

func prepareOutputs(outputs []*wire.TxOut) ([]Output, error) {
	var txOuts []Output
	for k, out := range outputs {
		if out.PkScript == nil {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value})
			txOuts = append(txOuts, Output{Value: out.Value})
		} else {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value, Vout: uint32(k)})
			_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
			if err != nil {
				return nil, err
			}
			txOuts = append(txOuts, Output{Value: out.Value, Vout: uint32(k), Address: addr[0].EncodeAddress()})
		}
	}

	return txOuts, nil
}

// StoreTx stores bitcoin transaction in the graph
func StoreTx(hash, block string, height int32, locktime uint32, inputs []*wire.TxIn, outputs []*wire.TxOut) error {
	// check if tx is already stored
	if _, err := GetTxUID(&hash); err == nil {
		logger.Debug("Dgraph", "already stored transaction", logger.Params{"hash": hash})
		return nil
	}

	txIns, err := prepareInputs(inputs, height)
	if err != nil {
		return err
	}
	txOuts, err := prepareOutputs(outputs)
	if err != nil {
		return err
	}

	node := Node{
		Hash:     hash,
		Block:    block,
		Height:   height,
		Locktime: locktime,
		Inputs:   txIns,
		Outputs:  txOuts,
	}
	out, err := json.Marshal(node)
	if err != nil {
		return err
	}
	resp, err := instance.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}
	logger.Debug("Dgraph", resp.String(), logger.Params{})
	return nil
}

// GetTx returnes the node from the query queried
func GetTx(field string, param *string) (Node, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: allofterms(%s, %s)) {
			block
			hash
			locktime
    	inputs {
				hash
				vout
			}
			outputs {
				value
				vout
				address
			}
		}
	}`, field, *param))
	if err != nil {
		return Node{}, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Node{}, err
	}
	if len(r.Q) == 0 {
		return Node{}, errors.New("transaction not found")
	}

	var node Node
	for _, e := range r.Q {
		if block := e.Node.Block; block != "" {
			node = e.Node
		}
	}
	if node.Block == "" {
		return Node{}, errors.New("Almost an object returned, but no one of them contains the block field")
	}
	return node, nil
}

// GetTxUID returnes the uid of the queried tx by hash
func GetTxUID(hash *string) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: allofterms(hash, %s)) {
			uid
		}
		}`, *hash))
	if err != nil {
		return "", err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.Q) == 0 {
		return "", errors.New("transaction not found")
	}
	uid := r.Q[0].UID
	return uid, nil
}

// GetTxOutputs returnes the outputs of the queried tx by hash
func GetTxOutputs(hash *string) ([]Output, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: allofterms(hash, %s)) {
			outputs {
				uid
				value
				vout
			}
		}
		}`, *hash))
	if err != nil {
		return []Output{}, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return []Output{}, err
	}
	if len(r.Q) == 0 {
		return []Output{}, errors.New("outputs not found")
	}
	outputs := r.Q[0].Outputs
	return outputs, nil
}

// GetFollowingTx returns the uid of the transaction spending the output (vout) of
// the transaction passed as input to the function
func GetFollowingTx(hash *string, vout *uint32) (Node, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
			q(func: has(inputs)) @cascade {
				uid
				block
				hash
				inputs @filter(eq(hash, %s) AND eq(vout, %d)){
					hash
					vout
				}
				outputs {
					value
					vout
				}
			}
			}`, *hash, *vout))
	if err != nil {
		return Node{}, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Node{}, err
	}
	if len(r.Q) == 0 {
		return Node{}, errors.New("transaction not found")
	}
	node := r.Q[0].Node
	return node, nil
}

// GetStoredTxs returnes all the stored transactions hashes
func GetStoredTxs() ([]string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), `{
			q(func: has(inputs)) {
				hash
			}
		}`)
	if err != nil {
		return nil, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Q) == 0 {
		return nil, errors.New("No transaction stored")
	}
	var transactions []string
	for _, tx := range r.Q {
		transactions = append(transactions, tx.Hash)
	}
	return transactions, nil
}

// GetAddressOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetAddressOccurences(address *btcutil.Address) ([]string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: has(outputs)) @cascade {
			uid
    	block
    	hash
			outputs @filter(allofterms(address, "%s")) {
				address
				value
			}
		}
	}`, (*address).String()))
	if err != nil {
		return nil, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Q) == 0 {
		return nil, errors.New("No address occurences")
	}
	var occurences []string
	for _, tx := range r.Q {
		occurences = append(occurences, tx.Hash)
	}
	return occurences, nil
}

// GetAddressBlocksOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetAddressBlocksOccurences(address *string) ([]string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: has(outputs)) @cascade {
			uid
    	block
    	hash
			outputs @filter(allofterms(address, "%s")) {
				address
				value
			}
		}
	}`, *address))
	if err != nil {
		return nil, err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Q) == 0 {
		return nil, errors.New("No address occurences")
	}
	var occurences []string
	for _, tx := range r.Q {
		occurences = append(occurences, tx.Block)
	}
	return occurences, nil
}

// GetBlockHashFromHeight returnes the hash of the block retrieving it based on its height
func GetBlockHashFromHeight(height int32) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: eq(height, %d), first: 1) {
			block
		}
	}`, height))
	if err != nil {
		return "", err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.Q) == 0 {
		return "", errors.New("No address occurences")
	}
	return r.Q[0].Block, nil
}
