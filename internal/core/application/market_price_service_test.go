package application

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"github.com/tdex-network/tdex-analytics/internal/core/port"

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
	referenceCurrency string
	price             domain.MarketPrice
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
						referenceCurrency: "EUR",
						price:             lbtcUsdtMarket,
					},
					raterSvc:                        mockRater("LBTC", "USDT", "LBTC", "USDT", decimal.NewFromFloat(37384.11), decimal.Zero, false, true, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(0.93),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(37384.11),
				},
				{
					name: "lbtc_usdt to eur with quote price conversion",
					args: args{
						referenceCurrency: "EUR",
						price:             lbtcUsdtMarket,
					},
					raterSvc:                        mockRater("LBTC", "USDT", "LBTC", "USDT", decimal.Zero, decimal.NewFromFloat(0.93), false, true, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(0.93),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(37384.11),
				},
				{
					name: "usdt_lbtc to eur with base price conversion",
					args: args{
						referenceCurrency: "EUR",
						price:             usdtLbtcMarket,
					},
					raterSvc:                        mockRater("USDT", "LBTC", "USDT", "LBTC", decimal.NewFromFloat(0.93), decimal.Zero, true, false, nil),
					expectedBasePriceInRefCurrency:  decimal.NewFromFloat(37384.11),
					expectedQuotePriceInRefCurrency: decimal.NewFromFloat(0.93),
				},
				{
					name: "usdt_lbtc to eur with quote price conversion",
					args: args{
						referenceCurrency: "EUR",
						price:             usdtLbtcMarket,
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

func TestGroupMarkets(t *testing.T) {
	markets := []domain.Market{
		{ID: 1, BaseAsset: "BTC", QuoteAsset: "USD"},
		{ID: 2, BaseAsset: "ETH", QuoteAsset: "USD"},
		{ID: 3, BaseAsset: "BTC", QuoteAsset: "USD"},
	}

	testCases := []struct {
		name                             string
		ids                              []string
		expectedMarketsMap               map[int]domain.Market
		expectedMarketsWithSameAssetPair map[string][]string
	}{
		{
			name: "Test with ids 1, 2, 3",
			ids:  []string{"1", "2", "3"},
			expectedMarketsMap: map[int]domain.Market{
				1: {ID: 1, BaseAsset: "BTC", QuoteAsset: "USD"},
				2: {ID: 2, BaseAsset: "ETH", QuoteAsset: "USD"},
				3: {ID: 3, BaseAsset: "BTC", QuoteAsset: "USD"},
			},
			expectedMarketsWithSameAssetPair: map[string][]string{
				"BTCUSD": {"1", "3"},
				"ETHUSD": {"2"},
			},
		},
		{
			name: "Test with id 1",
			ids:  []string{"1"},
			expectedMarketsMap: map[int]domain.Market{
				1: {ID: 1, BaseAsset: "BTC", QuoteAsset: "USD"},
				2: {ID: 2, BaseAsset: "ETH", QuoteAsset: "USD"},
				3: {ID: 3, BaseAsset: "BTC", QuoteAsset: "USD"},
			},
			expectedMarketsWithSameAssetPair: map[string][]string{
				"BTCUSD": {"1"},
			},
		},
		{
			name: "Test with ids 1, 2",
			ids:  []string{"1", "2"},
			expectedMarketsMap: map[int]domain.Market{
				1: {ID: 1, BaseAsset: "BTC", QuoteAsset: "USD"},
				2: {ID: 2, BaseAsset: "ETH", QuoteAsset: "USD"},
				3: {ID: 3, BaseAsset: "BTC", QuoteAsset: "USD"},
			},
			expectedMarketsWithSameAssetPair: map[string][]string{
				"BTCUSD": {"1"},
				"ETHUSD": {"2"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			marketsMap, marketsWithSameAssetPair, err := groupMarkets(markets, testCase.ids)
			require.NoError(t, err)

			if !reflect.DeepEqual(marketsMap, testCase.expectedMarketsMap) {
				t.Errorf("Expected markets map %v, but got %v", testCase.expectedMarketsMap, marketsMap)
			}

			if !reflect.DeepEqual(marketsWithSameAssetPair, testCase.expectedMarketsWithSameAssetPair) {
				t.Errorf(
					"Expected markets with same asset pair %v, but got %v",
					testCase.expectedMarketsWithSameAssetPair,
					marketsWithSameAssetPair,
				)
			}
		})
	}
}
