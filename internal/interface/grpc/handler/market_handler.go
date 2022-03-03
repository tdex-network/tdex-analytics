package grpchandler

import (
	"context"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/v1"
	"tdex-analytics/internal/core/application"
)

type marketHandler struct {
	tdexav1.UnimplementedMarketServer
	marketSvc application.MarketService
}

func NewMarketHandler(marketSvc application.MarketService) tdexav1.MarketServer {
	return &marketHandler{
		marketSvc: marketSvc,
	}
}

func (m *marketHandler) ListMarketIDs(
	ctx context.Context,
	req *tdexav1.ListMarketIDsRequest,
) (*tdexav1.ListMarketIDsReply, error) {
	r := make([]application.MarketRequest, 0)
	for _, v := range req.GetMarketsRequest() {
		r = append(r, application.MarketRequest{
			Url:        v.GetUrl(),
			BaseAsset:  v.GetBaseAsset(),
			QuoteAsset: v.GetQuoteAsset(),
		})
	}
	ids, err := m.marketSvc.ListMarketIDs(ctx, r)
	if err != nil {
		return nil, err
	}
	return &tdexav1.ListMarketIDsReply{
		Ids: ids,
	}, nil
}
