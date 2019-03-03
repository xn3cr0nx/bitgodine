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

type Input struct {
	UID  string `json:"uid,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type Resp struct {
	Q []struct {
		Node
	}
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

func Setup(c *dgo.Dgraph) error {
	err := c.Alter(context.Background(), &api.Operation{
		Schema: `
		hash: string @index(term) .
		block: string @index(term) .
		locktime: int .
		`,
	})
	// input: uid .
	return err
}

// StoreTx stores bitcoin transaction in the graph
func StoreTx(dgraph *dgo.Dgraph, hash, block string, locktime uint32, inputs []*wire.TxIn) error {
	// check if tx is already stored
	if _, err := GetTxUID(dgraph, &hash); err == nil {
		logger.Info("Dgraph", "already stored transaction", logger.Params{"hash": hash})
		return nil
	}

	var txIns []Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		uid, err := GetTxUID(dgraph, &h)
		if err != nil {
			if err.Error() == "transaction not found" {
				StoreTx(dgraph, h, "", 0, nil)
				uid, err = GetTxUID(dgraph, &h)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		txIns = append(txIns, Input{UID: uid, Hash: h})
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
	resp, err := dgraph.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}
	logger.Debug("Dgraph", resp.String(), logger.Params{})
	return nil
}

// GetTx returnes the node from the query queried
func GetTx(dgraph *dgo.Dgraph, field string, param *string) (Resp, error) {
	resp, err := dgraph.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: allofterms(%s, %s)) {
			block
			hash
			locktime
    	inputs {
      	hash
    	}
		}
	}`, field, *param))
	if err != nil {
		return Resp{}, err
	}
	var q Resp
	if err := json.Unmarshal(resp.GetJson(), &q); err != nil {
		return Resp{}, err
	}
	return q, nil
}

// GetTxUID returnes the uid of the queried tx by hash
func GetTxUID(dgraph *dgo.Dgraph, hash *string) (string, error) {
	resp, err := dgraph.NewTxn().Query(context.Background(), fmt.Sprintf(`{
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
