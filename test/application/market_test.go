package influxdbtest

import (
	"context"
	"tdex-analytics/internal/core/application"
)

func (a *AppSvcTestSuit) TestGetMarketsForFilter() {
	markets, err := marketSvc.ListMarketIDs(
		context.Background(),
		[]application.MarketRequest{},
	)
	if err != nil {
		a.FailNow(err.Error())
	}

	a.Equal(6, len(markets))

	markets, err = marketSvc.ListMarketIDs(
		context.Background(),
		[]application.MarketRequest{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
		},
	)
	a.Equal(
		"BaseAsset: asset is not in hex format; QuoteAsset: asset is not in hex format; Url: must be a valid URL.",
		err.Error(),
	)
}
