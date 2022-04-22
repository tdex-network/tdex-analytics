package application

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"strconv"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/core/port"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
	"time"
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
	explorerSvc              port.ExplorerService
}

func NewMarketPriceService(
	marketPriceRepository domain.MarketPriceRepository,
	marketRepository domain.MarketRepository,
	tdexMarketLoaderSvc tdexmarketloader.Service,
	jobPeriodInMinutes string,
	raterSvc port.RateService,
	explorerSvc port.ExplorerService,
) MarketPriceService {
	return &marketPriceService{
		marketPriceRepository:    marketPriceRepository,
		cronSvc:                  cron.New(),
		marketRepository:         marketRepository,
		tdexMarketLoaderSvc:      tdexMarketLoaderSvc,
		fetchPriceCronExpression: fmt.Sprintf("@every %vm", jobPeriodInMinutes),
		raterSvc:                 raterSvc,
		explorerSvc:              explorerSvc,
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

	for k, v := range marketsPrices {
		prices := make([]Price, 0)
		for _, v1 := range v {
			basePriceInRefCurrency, quotePriceInRefCurrency, err := m.getReferencePrices(
				ctx,
				referenceCurrency,
				v1,
			)
			if err != nil {
				return nil, err
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

func (m *marketPriceService) getReferencePrices(
	ctx context.Context,
	referenceCurrency string,
	price domain.MarketPrice,
) (decimal.Decimal, decimal.Decimal, error) {
	var basePriceInRefCurrency, quoteCurrencyInRefCurrency decimal.Decimal
	baseCurrency, err := m.explorerSvc.GetAssetCurrency(ctx, price.BaseAsset)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	baseConvertable := true
	oneUnitOfBasePerRef, err := m.raterSvc.ConvertCurrency(
		ctx,
		baseCurrency,
		referenceCurrency,
	)
	if err != nil {
		if err == port.ErrCurrencyNotFound {
			baseConvertable = false
		} else {
			return decimal.Zero, decimal.Zero, err
		}
	}

	if baseConvertable {
		quoteCurrencyInRefCurrency = oneUnitOfBasePerRef
		basePriceInRefCurrency = decimal.NewFromInt(1).Div(quoteCurrencyInRefCurrency)
	} else {
		quoteConvertable := true
		quoteCurrency, err := m.explorerSvc.GetAssetCurrency(ctx, price.QuoteAsset)
		if err != nil {
			return decimal.Zero, decimal.Zero, err
		}
		oneUnitOfQuotePerRef, err := m.raterSvc.ConvertCurrency(
			ctx,
			quoteCurrency,
			referenceCurrency,
		)
		if err != nil {
			if err == port.ErrCurrencyNotFound {
				quoteConvertable = false
			} else {
				return decimal.Zero, decimal.Zero, err
			}
		}

		if quoteConvertable {
			quoteCurrencyInRefCurrency = oneUnitOfQuotePerRef.Mul(price.QuotePrice)
			basePriceInRefCurrency = decimal.NewFromInt(1).Div(quoteCurrencyInRefCurrency)
		}
	}

	return basePriceInRefCurrency.Round(8), quoteCurrencyInRefCurrency.Round(8), nil
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
