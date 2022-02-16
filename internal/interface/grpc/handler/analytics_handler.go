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
		int(req.GetMarketId()),
		req.GetFromTime(),
	)
	if err != nil {
		return nil, err
	}

	marketsBalances := make(map[int32]*tdexav1.MarketBalances)

	for k, v := range mb.MarketsBalances {
		marketBalances := make([]*tdexav1.MarketBalance, 0)
		for _, v1 := range v {
			marketBalances = append(marketBalances, &tdexav1.MarketBalance{
				BaseBalance:  int64(v1.BaseBalance),
				QuoteBalance: int64(v1.QuoteBalance),
				Time:         v1.Time.String(),
			})
		}
		marketsBalances[int32(k)] = &tdexav1.MarketBalances{
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
		int(req.GetMarketId()),
		req.GetFromTime(),
	)
	if err != nil {
		return nil, err
	}

	marketsPrices := make(map[int32]*tdexav1.MarketPrices)

	for k, v := range mb.MarketsPrices {
		marketPrices := make([]*tdexav1.MarketPrice, 0)
		for _, v1 := range v {
			marketPrices = append(marketPrices, &tdexav1.MarketPrice{
				BasePrice:  int64(v1.BasePrice),
				QuotePrice: int64(v1.QuotePrice),
				Time:       v1.Time.String(),
			})
		}
		marketsPrices[int32(k)] = &tdexav1.MarketPrices{
			MarketPrice: marketPrices,
		}
	}

	return &tdexav1.MarketsPricesReply{
		MarketsPrices: marketsPrices,
	}, nil
}
