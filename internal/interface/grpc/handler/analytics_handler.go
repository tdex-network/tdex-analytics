package grpchandler

import (
	"context"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/tdexa/v1"
	"tdex-analytics/internal/core/application"
)

type analyticsHandler struct {
	tdexav1.UnimplementedAnalyticsServer
	marketBalanceSvc application.MarketBalanceService
	marketPriceSvc   application.MarketPriceService
	marketSvc        application.MarketService
}

func NewAnalyticsHandler(
	marketBalanceSvc application.MarketBalanceService,
	marketPriceSvc application.MarketPriceService,
	marketSvc application.MarketService,
) tdexav1.AnalyticsServer {
	return &analyticsHandler{
		marketBalanceSvc: marketBalanceSvc,
		marketPriceSvc:   marketPriceSvc,
		marketSvc:        marketSvc,
	}
}

func (a *analyticsHandler) MarketsBalances(
	ctx context.Context,
	req *tdexav1.MarketsBalancesRequest,
) (*tdexav1.MarketsBalancesReply, error) {
	mb, err := a.marketBalanceSvc.GetBalances(
		ctx,
		grpcTimeRangeToAppTimeRange(req.GetTimeRange()),
		parsePage(req.GetPage()),
		req.GetMarketIds()...,
	)
	if err != nil {
		return nil, err
	}

	marketsBalances := make(map[string]*tdexav1.MarketBalances)

	for k, v := range mb.MarketsBalances {
		marketBalances := make([]*tdexav1.MarketBalance, 0)
		for _, v1 := range v {
			marketBalances = append(marketBalances, &tdexav1.MarketBalance{
				BaseBalance:  int64(v1.BaseBalance),
				QuoteBalance: int64(v1.QuoteBalance),
				Time:         v1.Time.String(),
			})
		}
		marketsBalances[k] = &tdexav1.MarketBalances{
			MarketBalance: marketBalances,
		}
	}

	return &tdexav1.MarketsBalancesReply{
		MarketsBalances: marketsBalances,
	}, nil
}

func (a *analyticsHandler) MarketsPrices(
	ctx context.Context,
	req *tdexav1.MarketsPricesRequest,
) (*tdexav1.MarketsPricesReply, error) {
	mb, err := a.marketPriceSvc.GetPrices(
		ctx,
		grpcTimeRangeToAppTimeRange(req.GetTimeRange()),
		parsePage(req.GetPage()),
		req.GetReferenceCurrency(),
		req.GetMarketIds()...,
	)
	if err != nil {
		return nil, err
	}

	marketsPrices := make(map[string]*tdexav1.MarketPrices)

	for k, v := range mb.MarketsPrices {
		marketPrices := make([]*tdexav1.MarketPrice, 0)
		for _, v1 := range v {
			basePrice, _ := v1.BasePrice.Float64()
			BaseReferencePrice, _ := v1.BaseReferentPrice.Float64()
			quotePrice, _ := v1.QuotePrice.Float64()
			quoteReferencePrice, _ := v1.QuoteReferentPrice.Float64()
			marketPrices = append(marketPrices, &tdexav1.MarketPrice{
				BasePrice:           basePrice,
				BaseReferencePrice:  BaseReferencePrice,
				QuotePrice:          quotePrice,
				QuoteReferencePrice: quoteReferencePrice,
				Time:                v1.Time.String(),
			})
		}
		marketsPrices[k] = &tdexav1.MarketPrices{
			MarketPrice: marketPrices,
		}
	}

	return &tdexav1.MarketsPricesReply{
		MarketsPrices: marketsPrices,
	}, nil
}

func (a *analyticsHandler) ListMarkets(
	ctx context.Context,
	req *tdexav1.ListMarketsRequest,
) (*tdexav1.ListMarketsReply, error) {
	r := make([]application.MarketProvider, 0)
	for _, v := range req.GetMarketProviders() {
		r = append(r, application.MarketProvider{
			Url:        v.GetUrl(),
			BaseAsset:  v.GetBaseAsset(),
			QuoteAsset: v.GetQuoteAsset(),
		})
	}
	markets, err := a.marketSvc.ListMarkets(ctx, r, parsePage(req.GetPage()))
	if err != nil {
		return nil, err
	}

	resp := make([]*tdexav1.MarketIDInfo, 0)
	for _, v := range markets {
		resp = append(resp, &tdexav1.MarketIDInfo{
			Id: uint64(v.ID),
			MarketProvider: &tdexav1.MarketProvider{
				Url:        v.Url,
				BaseAsset:  v.BaseAsset,
				QuoteAsset: v.QuoteAsset,
			},
		})
	}

	return &tdexav1.ListMarketsReply{
		Markets: resp,
	}, nil
}

func grpcTimeRangeToAppTimeRange(timeRange *tdexav1.TimeRange) application.TimeRange {
	var predefinedPeriod *application.PredefinedPeriod
	if timeRange.GetPredefinedPeriod() > tdexav1.PredefinedPeriod_NULL {
		pp := application.PredefinedPeriod(timeRange.GetPredefinedPeriod())
		predefinedPeriod = &pp
	}

	var customPeriod *application.CustomPeriod
	if timeRange.GetCustomPeriod() != nil {
		customPeriod = &application.CustomPeriod{
			StartDate: timeRange.GetCustomPeriod().GetStartDate(),
			EndDate:   timeRange.GetCustomPeriod().GetEndDate(),
		}
	}

	return application.TimeRange{
		PredefinedPeriod: predefinedPeriod,
		CustomPeriod:     customPeriod,
	}
}

func parsePage(p *tdexav1.Page) application.Page {
	if p == nil {
		return application.Page{
			Number: 0,
			Size:   0,
		}
	}
	return application.Page{
		Number: int(p.PageNumber),
		Size:   int(p.PageSize),
	}
}
