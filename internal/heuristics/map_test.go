package heuristics

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/assert.v1"
)

type TestMapSuite struct {
	suite.Suite
}

// func (suite *TestMapSuite) SetupSuite() {
// }

// func (suite *TestMapSuite) TearDownSuite() {
// }

func (suite *TestMapSuite) TestMapFromHeuristics() {
	m := MapFromHeuristics(Locktime)
	assert.Equal(suite.T(), m[Locktime], uint32(1))
}

func (suite *TestMapSuite) TestMergeMap() {
	m1 := MapFromHeuristics(Peeling)
	m2 := MapFromHeuristics(Locktime)
	merged := MergeMaps(m1, m2)
	assert.Equal(suite.T(), merged, MapFromHeuristics(Peeling, Locktime))
}

func (suite *TestMapSuite) TestToList() {
	m := MapFromHeuristics(Locktime, AddressType)
	list := m.ToList()
	assert.Equal(suite.T(), list, []Heuristic{Locktime, AddressType})
}

func (suite *TestMapSuite) TestToHeuristicsList() {
	m := MapFromHeuristics(Locktime, AddressType)
	list := m.ToHeuristicsList()
	assert.Equal(suite.T(), list, []string{Locktime.String(), AddressType.String()})
}

func (suite *TestMapSuite) TestIsCoinbase() {
	m := MapFromHeuristics(Coinbase)
	assert.Equal(suite.T(), m.IsCoinbase(), true)
}

func (suite *TestMapSuite) TestIsSelfTransfer() {
	m := MapFromHeuristics(SelfTransfer)
	assert.Equal(suite.T(), m.IsSelfTransfer(), true)
}

func (suite *TestMapSuite) TestIsOffByOneBug() {
	m := MapFromHeuristics(OffByOne)
	assert.Equal(suite.T(), m.IsOffByOneBug(), true)
}

func (suite *TestMapSuite) TestIsPeelingLike() {
	m := MapFromHeuristics(PeelingLike)
	assert.Equal(suite.T(), m.IsPeelingLike(), true)
}

func TestMap(t *testing.T) {
	suite.Run(t, new(TestMapSuite))
}
