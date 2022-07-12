package influxdbtest

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAppSvcTestSuite(t *testing.T) {
	suite.Run(t, new(AppSvcTestSuit))
}
