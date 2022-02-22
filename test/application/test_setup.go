package influxdbtest

import (
	"context"
	"github.com/stretchr/testify/suite"
	"os"
	"tdex-analytics/internal/core/application"
	dbinflux "tdex-analytics/internal/infrastructure/db/influx"
)

var (
	dbSvc             dbinflux.Service
	marketBalanceSvc  application.MarketBalanceService
	marketPriceSvc    application.MarketPriceService
	ctx               = context.Background()
	nilPp             = application.NIL
	lastHourPp        = application.LastHour
	lastDayPp         = application.LastDay
	lastMonthPp       = application.LastMonth
	lastThreeMonthsPp = application.LastThreeMonths
	yearToDatePp      = application.YearToDate
	allPp             = application.All
)

type AppSvcTestSuit struct {
	suite.Suite
}

func (a *AppSvcTestSuit) SetupSuite() {
	token := os.Getenv("TDEXA_INFLUXDB_TOKEN")
	if token == "" {
		a.FailNow("TDEXA_INFLUXDB_TOKEN not set")
	}

	db, err := dbinflux.New(dbinflux.Config{
		Org:             "tdex-network",
		AuthToken:       token,
		DbUrl:           "http://localhost:8086",
		AnalyticsBucket: "analytics",
	})
	if err != nil {
		a.FailNow(err.Error())
	}

	dbSvc = db

	marketBalanceSvc = application.NewMarketBalanceService(dbSvc)
	marketPriceSvc = application.NewMarketPriceService(dbSvc)
}

func (a *AppSvcTestSuit) TearDownSuite() {
	dbSvc.Close()
}

func (a *AppSvcTestSuit) BeforeTest(suiteName, testName string) {}

func (a *AppSvcTestSuit) AfterTest(suiteName, testName string) {}
