package dbpg

import (
	"context"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/infrastructure/db/pg/sqlc/queries"
)

func (p *postgresDbService) InsertMarket(
	ctx context.Context,
	market domain.Market,
) error {
	if _, err := p.querier.InsertMarket(ctx, queries.InsertMarketParams{
		AccountIndex: int32(market.AccountIndex),
		ProviderName: market.ProviderName,
		Url:          market.Url,
		Credentials:  market.Credentials,
		BaseAsset:    market.BaseAsset,
		QuoteAsset:   market.QuoteAsset,
	}); err != nil {
		return err
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
			ID:           int(v.MarketID),
			AccountIndex: int(v.AccountIndex),
			ProviderName: v.ProviderName,
			Url:          v.Url,
			Credentials:  v.Credentials,
			BaseAsset:    v.BaseAsset,
			QuoteAsset:   v.QuoteAsset,
		})
	}

	return res, nil
}
