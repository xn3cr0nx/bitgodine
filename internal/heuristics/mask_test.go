package heuristics

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/assert.v1"
)

type TestMaskSuite struct {
	suite.Suite
}

// func (suite *TestMaskSuite) SetupSuite() {
// }

// func (suite *TestMaskSuite) TearDownSuite() {
// }

func (suite *TestMaskSuite) TestVulnerableMask() {
	m := Mask([3]byte{byte(1), byte(0), byte(0)})
	res := m.VulnerableMask(0)
	assert.Equal(suite.T(), res, true)
}

func (suite *TestMaskSuite) TestMergeMask() {
	m1 := Mask([3]byte{byte(1), byte(15), byte(0)})
	m2 := Mask([3]byte{byte(0), byte(4), byte(10)})
	merged := MergeMasks(m1, m2)
	assert.Equal(suite.T(), merged, Mask([3]byte{byte(1), byte(15), byte(10)}))
}

func (suite *TestMaskSuite) TestMaskFromPower() {
	m := MaskFromPower(0)
	assert.Equal(suite.T(), m, Mask([3]byte{byte(1), byte(0), byte(0)}))
	m = MaskFromPower(15)
	assert.Equal(suite.T(), m, Mask([3]byte{byte(0), byte(128), byte(0)}))
}

func (suite *TestMaskSuite) TestSum() {
	m1 := Mask([3]byte{byte(1), byte(15), byte(0)})
	m2 := Mask([3]byte{byte(0), byte(4), byte(10)})
	merged := m1.Sum(m2)
	assert.Equal(suite.T(), merged, Mask([3]byte{byte(1), byte(19), byte(10)}))
}

func (suite *TestMaskSuite) TestToList() {
	m := Mask([3]byte{byte(17), byte(0), byte(0)})
	list := m.ToList()
	assert.Equal(suite.T(), list, []Heuristic{Locktime, AddressType})
}

func (suite *TestMaskSuite) TestToHeuristicsList() {
	m := Mask([3]byte{byte(17), byte(0), byte(0)})
	list := m.ToHeuristicsList()
	assert.Equal(suite.T(), list, []string{Locktime.String(), AddressType.String()})
}

func (suite *TestMaskSuite) TestFromListToMask() {
	list := []Heuristic{Locktime, AddressType}
	m := FromListToMask(list)
	assert.Equal(suite.T(), m, Mask([3]byte{byte(17), byte(0), byte(0)}))
}

func (suite *TestMaskSuite) TestIsCoinbase() {
	m := MaskFromPower(Coinbase)
	assert.Equal(suite.T(), m.IsCoinbase(), true)
}

func (suite *TestMaskSuite) TestIsSelfTransfer() {
	m := MaskFromPower(SelfTransfer)
	assert.Equal(suite.T(), m.IsSelfTransfer(), true)
}

func (suite *TestMaskSuite) TestIsOffByOneBug() {
	m := MaskFromPower(OffByOne)
	assert.Equal(suite.T(), m.IsOffByOneBug(), true)
}

func (suite *TestMaskSuite) TestIsPeelingLike() {
	m := MaskFromPower(PeelingLike)
	assert.Equal(suite.T(), m.IsPeelingLike(), true)
}

func TestMask(t *testing.T) {
	suite.Run(t, new(TestMaskSuite))
}
