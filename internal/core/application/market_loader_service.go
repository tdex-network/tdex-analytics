package application

import (
	"context"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"tdex-analytics/internal/core/domain"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
	"time"
)

const (
	//every day at 00:00
	fetchMarketsCronExpression = "0 0 * * *"
)

type MarketsLoaderService interface {
	StartFetchingMarketsJob() error
}

type marketsLoaderService struct {
	marketRepository    domain.MarketRepository
	tdexMarketLoaderSvc tdexmarketloader.Service
	cronSvc             *cron.Cron
}

func NewMarketsLoaderService(
	marketRepository domain.MarketRepository,
	tdexMarketLoaderSvc tdexmarketloader.Service,
) MarketsLoaderService {
	return &marketsLoaderService{
		marketRepository:    marketRepository,
		tdexMarketLoaderSvc: tdexMarketLoaderSvc,
		cronSvc:             cron.New(),
	}
}

func (m *marketsLoaderService) StartFetchingMarketsJob() error {
	// run initially
	go m.FetchMarkets()

	if _, err := m.cronSvc.AddJob(
		fetchMarketsCronExpression,
		cron.FuncJob(m.FetchMarkets),
	); err != nil {
		return err
	}

	m.cronSvc.Start()

	return nil
}

func (m *marketsLoaderService) FetchMarkets() {
	log.Infof("job FetchMarkets at: %v", time.Now())
	//TODO add context with timeout
	liquidityProviders, err := m.tdexMarketLoaderSvc.FetchProvidersMarkets(
		context.Background(),
	)
	if err != nil {
		log.Errorf("FetchMarkets -> FetchProvidersMarkets: %v", err)
		return
	}

	for _, v := range liquidityProviders {
		for _, v1 := range v.Markets {
			if err := m.marketRepository.InsertMarket(context.Background(), domain.Market{
				ProviderName: v.Name,
				Url:          v.Endpoint,
				BaseAsset:    v1.BaseAsset,
				QuoteAsset:   v1.QuoteAsset,
			}); err != nil {
				log.Errorf("FetchMarkets -> InsertMarket: %v", err)
			}
		}
	}
}
