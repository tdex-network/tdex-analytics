package influxdbtest

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"time"
)

func (idb *InfluxDBTestSuit) TestCalcVWAP() {
	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)
	marketIDs := []string{"78"}
	wamp, err := dbSvc.CalculateVWAP(ctx, "5s", startTime, endTime, marketIDs...)
	idb.NoError(err)
	idb.Equal(wamp.Round(2), decimal.NewFromFloat(30592.04))
}

func (idb *InfluxDBTestSuit) TestInsertMarketPrice() {
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertPrice(ctx, domain.MarketPrice{
			MarketID:   "213",
			BasePrice:  decimal.NewFromInt(int64(50 + i)),
			BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuotePrice: decimal.NewFromInt(int64(500 + i)),
			QuoteAsset: "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:       time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}
}

func (idb *InfluxDBTestSuit) TestGetMarketPrice() {
	ctx := context.Background()

	marketID := "999"
	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertPrice(ctx, domain.MarketPrice{
			MarketID:   marketID,
			BasePrice:  decimal.NewFromInt(int64(50 + i)),
			BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuotePrice: decimal.NewFromInt(int64(500 + i)),
			QuoteAsset: "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:       time.Now(),
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

	marketsPrices, err := dbSvc.GetPricesForMarkets(
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

	idb.Equal(10, len(marketsPrices[marketID]))
}

func (idb *InfluxDBTestSuit) TestGetMarketPriceWithPagination() {
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

	marketsPrices, err := dbSvc.GetPricesForMarkets(
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

	idb.Equal(5, len(marketsPrices["1"]))
}

func (idb *InfluxDBTestSuit) TestGetMarketPriceWithoutGrouping() {
	ctx := context.Background()

	marketID := "999"
	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertPrice(ctx, domain.MarketPrice{
			MarketID:   marketID,
			BasePrice:  decimal.NewFromInt(int64(50 + i)),
			BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuotePrice: decimal.NewFromInt(int64(500 + i)),
			QuoteAsset: "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:       time.Now(),
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

	marketsPrices, err := dbSvc.GetPricesForMarkets(
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

	idb.Equal(10, len(marketsPrices[marketID]))
}
