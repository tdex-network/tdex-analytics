package domain

import (
	"context"
	"time"
)

type MarketPriceRepository interface {
	InsertPrice(ctx context.Context, price MarketPrice) error
	GetPricesForMarket(ctx context.Context, marketID string, fromTime time.Time) ([]MarketPrice, error)
	GetPricesForAllMarkets(ctx context.Context, fromTime time.Time) (map[string][]MarketPrice, error)
}
