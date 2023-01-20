package domain

import "context"

type MarketRepository interface {
	InsertMarket(ctx context.Context, market Market) error
	GetAllMarkets(ctx context.Context) ([]Market, error)
	GetMarketsForActiveIndicator(ctx context.Context, active bool) ([]Market, error)
	GetAllMarketsForFilter(
		ctx context.Context,
		filter []Filter,
		page Page,
	) ([]Market, error)
	ActivateMarket(ctx context.Context, marketID int) error
	InactivateMarket(ctx context.Context, marketID int) error
}
