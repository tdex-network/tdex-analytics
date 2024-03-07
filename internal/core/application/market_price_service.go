package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"github.com/tdex-network/tdex-analytics/internal/core/port"
	tdexmarketloader "github.com/tdex-network/tdex-analytics/pkg/tdex-market-loader"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
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
		timeRange TimeRange,
		page Page,
		referenceCurrency string,
		timeFrame TimeFrame,
		marketIDs ...string,
	) (*MarketsPrices, error)
	// StartFetchingPricesJob starts cron job that will periodically fetch and store prices for all markets
	StartFetchingPricesJob() error
}

type marketPriceService struct {
	marketPriceRepository    domain.MarketPriceRepository
	marketRepository         domain.MarketRepository
	tdexMarketLoaderSvc      tdexmarketloader.Service
	cronSvc                  *cron.Cron
	fetchPriceCronExpression string
	raterSvc                 port.RateService
}

func NewMarketPriceService(
	marketPriceRepository domain.MarketPriceRepository,
	marketRepository domain.MarketRepository,
	tdexMarketLoaderSvc tdexmarketloader.Service,
	jobPeriodInMinutes string,
	raterSvc port.RateService,
) MarketPriceService {
	return &marketPriceService{
		marketPriceRepository:    marketPriceRepository,
		cronSvc:                  cron.New(),
		marketRepository:         marketRepository,
		tdexMarketLoaderSvc:      tdexMarketLoaderSvc,
		fetchPriceCronExpression: fmt.Sprintf("@every %vm", jobPeriodInMinutes),
		raterSvc:                 raterSvc,
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
	timeRange TimeRange,
	page Page,
	referenceCurrency string,
	timeFrame TimeFrame,
	marketIDs ...string,
) (*MarketsPrices, error) {
	if referenceCurrency != "" {
		supportedFiat, err := m.raterSvc.IsFiatSymbolSupported(referenceCurrency)
		if err != nil {
			return nil, err
		}
		if !supportedFiat {
			return nil, fmt.Errorf("reference currency %s is not supported", referenceCurrency)
		}
	}

	result := make(map[string][]Price)
	startTime, endTime, err := timeRange.getStartAndEndTime(time.Now())
	if err != nil {
		return nil, err
	}

	if err := timeFrame.validate(); err != nil {
		return nil, err
	}

	if int(endTime.Sub(startTime).Minutes()) <= timeFrame.toMinutes() {
		return nil, ErrInvalidTimeFrame
	}

	markets, err := m.marketRepository.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsMap, marketsWithSameAssetPair, err := groupMarkets(markets, marketIDs)
	if err != nil {
		return nil, err
	}

	averagePricesInfos := make([]AveragePriceInfo, 0)
	if len(marketIDs) > 0 {
		averageWindow := getAverageWindow(startTime, endTime)
		for _, v := range marketsWithSameAssetPair {
			vwamp, err := m.marketPriceRepository.CalculateVWAP(
				ctx, averageWindow, startTime, endTime, v...)
			if err != nil {
				return nil, err
			}

			var averageReferentPrice decimal.Decimal
			if referenceCurrency != "" {
				mktId, err := strconv.Atoi(v[0])
				if err != nil {
					return nil, err
				}
				quoteAssetTicker, err := m.raterSvc.GetAssetCurrency(
					marketsMap[mktId].QuoteAsset,
				)
				if err == nil {
					unitOfQuotePriceInRefCurrency, err := m.raterSvc.ConvertCurrency(
						ctx,
						quoteAssetTicker,
						referenceCurrency,
					)
					if err == nil {
						averageReferentPrice = vwamp.Mul(unitOfQuotePriceInRefCurrency)
					}
				}
			}

			averagePricesInfos = append(averagePricesInfos, AveragePriceInfo{
				MarketIDs:            v,
				AveragePrice:         vwamp,
				AverageReferentPrice: averageReferentPrice,
			})
		}
	}

	marketsPrices, err := m.marketPriceRepository.GetPricesForMarkets(
		ctx,
		startTime,
		endTime,
		page.ToDomain(),
		timeFrame.toFluxDuration(),
		marketIDs...,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range marketsPrices {
		marketIdInt, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		baseAsset := marketsMap[marketIdInt].BaseAsset
		quoteAsset := marketsMap[marketIdInt].QuoteAsset

		prices := make([]Price, 0)
		for _, v1 := range v {
			v1.BaseAsset = baseAsset
			v1.QuoteAsset = quoteAsset

			var basePriceInRefCurrency, quotePriceInRefCurrency decimal.Decimal
			if referenceCurrency != "" {
				b, q, err := m.getPricesInReferenceCurrency(
					ctx,
					v1,
					referenceCurrency,
				)
				if err != nil {
					log.Debugf("GetPrices -> getPricesInReferenceCurrency: %v", err)
				}

				basePriceInRefCurrency = b
				quotePriceInRefCurrency = q
			}

			prices = append(prices, Price{
				BasePrice:          v1.BasePrice,
				BaseAsset:          v1.BaseAsset,
				BaseReferentPrice:  basePriceInRefCurrency,
				QuotePrice:         v1.QuotePrice,
				QuoteAsset:         v1.QuoteAsset,
				QuoteReferentPrice: quotePriceInRefCurrency,
				Time:               v1.Time,
			})
		}

		result[k] = prices
	}

	return &MarketsPrices{
		MarketsPrices: result,
		AveragePrices: averagePricesInfos,
	}, nil
}

func getAverageWindow(startTime, endTime time.Time) string {
	rangeDuration := endTime.Sub(startTime)

	if rangeDuration <= 3*time.Hour {
		return "1m"
	} else if rangeDuration <= 24*time.Hour {
		return "1h"
	} else if rangeDuration <= 7*24*time.Hour {
		return "6h"
	} else if rangeDuration <= 30*24*time.Hour {
		return "12h"
	} else if rangeDuration <= 365*24*time.Hour {
		return "1d"
	} else {
		return "15d"
	}
}

func groupMarkets(
	markets []domain.Market, marketIdsForVwap []string,
) (map[int]domain.Market, map[string][]string, error) {
	marketsMap := make(map[int]domain.Market)
	marketsWithSameAssetPair := make(map[string][]string)
	marketIdsForVwapMap := make(map[int]bool)
	for _, v := range marketIdsForVwap {
		mktId, err := strconv.Atoi(v)
		if err != nil {
			return nil, nil, err
		}
		marketIdsForVwapMap[mktId] = true
	}

	for _, v := range markets {
		marketsMap[v.ID] = v

		if marketIdsForVwapMap[v.ID] {
			if _, ok := marketsWithSameAssetPair[v.BaseAsset+v.QuoteAsset]; !ok {
				marketsWithSameAssetPair[v.BaseAsset+v.QuoteAsset] = make([]string, 0)
			}
			marketsWithSameAssetPair[v.BaseAsset+v.QuoteAsset] =
				append(marketsWithSameAssetPair[v.BaseAsset+v.QuoteAsset], strconv.Itoa(v.ID))
		}
	}
	return marketsMap, marketsWithSameAssetPair, nil
}

func (m *marketPriceService) getPricesInReferenceCurrency(
	ctx context.Context,
	mktPrice domain.MarketPrice,
	referenceCurrency string,
) (decimal.Decimal, decimal.Decimal, error) {
	baseAssetTickerFound, quoteAssetTickerFound := false, false
	baseAssetTicker, err := m.raterSvc.GetAssetCurrency(mktPrice.BaseAsset)
	if err == nil {
		baseAssetTickerFound = true
	}
	quoteAssetTicker, err := m.raterSvc.GetAssetCurrency(mktPrice.QuoteAsset)
	if err == nil {
		quoteAssetTickerFound = true
	}

	var basePriceInRefCurrency, quotePriceInRefCurrency decimal.Decimal

	if baseAssetTickerFound {
		quotePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
			ctx,
			baseAssetTicker,
			referenceCurrency,
		)
	}
	if quoteAssetTickerFound {
		basePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
			ctx,
			quoteAssetTicker,
			referenceCurrency,
		)
	}

	if basePriceInRefCurrency.IsZero() == quotePriceInRefCurrency.IsZero() {
		basePriceInRefCurrency = basePriceInRefCurrency.Round(2)
		quotePriceInRefCurrency = quotePriceInRefCurrency.Round(2)
		return basePriceInRefCurrency, quotePriceInRefCurrency, nil
	}

	if !basePriceInRefCurrency.IsZero() {
		quotePriceInRefCurrency = basePriceInRefCurrency.Mul(mktPrice.QuotePrice)
	} else if !quotePriceInRefCurrency.IsZero() {
		basePriceInRefCurrency = quotePriceInRefCurrency.Mul(mktPrice.BasePrice)
	}

	basePriceInRefCurrency = basePriceInRefCurrency.Round(2)
	quotePriceInRefCurrency = quotePriceInRefCurrency.Round(2)

	return basePriceInRefCurrency, quotePriceInRefCurrency, nil
}

