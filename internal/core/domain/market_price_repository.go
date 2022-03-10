package domain

import (
	"context"
	"time"
)

type MarketPriceRepository interface {
	InsertPrice(ctx context.Context, price MarketPrice) error
	GetPricesForMarkets(
		ctx context.Context,
		startTime time.Time,
		endTime time.Time,
		page Page,
		marketIDs ...string,
	) (map[string][]MarketPrice, error)
}
