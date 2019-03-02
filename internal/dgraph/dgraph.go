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
	var txIns []Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		txIns = append(txIns, Input{UID: fmt.Sprintf("_:%s", h), Hash: h})
	}
	node := Node{
		UID:      fmt.Sprintf("_:%s", hash),
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
	logger.Info("Dgraph", resp.String(), logger.Params{})
	return nil
}

// GetTx returnes the node from the query queried
func GetTx(dgraph *dgo.Dgraph, param string) (Resp, error) {
	resp, err := dgraph.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: has(%s)) {
			block
			hash
			locktime
    	inputs {
      	hash
    	}
		}
	}`, param))
	if err != nil {
		return Resp{}, err
	}
	var q Resp
	if err := json.Unmarshal(resp.GetJson(), &q); err != nil {
		return Resp{}, err
	}
	return q, nil
}
