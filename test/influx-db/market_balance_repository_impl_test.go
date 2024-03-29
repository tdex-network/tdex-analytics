package influxdbtest

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	dbinflux "github.com/tdex-network/tdex-analytics/internal/infrastructure/db/influx"
	"os"
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
			BaseBalance:  decimal.NewFromInt(int64(50 + i)),
			QuoteBalance: decimal.NewFromInt(int64(500 + i)),
			Time:         time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}
}

func (idb *InfluxDBTestSuit) TestGetMarketBalance() {
	ctx := context.Background()
	marketID := "90000"
	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertBalance(ctx, domain.MarketBalance{
			MarketID:     marketID,
			BaseBalance:  decimal.NewFromInt(int64(50 + i)),
			QuoteBalance: decimal.NewFromInt(int64(500 + i)),
			Time:         time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}

	startTime := time.Date(
		currentYear,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		currentYear,
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
		"1mo",
		[]string{marketID}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(10, len(marketsBalances[marketID]))
}

func (idb *InfluxDBTestSuit) TestGetMarketBalanceWithPagination() {
	startTime := time.Date(
		currentYear,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		currentYear,
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
		"1mo",
		[]string{"1"}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(5, len(marketsBalances["1"]))
}

func (idb *InfluxDBTestSuit) TestGetMarketBalanceWithoutGroupBy() {
	ctx := context.Background()

	marketID := "90000"
	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertBalance(ctx, domain.MarketBalance{
			MarketID:     marketID,
			BaseBalance:  decimal.NewFromInt(int64(50 + i)),
			QuoteBalance: decimal.NewFromInt(int64(500 + i)),
			Time:         time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}

	startTime := time.Date(
		currentYear,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		currentYear,
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
		"",
		[]string{marketID}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(10, len(marketsBalances[marketID]))
}
