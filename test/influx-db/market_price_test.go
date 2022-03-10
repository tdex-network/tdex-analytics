package influxdbtest

import (
	"context"
	"tdex-analytics/internal/core/application"
	"tdex-analytics/internal/core/domain"
	"time"
)

func (idb *InfluxDBTestSuit) TestInsertMarketPrice() {
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertPrice(ctx, domain.MarketPrice{
			MarketID:   "213",
			BasePrice:  float32(50 + i),
			BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuotePrice: float32(500 + i),
			QuoteAsset: "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:       time.Now(),
		}); err != nil {
			idb.FailNow(err.Error())
		}
	}
}

func (idb *InfluxDBTestSuit) TestGetMarketPrice() {
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := dbSvc.InsertPrice(ctx, domain.MarketPrice{
			MarketID:   "999",
			BasePrice:  float32(50 + i),
			BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuotePrice: float32(500 + i),
			QuoteAsset: "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:       time.Now(),
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

	marketsPrices, err := dbSvc.GetPricesForMarkets(
		ctx,
		startTime,
		endTime,
		page,
		[]string{"999"}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(10, len(marketsPrices["999"]))
}

func (idb *InfluxDBTestSuit) TestGetMarketPriceWithPagination() {
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

	marketsPrices, err := dbSvc.GetPricesForMarkets(
		context.Background(),
		startTime,
		endTime,
		page,
		[]string{"1"}...,
	)
	if err != nil {
		idb.FailNow(err.Error())
	}

	idb.Equal(5, len(marketsPrices["1"]))
}
