package domain

import (
	"context"
	"time"
)

type MarketBalanceRepository interface {
	InsertBalance(ctx context.Context, balance MarketBalance) error
	GetBalancesForMarkets(
		ctx context.Context,
		startTime time.Time,
		endTime time.Time,
		page Page,
		marketIDs ...string,
	) (map[string][]MarketBalance, error)
}
