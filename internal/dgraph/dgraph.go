package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
		prev_block: string @index(term) .
		block: string @index(term) .
		height: int @index(int) .
		vout: int @index(int) .
		value: int @index(int) .
		locktime: int @index(int) .
		address: string @index(term) .
		time: datetime .
		`,
	})
	// prev_block: string @index(term) @reverse .
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
