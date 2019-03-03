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
	err := Setup(suite.dgraph)
	assert.Equal(suite.T(), err, nil)
}

// func (suite *TestDGraphSuite) SetupTest() {
// 	suite.clean.Acquire("compliances", "users", "files", "compliance_documents", "compliance_documents_files")
// }

func (suite *TestDGraphSuite) TearDownSuite() {
	resp, err := suite.dgraph.NewTxn().Query(context.Background(), `{
		q(func: allofterms(block, "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")) {
			uid
			block
			hash
			locktime
			inputs {
				hash
				vout
			}
		}
	}`)
	assert.Equal(suite.T(), err, nil)
	var body Resp
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

func (suite *TestDGraphSuite) TestQuery() {
	resp, err := suite.dgraph.NewTxn().Query(context.Background(), `{
		q(func: allofterms(block, "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")) {
			block
			hash
			locktime
    	inputs {
				hash
				vout
    	}
		}
	}`)
	assert.Equal(suite.T(), err, nil)
	var body Resp
	if err := json.Unmarshal(resp.GetJson(), &body); err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(suite.T(), len(body.Q), 1)
	assert.Equal(suite.T(), body.Q[0].Block, chaincfg.MainNetParams.GenesisHash.String())
}

func (suite *TestDGraphSuite) TestStoreTx() {
	body := Node{
		Hash:     "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		Block:    "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		Locktime: uint32(1234),
		Inputs: []Input{
			Input{
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
				Vout: 4294967295,
			},
		},
	}
	out, err := json.Marshal(body)
	assert.Equal(suite.T(), err, nil)
	_, err2 := suite.dgraph.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	assert.Equal(suite.T(), err2, nil)
}

func TestDgraph(t *testing.T) {
	suite.Run(t, new(TestDGraphSuite))
}
