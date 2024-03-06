package applicationtest

import (
	"context"
	"os"

	"github.com/tdex-network/tdex-analytics/internal/core/application"
	dbinflux "github.com/tdex-network/tdex-analytics/internal/infrastructure/db/influx"
	dbpg "github.com/tdex-network/tdex-analytics/internal/infrastructure/db/pg"
	"github.com/tdex-network/tdex-analytics/pkg/rater"
	tdexmarketloader "github.com/tdex-network/tdex-analytics/pkg/tdex-market-loader"

	"github.com/stretchr/testify/suite"
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

	raterSvc, err := rater.NewExchangeRateClient(map[string]string{
		"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225": "bitcoin",
		"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d": "bitcoin",
	},
		nil,
		nil,
		nil,
	)
	if err != nil {
		a.FailNow(err.Error())
	}

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
		raterSvc,
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
