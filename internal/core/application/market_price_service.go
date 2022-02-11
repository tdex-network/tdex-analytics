package application

import (
	"context"
	"strconv"
	"tdex-analytics/internal/core/domain"
	"time"
)

type MarketPriceService interface {
	// InsertPrice inserts market price in current moment
	InsertPrice(
		ctx context.Context,
		marketPrice MarketPrice,
	) error
	// GetPrices returns all markets prices from time in past equal to passed arg fromTime
	//if marketID is passed method will return data for all market's, otherwise only for provided one
	GetPrices(
		ctx context.Context,
		marketID int,
		fromTime string,
	) (*MarketsPrices, error)
}

type marketPriceService struct {
	marketPriceRepository domain.MarketPriceRepository
}

func NewMarketPriceService(
	marketPriceRepository domain.MarketPriceRepository,
) MarketPriceService {
	return &marketPriceService{
		marketPriceRepository: marketPriceRepository,
	}
}

func (m *marketPriceService) InsertPrice(
	ctx context.Context,
	marketPrice MarketPrice,
) error {
	if err := marketPrice.validate(); err != nil {
		return err
	}

	mbDomain, err := marketPrice.toDomain()
	if err != nil {
		return err
	}

	return m.marketPriceRepository.InsertPrice(ctx, *mbDomain)
}

func (m *marketPriceService) GetPrices(
	ctx context.Context,
	marketID int,
	fromTime string,
) (*MarketsPrices, error) {
	result := make(map[int][]Price)

	if err := validateTimeFormat(fromTime); err != nil {
		return nil, err
	}

	tm, _ := time.Parse(time.RFC3339, fromTime)

	if marketID > 0 {
		marketPrices, err := m.marketPriceRepository.GetPricesForMarket(ctx, strconv.Itoa(marketID), tm)
		if err != nil {
			return nil, err
		}
		prices := make([]Price, 0)
		for _, v := range marketPrices {
			prices = append(prices, Price{
				BasePrice:  v.BasePrice,
				BaseAsset:  v.BaseAsset,
				QuotePrice: v.QuotePrice,
				QuoteAsset: v.QuoteAsset,
				Time:       v.Time,
			})

			marketID, err := strconv.Atoi(v.MarketID)
			if err != nil {
				return nil, err
			}

			result[marketID] = prices
		}
	} else {
		marketsPrices, err := m.marketPriceRepository.GetPricesForAllMarkets(ctx, tm)
		if err != nil {
			return nil, err
		}

		for k, v := range marketsPrices {
			prices := make([]Price, 0)
			for _, v1 := range v {
				prices = append(prices, Price{
					BasePrice:  v1.BasePrice,
					BaseAsset:  v1.BaseAsset,
					QuotePrice: v1.QuotePrice,
					QuoteAsset: v1.QuoteAsset,
					Time:       v1.Time,
				})
			}

			marketID, err := strconv.Atoi(k)
			if err != nil {
				return nil, err
			}

			result[marketID] = prices
		}
	}

	return &MarketsPrices{
		MarketsPrices: result,
	}, nil
}
