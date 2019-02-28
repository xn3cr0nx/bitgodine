package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	assert "gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestDGraphSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
}

func (suite *TestDGraphSuite) SetupSuite() {
	conf := &Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = Instance(conf)
}

// func (suite *TestDGraphSuite) SetupTest() {
// 	suite.clean.Acquire("compliances", "users", "files", "compliance_documents", "compliance_documents_files")
// }

func (suite *TestDGraphSuite) TearDownSuite() {
	resp, err := suite.dgraph.NewTxn().Query(context.Background(), `{
		q(func: allofterms(block, "0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449")) {
			uid
		}
	}`)
	assert.Equal(suite.T(), err, nil)
	var body struct {
		Q []struct {
			UID string `json:"uid,omitempty"`
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &body); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Deleting field with uid", body.Q[0].UID)
	out, err := json.Marshal(map[string]string{"uid": body.Q[0].UID})
	assert.Equal(suite.T(), err, nil)
	_, err2 := suite.dgraph.NewTxn().Mutate(context.Background(), &api.Mutation{DeleteJson: out, CommitNow: true})
	assert.Equal(suite.T(), err2, nil)
}

func (suite *TestDGraphSuite) TestSetup() {
	err := Setup(suite.dgraph)
	assert.Equal(suite.T(), err, nil)
}

func (suite *TestDGraphSuite) TestQuery() {
	resp, err := suite.dgraph.NewTxn().Query(context.Background(), `{
		q(func: has(input)) {
			block
			hash
			locktime
			input
			timestamp
		}
	}`)
	assert.Equal(suite.T(), err, nil)
	var body struct {
		Q []struct {
			Block    string `json:"block,omitempty"`
			Hash     string `json:"hash,omitempty"`
			Locktime int    `json:"locktime,omitempty"`
			Input    string `json:"input,omitempty"`
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &body); err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(suite.T(), len(body.Q), 3)
	assert.Equal(suite.T(), body.Q[0].Block, chaincfg.MainNetParams.GenesisHash.String())
}

func (suite *TestDGraphSuite) TestMutation() {
	type Node struct {
		Block    string `json:"block,omitempty"`
		Hash     string `json:"hash,omitempty"`
		Locktime int    `json:"locktime,omitempty"`
		Input    string `json:"input,omitempty"`
	}
	body := Node{
		Block:    "0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449",
		Hash:     "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		Locktime: 1234,
	}
	out, err := json.Marshal(body)
	assert.Equal(suite.T(), err, nil)
	_, err2 := suite.dgraph.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	assert.Equal(suite.T(), err2, nil)
}

func TestDgraph(t *testing.T) {
	suite.Run(t, new(TestDGraphSuite))
}
