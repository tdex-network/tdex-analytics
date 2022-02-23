package port

import (
	"context"
	"tdex-analytics/internal/core/domain"
)

type MarketDataFetcher interface {
	FetchBalance(ctx context.Context, market domain.Market) (Balance, error)
	FetchPrice(ctx context.Context, market domain.Market) (Price, error)
}

type Balance struct {
	BaseBalance  int
	QuoteBalance int
}

type Price struct {
	BasePrice  int
	QuotePrice int
}
