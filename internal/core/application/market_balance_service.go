package application

import (
	"context"
	"errors"
	"strconv"
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
		marketID int,
		fromTime string,
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
	marketID int,
	fromTime string,
) (*MarketsBalances, error) {
	result := make(map[int][]Balance)

	if err := validateTimeFormat(fromTime); err != nil {
		return nil, err
	}

	tm, _ := time.Parse(time.RFC3339, fromTime)

	if marketID > 0 {
		marketBalances, err := m.marketBalanceRepository.GetBalancesForMarket(ctx, strconv.Itoa(marketID), tm)
		if err != nil {
			return nil, err
		}
		balances := make([]Balance, 0)
		for _, v := range marketBalances {
			balances = append(balances, Balance{
				BaseBalance:  v.BaseBalance,
				BaseAsset:    v.BaseAsset,
				QuoteBalance: v.QuoteBalance,
				QuoteAsset:   v.QuoteAsset,
				Time:         v.Time,
			})

			marketID, err := strconv.Atoi(v.MarketID)
			if err != nil {
				return nil, err
			}

			result[marketID] = balances
		}
	} else {
		marketsBalances, err := m.marketBalanceRepository.GetBalancesForAllMarkets(ctx, tm)
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
			marketID, err := strconv.Atoi(k)
			if err != nil {
				return nil, err
			}
			result[marketID] = balances
		}
	}

	return &MarketsBalances{
		MarketsBalances: result,
	}, nil
}