func (m *marketPriceService) StartFetchingPricesJob() error {
	if _, err := m.cronSvc.AddJob(
		m.fetchPriceCronExpression,
		cron.FuncJob(m.FetchPricesForAllMarkets),
	); err != nil {
		return err
	}

	m.cronSvc.Start()

	return nil
}

func (m *marketPriceService) FetchPricesForAllMarkets() {
	log.Infof("job FetchPricesForAllMarkets at: %v", time.Now())
	ctx := context.Background()

	markets, err := m.marketRepository.GetMarketsForActiveIndicator(ctx, true)
	if err != nil {
		log.Errorf("FetchPricesForAllMarkets -> GetAllMarkets: %v", err)
		return
	}

	for _, v := range markets {
		go func(market domain.Market) {
			m.FetchAndInsertPrice(ctx, market)
		}(v)
	}
}

func (m *marketPriceService) FetchAndInsertPrice(
	ctx context.Context,
	market domain.Market,
) {
	price, err := m.tdexMarketLoaderSvc.FetchPrice(
		ctx,
		tdexmarketloader.Market{
			Url:        market.Url,
			QuoteAsset: market.QuoteAsset,
			BaseAsset:  market.BaseAsset,
		},
	)
	if err != nil {
		log.Errorf("FetchAndInsertPrice for %s -> FetchPrice: %v", market.Url, err)
		return
	}

	if err := m.InsertPrice(ctx, MarketPrice{
		MarketID:   strconv.Itoa(market.ID),
		BasePrice:  price.BasePrice,
		BaseAsset:  market.BaseAsset,
		QuotePrice: price.QuotePrice,
		QuoteAsset: market.QuoteAsset,
		Time:       time.Now(),
	}); err != nil {
		log.Errorf("FetchAndInsertPrice for %s -> InsertPrice: %v", market.Url, err)
		return
	}
}

type referenceCurrencyPrice struct {
	basePrice  decimal.Decimal
	quotePrice decimal.Decimal
}
