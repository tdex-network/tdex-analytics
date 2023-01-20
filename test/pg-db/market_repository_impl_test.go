package pgtest

import (
	"context"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
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

	// check if duplicate will be inserted
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

	s.Equal(7, len(markets)) //fixtures + new one
}

func (s *PgDbTestSuite) TestGetMarketsForFilter() {
	page := domain.Page{
		Number: 1,
		Size:   10,
	}

	markets, err := pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		nil,
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(6, len(markets))

	markets, err = pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		[]domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
			{
				Url:        "dummyurl2",
				BaseAsset:  "dummybaseasset2",
				QuoteAsset: "dummyquoteasset2",
			},
		},
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(2, len(markets))

	markets, err = pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		[]domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
		},
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(1, len(markets))

	markets, err = pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		[]domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
			{
				Url:        "dummyurl3",
				BaseAsset:  "dummybaseasset3",
				QuoteAsset: "dummyquoteasset3",
			},
			{
				Url:        "dummyurl4",
				BaseAsset:  "dummybaseasset4",
				QuoteAsset: "dummyquoteasset4",
			},
		},
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(3, len(markets))

	markets, err = pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		[]domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
			{
				Url:        "dummyurl3",
				BaseAsset:  "dummybaseasset3",
				QuoteAsset: "dummyquoteasset3",
			},
			{
				Url:        "dummyurl4",
				BaseAsset:  "dummybaseasset3",
				QuoteAsset: "dummyquoteasset2",
			},
		},
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(2, len(markets))
}

func (s *PgDbTestSuite) TestGetMarketsForFilterWithPagination() {
	page := domain.Page{
		Number: 1,
		Size:   1,
	}

	markets, err := pgDbSvc.GetAllMarketsForFilter(
		context.Background(),
		[]domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
			{
				Url:        "dummyurl3",
				BaseAsset:  "dummybaseasset3",
				QuoteAsset: "dummyquoteasset3",
			},
			{
				Url:        "dummyurl4",
				BaseAsset:  "dummybaseasset3",
				QuoteAsset: "dummyquoteasset2",
			},
		},
		page,
	)
	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(1, len(markets))
}

func (s *PgDbTestSuite) TestActivateInactivateMarket() {
	var (
		filter1 = []domain.Filter{
			{
				Url:        "dummyurl1",
				BaseAsset:  "dummybaseasset1",
				QuoteAsset: "dummyquoteasset1",
			},
		}
		page = domain.Page{
			Number: 1,
			Size:   20,
		}
	)
	markets, err := pgDbSvc.GetAllMarkets(context.Background())
	s.NoError(err)

	activeCount := 0
	for _, v := range markets {
		if v.Active {
			activeCount++
		}
	}

	s.Equal(2, activeCount)

	err = pgDbSvc.ActivateMarket(context.Background(), 1)
	s.NoError(err)

	market, err := pgDbSvc.GetAllMarketsForFilter(context.Background(), filter1, page)
	s.NoError(err)
	s.Equal(true, market[0].Active)

	err = pgDbSvc.InactivateMarket(context.Background(), 1)
	s.NoError(err)

	market, err = pgDbSvc.GetAllMarketsForFilter(context.Background(), filter1, page)
	s.NoError(err)
	s.Equal(false, market[0].Active)

	activeMarkets, err := pgDbSvc.GetMarketsForActiveIndicator(context.Background(), true)
	s.NoError(err)
	s.Equal(2, len(activeMarkets))

	inactiveMarkets, err := pgDbSvc.GetMarketsForActiveIndicator(context.Background(), false)
	s.NoError(err)
	s.Equal(4, len(inactiveMarkets))
}
