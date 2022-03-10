package domain

import "context"

type MarketRepository interface {
	InsertMarket(ctx context.Context, market Market) error
	GetAllMarkets(ctx context.Context) ([]Market, error)
	GetAllMarketsForFilter(
		ctx context.Context,
		filter []Filter,
		page Page,
	) ([]Market, error)
}
