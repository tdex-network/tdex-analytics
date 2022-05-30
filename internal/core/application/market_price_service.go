package application

import (
	"context"
	"fmt"
	"strconv"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/core/port"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
	"time"

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

	marketsPrices, err := m.marketPriceRepository.GetPricesForMarkets(
		ctx,
		startTime,
		endTime,
		page.ToDomain(),
		marketIDs...,
	)
	if err != nil {
		return nil, err
	}

	//refPricesPerAssetPair holds prices for all assets pairs for reference currency
	//purpose is to avoid multiple queries for same asset pair
	refPricesPerAssetPair := make(map[string]referenceCurrencyPrice)
	for k, v := range marketsPrices {
		prices := make([]Price, 0)
		for _, v1 := range v {
			var basePriceInRefCurrency, quotePriceInRefCurrency decimal.Decimal
			if referenceCurrency != "" {
				b, q, err := m.getPricesInReferenceCurrency(
					ctx,
					v1,
					referenceCurrency,
					refPricesPerAssetPair,
				)
				if err != nil {
					return nil, err
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
	}, nil
}

func (m *marketPriceService) getPricesInReferenceCurrency(
	ctx context.Context,
	mktPrice domain.MarketPrice,
	referenceCurrency string,
	refPricesPerAssetPair map[string]referenceCurrencyPrice,
) (
	basePriceInRefCurrency decimal.Decimal,
	quotePriceInRefCurrency decimal.Decimal,
	err error,
) {
	assetPair := fmt.Sprintf("%s_%s", mktPrice.BaseAsset, mktPrice.QuoteAsset)
	if v, ok := refPricesPerAssetPair[assetPair]; ok {
		return v.basePrice, v.quotePrice, nil
	}

	baseAssetTicker, err := m.raterSvc.GetAssetCurrency(mktPrice.BaseAsset)
	if err != nil {
		return
	}
	quoteAssetTicker, err := m.raterSvc.GetAssetCurrency(mktPrice.QuoteAsset)
	if err != nil {
		return
	}
	isBaseAssetStable, _ := m.raterSvc.IsFiatSymbolSupported(baseAssetTicker)
	isQuoteAssetStable, _ := m.raterSvc.IsFiatSymbolSupported(quoteAssetTicker)

	defer func() {
		if !basePriceInRefCurrency.IsZero() && !quotePriceInRefCurrency.IsZero() {
			refPricesPerAssetPair[assetPair] = referenceCurrencyPrice{
				basePrice:  basePriceInRefCurrency,
				quotePrice: quotePriceInRefCurrency,
			}
		}
	}()

	switch {
	case isBaseAssetStable:
		unitOfBasePriceInRefCurrency, _ := m.raterSvc.ConvertCurrency(
			ctx,
			baseAssetTicker,
			referenceCurrency,
		)
		if !unitOfBasePriceInRefCurrency.IsZero() {
			basePriceInRefCurrency = unitOfBasePriceInRefCurrency.Mul(mktPrice.BasePrice)
			quotePriceInRefCurrency = basePriceInRefCurrency.Mul(mktPrice.QuotePrice)
		} else {
			basePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
				ctx,
				quoteAssetTicker,
				referenceCurrency,
			)
			quotePriceInRefCurrency = basePriceInRefCurrency.Mul(mktPrice.QuotePrice)
		}
	case isQuoteAssetStable:
		unitOfQuotePriceInRefCurrency, _ := m.raterSvc.ConvertCurrency(
			ctx,
			quoteAssetTicker,
			referenceCurrency,
		)
		if !unitOfQuotePriceInRefCurrency.IsZero() {
			quotePriceInRefCurrency = unitOfQuotePriceInRefCurrency.Mul(mktPrice.QuotePrice)
			basePriceInRefCurrency = quotePriceInRefCurrency.Mul(mktPrice.BasePrice)
		} else {
			quotePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
				ctx,
				baseAssetTicker,
				referenceCurrency,
			)
			basePriceInRefCurrency = quotePriceInRefCurrency.Mul(mktPrice.BasePrice)
		}
	case !isBaseAssetStable && !isQuoteAssetStable:
		basePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
			ctx,
			baseAssetTicker,
			referenceCurrency,
		)
		quotePriceInRefCurrency, _ = m.raterSvc.ConvertCurrency(
			ctx,
			quoteAssetTicker,
			referenceCurrency,
		)

		// TODO: uncomment this when we find a way to convert the prices of a
		// market with no stable coins to ref currency in case we fail to retrieve
		// one of the 2. For now, we return what we got from the rater service to
		// keep it simple.
		//
		// if basePriceInRefCurrency.IsZero() == quotePriceInRefCurrency.IsZero() {
		// 	return
		// }
	}
	basePriceInRefCurrency = basePriceInRefCurrency.Round(2)
	quotePriceInRefCurrency = quotePriceInRefCurrency.Round(2)
	return
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

	markets, err := m.marketRepository.GetAllMarkets(ctx)
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
		log.Errorf("FetchAndInsertPrice -> FetchPrice: %v", err)
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
		log.Errorf("FetchAndInsertPrice -> InsertPrice: %v", err)
		return
	}
}

type referenceCurrencyPrice struct {
	basePrice  decimal.Decimal
	quotePrice decimal.Decimal
}
