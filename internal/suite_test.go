package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DedbSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func TestDedbSuite(t *testing.T) {
	s := DedbSuite{}
	suite.Run(t, &s)
}

func (suite *DedbSuite) SetupSuite() {

}
