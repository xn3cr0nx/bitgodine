package dgraph

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

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

// Empty removes all data from dgraph with a drop all command
func Empty() error {
	var cmd = []byte(`{ "drop_all": true }`)
	req, err := http.NewRequest("POST", "localhost:8080/alter", bytes.NewBuffer(cmd))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
