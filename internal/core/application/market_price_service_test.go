package application

import (
	"context"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/core/port"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

var (
	lbtcUsdtMarket = domain.MarketPrice{
		BasePrice:  decimal.NewFromFloat(0.000025),
		BaseAsset:  "LBTC",
		QuotePrice: decimal.NewFromFloat(40381.20),
		QuoteAsset: "USDT",
	}
	usdtLbtcMarket = domain.MarketPrice{
		BasePrice:  decimal.NewFromFloat(40381.20),
		BaseAsset:  "USDT",
		QuotePrice: decimal.NewFromFloat(0.000025),
		QuoteAsset: "LBTC",
	}
)

type args struct {
	referenceCurrency     string
	price                 domain.MarketPrice
	refPricesPerAssetPair map[string]referenceCurrencyPrice
}

type test struct {
	name                            string
	args                            args
	raterSvc                        port.RateService
	expectedErr                     error
	expectedBasePriceInRefCurrency  decimal.Decimal
	expectedQuotePriceInRefCurrency decimal.Decimal
}

func TestMarketPriceService(t *testing.T) {
	t.Run("getPricesInReferenceCurrency", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			tests := []test{
				{
					name: "lbtc_usdt to eur with base price conversion",
					args: args{
						referenceCurrency:     "EUR",
						price:                 lbtcUsdtMarket,
						refPricesPerAssetPair: make(map[string]referenceCurrencyPrice),
					},
					raterSvc:                        mockRater("LBTC", "USDT", "LBTC", "USDT", decimal.NewFromFloat(37384.11), decimal.Zero, false, true, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(0.93),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(37384.11),
				},
				{
					name: "lbtc_usdt to eur with quote price conversion",
					args: args{
						referenceCurrency:     "EUR",
						price:                 lbtcUsdtMarket,
						refPricesPerAssetPair: make(map[string]referenceCurrencyPrice),
					},
					raterSvc:                        mockRater("LBTC", "USDT", "LBTC", "USDT", decimal.Zero, decimal.NewFromFloat(0.93), false, true, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(0.93),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(37384.11),
				},
				{
					name: "usdt_lbtc to eur with base price conversion",
					args: args{
						referenceCurrency:     "EUR",
						price:                 usdtLbtcMarket,
						refPricesPerAssetPair: make(map[string]referenceCurrencyPrice),
					},
					raterSvc:                        mockRater("USDT", "LBTC", "USDT", "LBTC", decimal.NewFromFloat(0.93), decimal.Zero, true, false, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(37384.11),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(0.93),
				},
				{
					name: "usdt_lbtc to eur with quote price conversion",
					args: args{
						referenceCurrency:     "EUR",
						price:                 usdtLbtcMarket,
						refPricesPerAssetPair: make(map[string]referenceCurrencyPrice),
					},
					raterSvc:                        mockRater("USDT", "LBTC", "USDT", "LBTC", decimal.Zero, decimal.NewFromFloat(37384.11), true, false, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(37384.11),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(0.93),
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					m := &marketPriceService{raterSvc: tt.raterSvc}
					basePriceInRefCurrency, quotePriceInRefCurrency, err := m.getPricesInReferenceCurrency(
						context.Background(),
						tt.args.price,
						tt.args.referenceCurrency,
						tt.args.refPricesPerAssetPair,
					)
					if err != nil {
						t.Fatal(err)
					}
					if basePriceInRefCurrency.LessThan(tt.expectedBasePriceInRefCurrency) {
						t.Errorf("got basePriceInRefCurrency = %s, less than expected %s", basePriceInRefCurrency, tt.expectedBasePriceInRefCurrency)
					}
					if quotePriceInRefCurrency.LessThan(tt.expectedQuotePriceInRefCurrency) {
						t.Errorf("got quotePriceInRefCurrency = %s, less than expected %s", quotePriceInRefCurrency, tt.expectedQuotePriceInRefCurrency)
					}
				})
			}
		})

		t.Run("invalid", func(t *testing.T) {
			tests := []test{}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					m := &marketPriceService{raterSvc: tt.raterSvc}
					basePriceInRefCurrency, quotePriceInRefCurrency, err := m.getPricesInReferenceCurrency(
						context.Background(),
						tt.args.price,
						tt.args.referenceCurrency,
						tt.args.refPricesPerAssetPair,
					)
					if err == nil || err.Error() != tt.expectedErr.Error() {
						t.Errorf("got error = %v, wantErr %v", err, tt.expectedErr)
						return
					}
					if !basePriceInRefCurrency.IsZero() {
						t.Errorf("got basePriceInRefCurrency = %s, expected 0", basePriceInRefCurrency)
					}
					if !quotePriceInRefCurrency.IsZero() {
						t.Errorf("got quotePriceInRefCurrency = %s, expected 0", quotePriceInRefCurrency)
					}
				})
			}
		})
	})
}

func mockRater(
	baseAssetID, quoteAssetID, baseCurrency, quoteCurrency string,
	baseResult, quoteResult decimal.Decimal,
	isBaseAssetStable, isQuoteAssetStable bool,
	err error,
) port.RateService {
	raterMock := new(port.MockRateService)
	raterMock.On("ConvertCurrency", mock.Anything, baseCurrency, mock.Anything).Return(baseResult, err)
	raterMock.On("ConvertCurrency", mock.Anything, quoteCurrency, mock.Anything).Return(quoteResult, err)
	raterMock.On("GetAssetCurrency", baseAssetID).Return(baseCurrency, err)
	raterMock.On("GetAssetCurrency", quoteAssetID).Return(quoteCurrency, err)
	raterMock.On("IsFiatSymbolSupported", baseAssetID).Return(isBaseAssetStable, nil)
	raterMock.On("IsFiatSymbolSupported", quoteAssetID).Return(isQuoteAssetStable, nil)

	return raterMock
}
