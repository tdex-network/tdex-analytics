package domain

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type MarketPriceRepository interface {
	InsertPrice(ctx context.Context, price MarketPrice) error
	GetPricesForMarkets(
		ctx context.Context,
		startTime time.Time,
		endTime time.Time,
		page Page,
		groupBy string,
		marketIDs ...string,
	) (map[string][]MarketPrice, error)
	CalculateVWAP(
		ctx context.Context,
		averageWindow string,
		startTime time.Time,
		endTime time.Time,
		marketIDs ...string,
	) (decimal.Decimal, error)
}
