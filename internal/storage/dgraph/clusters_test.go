package dgraph_test

import (
	"github.com/xn3cr0nx/bitgodine/internal/storage/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dgraph clusters", func() {
	var (
		dg *dgraph.Dgraph
	)

	BeforeEach(func() {
		logger.Setup()

		bigCache, err := cache.NewCache(nil)
		Expect(err).ToNot(HaveOccurred())

		dg = dgraph.Instance(dgraph.Conf(), bigCache)
		err = dg.Setup()
		Expect(err).ToNot(HaveOccurred())

	})

	AfterEach(func() {
		err := dg.Empty()
		Expect(err).ToNot(HaveOccurred())
	})
})

// func (suite *TestDgraphClustersSuite) Setup() {
// 	clusters := Clusters{
// 		Size: 2,
// 		Parents: []Parent{
// 			{
// 				Pos:    0,
// 				Parent: 1,
// 			},
// 			{
// 				Pos:    1,
// 				Parent: 2,
// 			},
// 		},
// 		Ranks: []Rank{
// 			{
// 				Pos:  0,
// 				Rank: 1,
// 			},
// 			{
// 				Pos:  1,
// 				Rank: 2,
// 			},
// 		},
// 		Set: []Cluster{
// 			{
// 				Addresses: []Address{
// 					{Address: "1BoatSLRHtKNngkdXEeobR76b53LETtpyT"},
// 					{Address: "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy"},
// 				},
// 				Cluster: 0,
// 			},
// 		},
// 	}
// 	err := Store(clusters)
// 	assert.Equal(suite.T(), err, nil)

// 	cluster, err := GetClusters()
// 	assert.Equal(suite.T(), err, nil)
// 	suite.cluster.UID = cluster.UID
// 	suite.cluster.Size = cluster.Size
// 	suite.cluster.Parents = cluster.Parents
// 	suite.cluster.Ranks = cluster.Ranks
// 	suite.cluster.Set = cluster.Set
// }

// func (suite *TestDgraphClustersSuite) TearDownSuite() {
// 	err := Delete(suite.cluster.UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Parents[0].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Parents[1].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Ranks[0].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Ranks[1].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[0].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[0].Addresses[0].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[0].Addresses[1].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[0].Addresses[2].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[1].UID)
// 	assert.Equal(suite.T(), err, nil)
// 	err = Delete(suite.cluster.Set[1].Addresses[0].UID)
// 	assert.Equal(suite.T(), err, nil)
// }

// func (suite *TestDgraphClustersSuite) TestGetClusterUID() {
// 	UID, err := GetClusterUID()
// 	assert.Equal(suite.T(), err, nil)
// 	assert.Equal(suite.T(), suite.cluster.UID, UID)
// }

// func (suite *TestDgraphClustersSuite) TestNewSet() {
// 	err := NewSet("1C84keNhdyCBZR7NTEV5mX6cAdVpf111mJ", 1)
// 	assert.Equal(suite.T(), err, nil)
// 	c, err := GetClusters()
// 	suite.cluster.Set = c.Set
// 	assert.Equal(suite.T(), err, nil)
// 	assert.Equal(suite.T(), len(c.Set), 2)
// }

// func (suite *TestDgraphClustersSuite) TestUpdateSet() {
// 	err := UpdateSet("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", 0)
// 	assert.Equal(suite.T(), err, nil)
// 	c, err := GetClusters()
// 	suite.cluster.Set = c.Set
// 	assert.Equal(suite.T(), err, nil)
// 	assert.Equal(suite.T(), len(c.Set[0].Addresses), 3)
// }

// func (suite *TestDgraphClustersSuite) TestUpdateSize() {
// 	err := UpdateSize(1)
// 	assert.Equal(suite.T(), err, nil)
// 	suite.cluster.Size = 1
// 	c, err := GetClusters()
// 	assert.Equal(suite.T(), err, nil)
// 	assert.Equal(suite.T(), c.Size, suite.cluster.Size)
// }
