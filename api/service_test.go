package api

import (
	"testing"

	"github.com/nu7hatch/gouuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	ba Backend
	cr CreateRequest
}

// before each test
func (suite *ServiceTestSuite) SetupTest() {
	suite.ba = Backend{
		make(chan CreateRequest),
		make(chan CreateResult),
		make(chan ArtifactRequest),
		make(chan Artifact),
	}
	suite.cr = CreateRequest{}
}

// run testsuite on "go test"
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestDeploy() {
	cs, err := suite.ba.Deploy(&suite.cr)

	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), cs) {
		assert.NotEmpty(suite.T(), cs)
		id, _ := uuid.NewV4()
		assert.IsType(suite.T(), id, cs.Id)
	}
}

func (suite *ServiceTestSuite) TestPutDeploy() {
	cs, err := PutDeploy(&suite.cr)

	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), cs) {
		assert.NotEmpty(suite.T(), cs)
		id, _ := uuid.NewV4()
		assert.IsType(suite.T(), id, cs.Id)
	}
}
