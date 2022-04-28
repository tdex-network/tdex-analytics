package application

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/core/port"
	"testing"
)

func TestMarketPriceServiceGetReferencePrices(t *testing.T) {
	type args struct {
		ctx                   context.Context
		referenceCurrency     string
		price                 domain.MarketPrice
		refPricesPerAssetPair map[string]struct {
			basePriceInRefCurrency  decimal.Decimal
			quotePriceInRefCurrency decimal.Decimal
		}
	}
	tests := []struct {
		name                string
		args                args
		baseCurrency        string
		baseConversionRate  decimal.Decimal
		quoteCurrency       string
		quoteConversionRate decimal.Decimal
		baseRateErr         error
		quoteRateErr        error
		want                decimal.Decimal
		want1               decimal.Decimal
		wantErr             bool
	}{
		{
			name: "1",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.5),
					BaseAsset:  "baseAsset",
					QuotePrice: decimal.NewFromFloat(2),
					QuoteAsset: "quoteAsset",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:       "LBTC",
			baseConversionRate: decimal.NewFromFloat(1),
			quoteCurrency:      "USDT",
			want:               decimal.NewFromFloat(1),
			want1:              decimal.NewFromFloat(1),
			wantErr:            false,
		},
		{
			name: "2",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.5),
					BaseAsset:  "baseAsset",
					QuotePrice: decimal.NewFromFloat(2),
					QuoteAsset: "quoteAsset",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:        "LBTC",
			baseConversionRate:  decimal.NewFromFloat(1),
			baseRateErr:         errors.New("error"),
			quoteCurrency:       "USDT",
			quoteConversionRate: decimal.NewFromFloat(2),
			want:                decimal.NewFromFloat(0),
			want1:               decimal.NewFromFloat(0),
			wantErr:             true,
		},
		{
			name: "3",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.5),
					BaseAsset:  "baseAsset",
					QuotePrice: decimal.NewFromFloat(2),
					QuoteAsset: "quoteAsset",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:        "LBTC",
			baseRateErr:         port.ErrCurrencyNotFound,
			quoteCurrency:       "USDT",
			quoteConversionRate: decimal.NewFromFloat(0.5),
			want:                decimal.NewFromFloat(1),
			want1:               decimal.NewFromFloat(1),
			wantErr:             false,
		},
		// 1 lbtc = 37384,11 eur
		// 1 eur = 0,000027 lbtc
		// 1 usdt = 0,000025 lbtc
		// 1 lbtc = 40381,20 usdt
		// 1 lcad = 0,000020 lbtc
		// 1 lbtc = 51151,07 cad
		// 1 usdt = 0.93 eur
		// 1 cad = 0.73 eur
		{
			name: "lbtc/usdt to eur, with base conversion",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.000025),
					BaseAsset:  "LBTC",
					QuotePrice: decimal.NewFromFloat(40381.20),
					QuoteAsset: "USDT",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:       "LBTC",
			baseConversionRate: decimal.NewFromFloat(37384.11),
			quoteCurrency:      "USDT",
			want:               decimal.NewFromFloat(0.00002675),
			want1:              decimal.NewFromFloat(37384.11),
			wantErr:            false,
		},
		{
			name: "lbtc/lcad to eur, with base conversion",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.00002),
					BaseAsset:  "LBTC",
					QuotePrice: decimal.NewFromFloat(51151.07),
					QuoteAsset: "LCAD",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:       "LBTC",
			quoteCurrency:      "LCAD",
			baseConversionRate: decimal.NewFromFloat(37384.11),
			want:               decimal.NewFromFloat(0.00002675),
			want1:              decimal.NewFromFloat(37384.11),
			wantErr:            false,
		},
		{
			name: "not able to convert both currencies",
			args: args{
				ctx:               nil,
				referenceCurrency: "EUR",
				price: domain.MarketPrice{
					BasePrice:  decimal.NewFromFloat(0.00002),
					BaseAsset:  "LBTC",
					QuotePrice: decimal.NewFromFloat(51151.07),
					QuoteAsset: "LCAD",
				},
				refPricesPerAssetPair: refPricesPerAssetPair(),
			},
			baseCurrency:  "LBTC",
			baseRateErr:   port.ErrCurrencyNotFound,
			quoteCurrency: "LCAD",
			quoteRateErr:  port.ErrCurrencyNotFound,
			want:          decimal.NewFromFloat(0),
			want1:         decimal.NewFromFloat(0),
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &marketPriceService{
				raterSvc: mockRater(
					tt.args.price.BaseAsset,
					tt.args.price.QuoteAsset,
					tt.baseCurrency,
					tt.quoteCurrency,
					tt.args.referenceCurrency,
					tt.baseConversionRate,
					tt.quoteConversionRate,
					tt.baseRateErr,
					tt.quoteRateErr,
				),
			}
			got, got1, err := m.getReferencePrices(tt.args.ctx, tt.args.referenceCurrency, tt.args.price.BaseAsset, tt.args.price.QuoteAsset, tt.args.refPricesPerAssetPair, tt.args.price.QuotePrice)
			if (err != nil) != tt.wantErr {
				t.Errorf("getReferencePrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("getReferencePrices() got = %v, want %v", got, tt.want)
			}
			if !got1.Equal(tt.want1) {
				t.Errorf("getReferencePrices() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func mockRater(baseAssetID, quoteAssetID, baseCurrency, quoteCurrency, refCurrency string, baseResult, quoteResult decimal.Decimal, baseErr, quoteErr error) port.RateService {
	raterMock := new(port.MockRateService)
	raterMock.On("ConvertCurrency", mock.Anything, baseCurrency, refCurrency).Return(baseResult, baseErr)
	raterMock.On("ConvertCurrency", mock.Anything, quoteCurrency, refCurrency).Return(quoteResult, quoteErr)
	raterMock.On("GetAssetCurrency", baseAssetID).Return(baseCurrency, nil)
	raterMock.On("GetAssetCurrency", quoteAssetID).Return(quoteCurrency, nil)
	raterMock.On("IsFiatSymbolSupported", mock.Anything).Return(true, nil)

	return raterMock
}

func refPricesPerAssetPair() map[string]struct {
	basePriceInRefCurrency  decimal.Decimal
	quotePriceInRefCurrency decimal.Decimal
} {
	return make(map[string]struct {
		basePriceInRefCurrency  decimal.Decimal
		quotePriceInRefCurrency decimal.Decimal
	})
}
