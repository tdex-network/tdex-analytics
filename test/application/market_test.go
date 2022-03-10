package influxdbtest

import (
	"context"
	"tdex-analytics/internal/core/application"
)

func (a *AppSvcTestSuit) TestGetMarketsForFilter() {
	page := application.Page{
		Number: 1,
		Size:   10,
	}

	markets, err := marketSvc.ListMarkets(
		context.Background(),
		[]application.MarketProvider{},
		page,
	)
	if err != nil {
		a.FailNow(err.Error())
	}

	a.Equal(6, len(markets))

	markets, err = marketSvc.ListMarkets(
		context.Background(),
		[]application.MarketProvider{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
		},
		page,
	)
	a.Equal(
		"BaseAsset: asset is not in hex format; QuoteAsset: asset is not in hex format; Url: must be a valid URL.",
		err.Error(),
	)
}
