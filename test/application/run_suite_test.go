package influxdbtest

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAppSvcTestSuit(t *testing.T) {
	suite.Run(t, new(AppSvcTestSuit))
}
