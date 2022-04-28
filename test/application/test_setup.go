package influxdbtest

import (
	"context"
	"github.com/stretchr/testify/suite"
	"os"
	"tdex-analytics/internal/core/application"
	dbinflux "tdex-analytics/internal/infrastructure/db/influx"
	dbpg "tdex-analytics/internal/infrastructure/db/pg"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
)

var (
	influxDbSvc       dbinflux.Service
	marketBalanceSvc  application.MarketBalanceService
	marketPriceSvc    application.MarketPriceService
	marketLoaderSvc   application.MarketsLoaderService
	marketSvc         application.MarketService
	marketRepository  dbpg.Service
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

	influxDbSvc = db

	mr, err := dbpg.New(dbpg.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "tdexa-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/db/pg/migrations",
		DbInsecure: true,
	})
	if err != nil {
		a.FailNow(err.Error())
	}

	if mr != nil {
		err := mr.CreateLoader("../fixtures")
		if err != nil {
			a.FailNow(err.Error())
		}
	}
	marketRepository = mr

	tdexMarketLoaderSvc := tdexmarketloader.NewService(
		"127.0.0.1:9050",
		"https://api.github.com/repos/tdex-network/tdex-registry/contents/registry.json",
		1000,
	)

	marketLoaderSvc = application.NewMarketsLoaderService(
		marketRepository,
		tdexMarketLoaderSvc,
	)
	marketBalanceSvc = application.NewMarketBalanceService(
		influxDbSvc,
		marketRepository,
		tdexMarketLoaderSvc,
		"5",
	)
	marketPriceSvc = application.NewMarketPriceService(
		influxDbSvc,
		marketRepository,
		tdexMarketLoaderSvc,
		"5",
		nil, //TODO
	)
	marketSvc = application.NewMarketService(marketRepository)
}

func (a *AppSvcTestSuit) TearDownSuite() {
	influxDbSvc.Close()
}

func (a *AppSvcTestSuit) BeforeTest(suiteName, testName string) {
	if err := marketRepository.LoadFixtures(); err != nil {
		a.FailNow(err.Error())
	}
}

func (a *AppSvcTestSuit) AfterTest(suiteName, testName string) {}
