package influxdbtest

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestInfluxDBTestSuite(t *testing.T) {
	suite.Run(t, new(AppSvcTestSuit))
}
