package influxdbtest

import (
	"context"
	"os"
	"tdex-analytics/internal/core/application"
	"tdex-analytics/internal/core/domain"
	dbinflux "tdex-analytics/internal/infrastructure/db/influx"
	"time"
)

func (idb *InfluxDBTestSuit) TestInsertMarketBalance() {
	ctx := context.Background()

	token := os.Getenv("TDEXA_INFLUXDB_TOKEN")

	db, err := dbinflux.New(dbinflux.Config{
		Org:             "tdex-network",
		AuthToken:       token,
		DbUrl:           "http://localhost:8086",
		AnalyticsBucket: "analytics",
	})
	if err != nil {
		idb.FailNow(err.Error())
	}

	for i := 0; i < 10; i++ {
		if err := db.InsertBalance(ctx, domain.MarketBalance{
			MarketID:     "2203",
			BaseBalance:  50 + i,
			BaseAsset:    "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuoteBalance: 500 + i,
			QuoteAsset:   "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:         time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}
}

func (idb *InfluxDBTestSuit) TestGetMarketBalance() {
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertBalance(ctx, domain.MarketBalance{
			MarketID:     "90000",
			BaseBalance:  50 + i,
			BaseAsset:    "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuoteBalance: 500 + i,
			QuoteAsset:   "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:         time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}

	startTime := time.Date(
		application.StartYear,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		application.StartYear,
		12,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	page := domain.Page{
		Number: 1,
		Size:   10,
	}

	marketsBalances, err := dbSvc.GetBalancesForMarkets(
		ctx,
		startTime,
		endTime,
		page,
		[]string{"90000"}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(10, len(marketsBalances["90000"]))
}

func (idb *InfluxDBTestSuit) TestGetMarketBalanceWithPagination() {
	startTime := time.Date(
		application.StartYear,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		application.StartYear,
		12,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	page := domain.Page{
		Number: 1,
		Size:   5,
	}

	marketsBalances, err := dbSvc.GetBalancesForMarkets(
		context.Background(),
		startTime,
		endTime,
		page,
		[]string{"1"}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(5, len(marketsBalances["1"]))
}
