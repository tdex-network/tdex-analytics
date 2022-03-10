package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"strconv"
	"tdex-analytics/internal/core/domain"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
	"time"
)

var (
	ErrInvalidTimeFormat = errors.New("fromTime must be valid RFC3339 format")
)

type MarketBalanceService interface {
	// InsertBalance inserts market balance in current moment
	InsertBalance(
		ctx context.Context,
		marketBalance MarketBalance,
	) error
	// GetBalances returns all markets balances from time in past equal to passed arg fromTime
	//if marketID is passed method will return data for all market's, otherwise only for provided one
	GetBalances(
		ctx context.Context,
		timeRange TimeRange,
		page Page,
		marketIDs ...string,
	) (*MarketsBalances, error)
	// StartFetchingBalancesJob starts cron job that will periodically fetch and store balances for all markets
	StartFetchingBalancesJob() error
}

type marketBalanceService struct {
	marketBalanceRepository    domain.MarketBalanceRepository
	marketRepository           domain.MarketRepository
	tdexMarketLoaderSvc        tdexmarketloader.Service
	cronSvc                    *cron.Cron
	fetchBalanceCronExpression string
}

func NewMarketBalanceService(
	marketBalanceRepository domain.MarketBalanceRepository,
	marketRepository domain.MarketRepository,
	tdexMarketLoaderSvc tdexmarketloader.Service,
	jobPeriodInMinutes string,
) MarketBalanceService {

	return &marketBalanceService{
		marketBalanceRepository:    marketBalanceRepository,
		cronSvc:                    cron.New(),
		marketRepository:           marketRepository,
		tdexMarketLoaderSvc:        tdexMarketLoaderSvc,
		fetchBalanceCronExpression: fmt.Sprintf("@every %vm", jobPeriodInMinutes),
	}
}

func (m *marketBalanceService) InsertBalance(
	ctx context.Context,
	marketBalance MarketBalance,
) error {
	if err := marketBalance.validate(); err != nil {
		return err
	}

	mbDomain, err := marketBalance.toDomain()
	if err != nil {
		return err
	}

	return m.marketBalanceRepository.InsertBalance(ctx, *mbDomain)
}

func (m *marketBalanceService) GetBalances(
	ctx context.Context,
	timeRange TimeRange,
	page Page,
	marketIDs ...string,
) (*MarketsBalances, error) {
	result := make(map[string][]Balance)

	startTime, endTime, err := timeRange.getStartAndEndTime(time.Now())
	if err != nil {
		return nil, err
	}

	marketsBalances, err := m.marketBalanceRepository.GetBalancesForMarkets(
		ctx,
		startTime,
		endTime,
		page.ToDomain(),
		marketIDs...,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range marketsBalances {
		balances := make([]Balance, 0)
		for _, v1 := range v {
			balances = append(balances, Balance{
				BaseBalance:  v1.BaseBalance,
				BaseAsset:    v1.BaseAsset,
				QuoteBalance: v1.QuoteBalance,
				QuoteAsset:   v1.QuoteAsset,
				Time:         v1.Time,
			})
		}

		result[k] = balances
	}

	return &MarketsBalances{
		MarketsBalances: result,
	}, nil
}

func (m *marketBalanceService) StartFetchingBalancesJob() error {
	if _, err := m.cronSvc.AddJob(
		m.fetchBalanceCronExpression,
		cron.FuncJob(m.FetchBalancesForAllMarkets),
	); err != nil {
		return err
	}

	m.cronSvc.Start()

	return nil
}

func (m *marketBalanceService) FetchBalancesForAllMarkets() {
	log.Infof("job FetchBalancesForAllMarkets at: %v", time.Now())
	ctx := context.Background()

	markets, err := m.marketRepository.GetAllMarkets(ctx)
	if err != nil {
		log.Errorf("FetchBalancesForAllMarkets -> GetAllMarkets: %v", err)
		return
	}

	for _, v := range markets {
		go func(market domain.Market) {
			m.FetchAndInsertBalance(ctx, market)
		}(v)
	}
}

func (m *marketBalanceService) FetchAndInsertBalance(
	ctx context.Context,
	market domain.Market,
) {
	balance, err := m.tdexMarketLoaderSvc.FetchBalance(
		ctx,
		tdexmarketloader.Market{
			Url:        market.Url,
			QuoteAsset: market.QuoteAsset,
			BaseAsset:  market.BaseAsset,
		},
	)
	if err != nil {
		log.Errorf("FetchAndInsertBalance -> FetchBalance: %v", err)
		return
	}

	if err := m.InsertBalance(ctx, MarketBalance{
		MarketID:     strconv.Itoa(market.ID),
		BaseBalance:  balance.BaseBalance,
		BaseAsset:    market.BaseAsset,
		QuoteBalance: balance.QuoteBalance,
		QuoteAsset:   market.QuoteAsset,
		Time:         time.Now(),
	}); err != nil {
		log.Errorf("FetchAndInsertBalance -> InsertBalance: %v", err)
		return
	}
}
