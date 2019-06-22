package heuristics

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/assert.v1"
)

type TestHeuristicsSuite struct {
	suite.Suite
}

func (suite *TestHeuristicsSuite) SetupSuite() {
}

func (suite *TestHeuristicsSuite) TearDownSuite() {
}

func (suite *TestHeuristicsSuite) TestSetCardinality() {
	assert.Equal(suite.T(), SetCardinality(), 8)	
}


func TestHeuristics(t *testing.T) {
	suite.Run(t, new(TestHeuristicsSuite))
}
