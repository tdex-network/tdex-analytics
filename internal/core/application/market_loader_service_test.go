package application

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"github.com/tdex-network/tdex-analytics/internal/infrastructure/db/inmemory"
	"testing"
)

var (
	ctx     context.Context
	filter1 = domain.Filter{
		Url:        "url1",
		BaseAsset:  "ba1",
		QuoteAsset: "qa1",
	}
	filter2 = domain.Filter{
		Url:        "url2",
		BaseAsset:  "ba2",
		QuoteAsset: "qa2",
	}
	filter3 = domain.Filter{
		Url:        "url3",
		BaseAsset:  "ba3",
		QuoteAsset: "qa3",
	}
	page = domain.Page{
		Number: 1,
		Size:   20,
	}
)

func TestUpdateMarketActiveStatusAndInsertNew(t *testing.T) {
	type fields struct {
		marketRepository domain.MarketRepository
	}
	type args struct {
		activeMarkets []domain.Market
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		prepareData func(repo domain.MarketRepository) ([]domain.Market, error)
		validate    func(t *testing.T, repo domain.MarketRepository)
	}{
		{
			name: "activate_existing_market",
			fields: fields{
				marketRepository: inmemory.NewRepository(),
			},
			args: args{
				activeMarkets: []domain.Market{
					{
						ProviderName: "pn2",
						Url:          "url2",
						BaseAsset:    "ba2",
						QuoteAsset:   "qa2",
						Active:       true,
					},
				},
			},
			wantErr: false,
			prepareData: func(repo domain.MarketRepository) ([]domain.Market, error) {
				for _, v := range prepareMarkets1() {
					if err := repo.InsertMarket(ctx, v); err != nil {
						return nil, err
					}
				}

				return prepareMarkets1(), nil
			},
			validate: func(t *testing.T, repo domain.MarketRepository) {
				m1, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter1}, page)
				require.NoError(t, err)
				require.Equal(t, false, m1[0].Active)

				m2, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter2}, page)
				require.NoError(t, err)
				require.Equal(t, true, m2[0].Active)

				m3, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter3}, page)
				require.NoError(t, err)
				require.Equal(t, false, m3[0].Active)
			},
		},
		{
			name: "deactivate_existing_markets",
			fields: fields{
				marketRepository: inmemory.NewRepository(),
			},
			args: args{
				activeMarkets: []domain.Market{},
			},
			wantErr: false,
			prepareData: func(repo domain.MarketRepository) ([]domain.Market, error) {
				for _, v := range prepareMarkets1() {
					if err := repo.InsertMarket(ctx, v); err != nil {
						return nil, err
					}
				}

				return prepareMarkets1(), nil
			},
			validate: func(t *testing.T, repo domain.MarketRepository) {
				m1, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter1}, page)
				require.NoError(t, err)
				require.Equal(t, false, m1[0].Active)

				m2, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter2}, page)
				require.NoError(t, err)
				require.Equal(t, false, m2[0].Active)

				m3, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter3}, page)
				require.NoError(t, err)
				require.Equal(t, false, m3[0].Active)
			},
		},
		{
			name: "activate_existing_markets_and_insert_new_ones",
			fields: fields{
				marketRepository: inmemory.NewRepository(),
			},
			args: args{
				activeMarkets: []domain.Market{
					{
						ProviderName: "pn1",
						Url:          "url1",
						BaseAsset:    "ba1",
						QuoteAsset:   "qa1",
						Active:       true,
					},
					{
						ProviderName: "pn2",
						Url:          "url2",
						BaseAsset:    "ba2",
						QuoteAsset:   "qa2",
						Active:       true,
					},
				},
			},
			wantErr: false,
			prepareData: func(repo domain.MarketRepository) ([]domain.Market, error) {
				for _, v := range prepareMarkets2() {
					if err := repo.InsertMarket(ctx, v); err != nil {
						return nil, err
					}
				}

				return prepareMarkets2(), nil
			},
			validate: func(t *testing.T, repo domain.MarketRepository) {
				m2, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter1}, page)
				require.NoError(t, err)
				require.Equal(t, true, m2[0].Active)

				m3, err := repo.GetAllMarketsForFilter(ctx, []domain.Filter{filter2}, page)
				require.NoError(t, err)
				require.Equal(t, true, m3[0].Active)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingMarkets, err := tt.prepareData(tt.fields.marketRepository)
			if err != nil {
				t.Errorf("prepare func failed: %v", err)
			}

			m := &marketsLoaderService{
				marketRepository: tt.fields.marketRepository,
			}

			if err := m.updateMarketActiveStatusAndInsertNew(existingMarkets, tt.args.activeMarkets); (err != nil) != tt.wantErr {
				t.Errorf("updateMarketActiveStatusAndInsertNew() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, tt.fields.marketRepository)
		})
	}
}

func prepareMarkets1() []domain.Market {
	return []domain.Market{
		{
			ID:           1,
			ProviderName: "pn1",
			Url:          "url1",
			BaseAsset:    "ba1",
			QuoteAsset:   "qa1",
			Active:       true,
		},
		{
			ID:           2,
			ProviderName: "pn2",
			Url:          "url2",
			BaseAsset:    "ba2",
			QuoteAsset:   "qa2",
			Active:       true,
		},
		{
			ID:           3,
			ProviderName: "pn3",
			Url:          "url3",
			BaseAsset:    "ba3",
			QuoteAsset:   "qa3",
			Active:       true,
		},
	}
}

func prepareMarkets2() []domain.Market {
	return []domain.Market{
		{
			ID:           1,
			ProviderName: "pn1",
			Url:          "url1",
			BaseAsset:    "ba1",
			QuoteAsset:   "qa1",
			Active:       false,
		},
	}
}
