package domain

import (
	"context"
	"time"
)

type MarketBalanceRepository interface {
	InsertBalance(ctx context.Context, balance MarketBalance) error
	GetBalancesForMarket(ctx context.Context, marketID string, fromTime time.Time) ([]MarketBalance, error)
	GetBalancesForAllMarkets(ctx context.Context, fromTime time.Time) (map[string][]MarketBalance, error)
}
