package application

import (
	"context"
	"tdex-analytics/internal/core/domain"
)

type MarketService interface {
	ListMarketIDs(ctx context.Context, req []MarketRequest) ([]int64, error)
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

func (m marketService) ListMarketIDs(
	ctx context.Context,
	req []MarketRequest,
) ([]int64, error) {
	resp := make([]int64, 0)
	filter := make([]domain.Filter, 0)

	for _, v := range req {
		if err := v.validate(); err != nil {
			return nil, err
		}

		filter = append(filter, v.toDomain())
	}

	markets, err := m.marketRepository.GetAllMarketsForFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, v := range markets {
		resp = append(resp, int64(v.ID))
	}

	return resp, nil
}
