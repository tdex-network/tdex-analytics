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