package grpchandler

import (
	"context"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/v1"
	"tdex-analytics/internal/core/application"
)

type analyticsHandler struct {
	tdexav1.UnimplementedAnalyticsServer
	marketBalanceSvc application.MarketBalanceService
	marketPriceSvc   application.MarketPriceService
}

func NewAnalyticsHandler(
	marketBalanceSvc application.MarketBalanceService,
	marketPriceSvc application.MarketPriceService,
) tdexav1.AnalyticsServer {
	return &analyticsHandler{
		marketBalanceSvc: marketBalanceSvc,
		marketPriceSvc:   marketPriceSvc,
	}
}

func (a *analyticsHandler) MarketsBalances(
	ctx context.Context,
	req *tdexav1.MarketsBalancesRequest,
) (*tdexav1.MarketsBalancesReply, error) {
	mb, err := a.marketBalanceSvc.GetBalances(
		ctx,
		grpcTimeRangeToAppTimeRange(req.GetTimeRange()),
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
		req.GetMarketIds()...,
	)
	if err != nil {
		return nil, err
	}

	marketsPrices := make(map[string]*tdexav1.MarketPrices)

	for k, v := range mb.MarketsPrices {
		marketPrices := make([]*tdexav1.MarketPrice, 0)
		for _, v1 := range v {
			marketPrices = append(marketPrices, &tdexav1.MarketPrice{
				BasePrice:  v1.BasePrice,
				QuotePrice: v1.QuotePrice,
				Time:       v1.Time.String(),
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
