package influxdbtest

import (
	"github.com/stretchr/testify/suite"
	dbinflux "github.com/tdex-network/tdex-analytics/internal/infrastructure/db/influx"
	"os"
)

var (
	dbSvc dbinflux.Service
)

type InfluxDBTestSuit struct {
	suite.Suite
}

func (idb *InfluxDBTestSuit) SetupSuite() {
	token := os.Getenv("TDEXA_INFLUXDB_TOKEN")
	if token == "" {
		idb.FailNow("TDEXA_INFLUXDB_TOKEN not set")
	}

	db, err := dbinflux.New(dbinflux.Config{
		Org:             "tdex-network",
		AuthToken:       token,
		DbUrl:           "http://localhost:8086",
		AnalyticsBucket: "analytics",
	})
	if err != nil {
		idb.FailNow(err.Error())
	}

	dbSvc = db
}

func (idb *InfluxDBTestSuit) TearDownSuite() {
	dbSvc.Close()
}

func (idb *InfluxDBTestSuit) BeforeTest(suiteName, testName string) {}

func (idb *InfluxDBTestSuit) AfterTest(suiteName, testName string) {}
