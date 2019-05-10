package dgraph

import (
	"testing"

	"github.com/dgraph-io/dgo"
	assert "gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestTransactionsSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
}

func (suite *TestTransactionsSuite) SetupSuite() {
	conf := &Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = Instance(conf)
	err := Setup(suite.dgraph)
	assert.Equal(suite.T(), err, nil)
}

// // func (suite *TestTransactionsSuite) SetupTest() {
// // 	suite.clean.Acquire("compliances", "users", "files", "compliance_documents", "compliance_documents_files")
// // }

// func (suite *TestTransactionsSuite) TearDownSuite() {
// 	resp, err := suite.dgraph.NewTxn().Query(context.Background(), `{
// 		q(func: allofterms(block, "00000000000000009c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")) {
// 			uid
// 			block
// 			hash
// 			locktime
// 			inputs {
// 				hash
// 				vout
// 			}
// 			outputs {
// 				value
// 				vout
// 			}
// 		}
// 	}`)
// 	assert.Equal(suite.T(), err, nil)
// 	var body Resp
// 	if err := json.Unmarshal(resp.GetJson(), &body); err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println("Deleting field with uid", body.Q[0].UID)
// 	out, err := json.Marshal(map[string]string{"uid": body.Q[0].UID})
// 	assert.Equal(suite.T(), err, nil)
// 	_, err2 := suite.dgraph.NewTxn().Mutate(context.Background(), &api.Mutation{DeleteJson: out, CommitNow: true})
// 	assert.Equal(suite.T(), err2, nil)
// }

// func (suite *TestTransactionsSuite) TestGetAddressOccurences() {
// 	addr := "12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S"
// 	address, err := btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
// 	assert.Equal(suite.T(), err, nil)
// 	occurences, err := GetAddressOccurences(&address)
// 	assert.Equal(suite.T(), err, nil)
// 	assert.Equal(suite.T(), len(occurences), 6)
// }

func TestDgraphTransactions(t *testing.T) {
	suite.Run(t, new(TestTransactionsSuite))
}
