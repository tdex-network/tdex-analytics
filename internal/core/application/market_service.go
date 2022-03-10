package application

import (
	"context"
	"tdex-analytics/internal/core/domain"
)

type MarketService interface {
	ListMarkets(
		ctx context.Context,
		req []MarketProvider,
		page Page,
	) ([]Market, error)
}

type marketService struct {
	marketRepository domain.MarketRepository
}

func NewMarketService(
	marketRepository domain.MarketRepository,
) MarketService {

	return &marketService{
		marketRepository: marketRepository,
	}
}

func (m marketService) ListMarkets(
	ctx context.Context,
	req []MarketProvider,
	page Page,
) ([]Market, error) {
	resp := make([]Market, 0)
	filter := make([]domain.Filter, 0)

	for _, v := range req {
		if err := v.validate(); err != nil {
			return nil, err
		}

		filter = append(filter, v.toDomain())
	}

	markets, err := m.marketRepository.GetAllMarketsForFilter(
		ctx,
		filter, page.ToDomain(),
	)
	if err != nil {
		return nil, err
	}

	for _, v := range markets {
		resp = append(resp, Market{
			ID:         v.ID,
			Url:        v.Url,
			BaseAsset:  v.BaseAsset,
			QuoteAsset: v.QuoteAsset,
		})
	}

	return resp, nil
}
