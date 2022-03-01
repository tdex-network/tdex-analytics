package pgtest

import (
	"context"
	"tdex-analytics/internal/core/domain"
)

func (s *PgDbTestSuite) TestInsertAndFetchMarket() {
	if err := pgDbSvc.InsertMarket(context.Background(), domain.Market{
		ProviderName: "test1",
		Url:          "test1",
		BaseAsset:    "test1",
		QuoteAsset:   "test1",
	}); err != nil {
		s.FailNow(err.Error())
	}

	if err := pgDbSvc.InsertMarket(context.Background(), domain.Market{
		ProviderName: "test1",
		Url:          "test1",
		BaseAsset:    "test1",
		QuoteAsset:   "test1",
	}); err != nil {
		s.FailNow(err.Error())
	}

	markets, err := pgDbSvc.GetAllMarkets(context.Background())
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(1, len(markets))
}
