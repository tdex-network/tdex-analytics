package dbpg

import (
	"context"
	"github.com/lib/pq"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/infrastructure/db/pg/sqlc/queries"
)

const (
	uniqueViolation = "23505"
)

func (p *postgresDbService) InsertMarket(
	ctx context.Context,
	market domain.Market,
) error {
	if _, err := p.querier.InsertMarket(ctx, queries.InsertMarketParams{
		ProviderName: market.ProviderName,
		Url:          market.Url,
		BaseAsset:    market.BaseAsset,
		QuoteAsset:   market.QuoteAsset,
	}); err != nil {
		if pqErr := err.(*pq.Error); pqErr != nil {
			if pqErr.Code == uniqueViolation {
				return nil
			}
		}
	}

	return nil
}

func (p *postgresDbService) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	markets, err := p.querier.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Market, 0, len(markets))
	for _, v := range markets {
		res = append(res, domain.Market{
			ID:           int(v.MarketID.Int32),
			ProviderName: v.ProviderName,
			Url:          v.Url,
			BaseAsset:    v.BaseAsset,
			QuoteAsset:   v.QuoteAsset,
		})
	}

	return res, nil
}
