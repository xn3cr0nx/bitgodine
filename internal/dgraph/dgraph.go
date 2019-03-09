package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
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
	UID      string  `json:"uid,omitempty"`
	Hash     string  `json:"hash,omitempty"`
	Block    string  `json:"block,omitempty"`
	Locktime uint32  `json:"locktime,omitempty"`
	Inputs   []Input `json:"inputs,omitempty"`
}

// Input represent input transaction, e.g. the link to a previous spent tx hash
type Input struct {
	UID  string `json:"uid,omitempty"`
	Hash string `json:"hash,omitempty"`
	Vout uint32 `json:"vout,omitempty"`
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
		vout: int @index(int) .
		locktime: datetime @index(int) .
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
	if len(r.Q) == 0 {
		return errors.New("Dgraph is empty")
	}
	qx, err := json.Marshal(r.Q)
	if err != nil {
		return err
	}
	_, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{DeleteJson: qx, CommitNow: true})
	if err != nil {
		return err
	}
	return nil
}

// StoreTx stores bitcoin transaction in the graph
func StoreTx(hash, block string, locktime uint32, inputs []*wire.TxIn) error {
	// check if tx is already stored
	if _, err := GetTxUID(&hash); err == nil {
		logger.Debug("Dgraph", "already stored transaction", logger.Params{"hash": hash})
		return nil
	}

	var txIns []Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		uid, err := GetTxUID(&h)
		if err != nil {
			if err.Error() == "transaction not found" {
				StoreTx(h, "", 0, nil)
				uid, err = GetTxUID(&h)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		txIns = append(txIns, Input{UID: uid, Hash: h, Vout: in.PreviousOutPoint.Index})
	}
	node := Node{
		Hash:     hash,
		Block:    block,
		Locktime: locktime,
		Inputs:   txIns,
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
	node := r.Q[0].Node
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
