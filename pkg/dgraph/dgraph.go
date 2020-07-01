package dgraph

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Dgraph wrapper of dgraph client
type Dgraph struct {
	*dgo.Dgraph
	cache *cache.Cache
}

// Config strcut containing initialization fields
type Config struct {
	Host string
	Port int
}

// Queue data structure used as broker to load insertion queue
type Queue struct {
	q    []interface{}
	size int
}

// Conf exports the Config object to initialize indexing dgraph
func Conf() *Config {
	return &Config{
		Host: viper.GetString("dgHost"),
		Port: viper.GetInt("dgPort"),
	}
}

var queue *Queue
var instance *Dgraph

// Instance implements the singleton pattern to return the DGraph instance
func Instance(conf *Config, c *cache.Cache) *Dgraph {
	if instance == nil {
		if conf == nil {
			logger.Panic("DGraph", errors.New("missing configuration"), logger.Params{})
		}
		// Dial a gRPC connection. The address to dial to can be configured when
		// setting up the dgraph cluster.
		d, err := grpc.Dial(fmt.Sprintf("%s:%d", conf.Host, conf.Port), grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*1024),
			grpc.MaxCallSendMsgSize(1024*1024*1024)),
			grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
		if err != nil {
			logger.Panic("DGraph", err, logger.Params{})
		}

		i := dgo.NewDgraphClient(
			api.NewDgraphClient(d),
		)
		instance = &Dgraph{i, c}
		return instance
	}

	return instance
}

// Setup initializes the schema in dgraph
func (d *Dgraph) Setup() error {
	err := d.Alter(context.Background(), &api.Operation{
		Schema: `
		txid: string @index(term) .
		id: string @index(term) .
		prev_block: string @index(term) .
		height: int @index(int) .
		vout: int @index(int) .
		index: int @index(int) .
		value: int @index(int) .
		locktime: int @index(int) .
		scriptpubkey_address: string @index(term) .
		timestamp: datetime .
		size: int .
		cluster: int @index(int) .
		parent: int @index(int) .
		rank: int @index(int) .
		pos: int @index(int) .
		`,
	})
	// prev_block: string @index(term) @reverse .
	return err
}

// Empty removes all data from dgraph with a drop all command
func (d *Dgraph) Empty() (err error) {
	var cmd = []byte(`{ "drop_all": true }`)
	req, err := http.NewRequest("POST", "http://localhost:8080/alter", bytes.NewBuffer(cmd))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

// Store encodes received object as json object and stores it in dgraph
func (d *Dgraph) Store(v interface{}) (err error) {
	// out, err := json.Marshal(v)
	// if err != nil {
	// 	return
	// }
	// resp, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	// // Basic fallback mechanism to retry on high load traffic
	// if err != nil && strings.Contains(err.Error(), "transport is closing") {
	// 	time.Sleep(2 * time.Second)
	// 	resp, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	// }

	return
}

// StoreBatch loads a queue untill a threshold to perform a bulk insertion on dgraph
func (d *Dgraph) StoreBatch(v interface{}) (err error) {
	if queue == nil {
		queue = &Queue{
			q:    make([]interface{}, 0),
			size: 0,
		}
	}
	queue.q = append(queue.q, v)
	queue.size += int(unsafe.Sizeof(v))
	if queue.size >= 20000 {
		txn := instance.NewTxn()
		defer txn.Discard(context.Background())
		for i, e := range queue.q {
			out, err := json.Marshal(e)
			if err != nil {
				return err
			}
			if _, err := txn.Mutate(context.Background(), &api.Mutation{SetJson: out}); err != nil {
				return err
			}
			fmt.Println("mutated block", i)
		}
		if err = txn.Commit(context.Background()); err != nil {
			return
		}
		queue.q = make([]interface{}, 0)
		queue.size = 0
	}
	return
}

// Delete removes the node corresponding to the passed uid
func (d *Dgraph) Delete(UID string) (err error) {
	var delete = struct {
		UID string `json:"uid"`
	}{
		UID: UID,
	}
	out, err := json.Marshal(delete)
	if err != nil {
		return
	}
	_, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{DeleteJson: out, CommitNow: true})
	return
}
