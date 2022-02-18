package application

import (
	"context"
	"errors"
	"tdex-analytics/internal/core/domain"
	"time"
)

var (
	ErrInvalidTimeFormat = errors.New("fromTime must be valid RFC3339 format")
)

type MarketBalanceService interface {
	// InsertBalance inserts market balance in current moment
	InsertBalance(
		ctx context.Context,
		marketBalance MarketBalance,
	) error
	// GetBalances returns all markets balances from time in past equal to passed arg fromTime
	//if marketID is passed method will return data for all market's, otherwise only for provided one
	GetBalances(
		ctx context.Context,
		timeRange TimeRange,
		marketIDs ...string,
	) (*MarketsBalances, error)
}

type marketBalanceService struct {
	marketBalanceRepository domain.MarketBalanceRepository
}

func NewMarketBalanceService(
	marketBalanceRepository domain.MarketBalanceRepository,
) MarketBalanceService {
	return &marketBalanceService{
		marketBalanceRepository: marketBalanceRepository,
	}
}

func (m *marketBalanceService) InsertBalance(
	ctx context.Context,
	marketBalance MarketBalance,
) error {
	if err := marketBalance.validate(); err != nil {
		return err
	}

	mbDomain, err := marketBalance.toDomain()
	if err != nil {
		return err
	}

	return m.marketBalanceRepository.InsertBalance(ctx, *mbDomain)
}

func (m *marketBalanceService) GetBalances(
	ctx context.Context,
	timeRange TimeRange,
	marketIDs ...string,
) (*MarketsBalances, error) {
	result := make(map[string][]Balance)

	startTime, endTime, err := timeRange.getStartAndEndTime(time.Now())
	if err != nil {
		return nil, err
	}

	marketsBalances, err := m.marketBalanceRepository.GetBalancesForMarkets(
		ctx,
		startTime,
		endTime,
		marketIDs...,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range marketsBalances {
		balances := make([]Balance, 0)
		for _, v1 := range v {
			balances = append(balances, Balance{
				BaseBalance:  v1.BaseBalance,
				BaseAsset:    v1.BaseAsset,
				QuoteBalance: v1.QuoteBalance,
				QuoteAsset:   v1.QuoteAsset,
				Time:         v1.Time,
			})
		}

		result[k] = balances
	}

	return &MarketsBalances{
		MarketsBalances: result,
	}, nil
}
