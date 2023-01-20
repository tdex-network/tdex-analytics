package inmemory

import (
	"context"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"sync"
)

type inMemoryMarketRepository struct {
	mtx     *sync.RWMutex
	markets map[int]domain.Market
}

func NewRepository() domain.MarketRepository {
	return &inMemoryMarketRepository{
		mtx:     &sync.RWMutex{},
		markets: make(map[int]domain.Market),
	}
}

func (m *inMemoryMarketRepository) InsertMarket(
	ctx context.Context,
	market domain.Market,
) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if market.ID > 0 {
		if _, ok := m.markets[market.ID]; ok {
			return nil
		}
	}

	market.ID = m.getNextID()
	m.markets[market.ID] = market

	return nil
}

func (m *inMemoryMarketRepository) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	resp := make([]domain.Market, 0)
	for _, v := range m.markets {
		resp = append(resp, v)
	}

	return resp, nil
}

func (m *inMemoryMarketRepository) GetMarketsForActiveIndicator(
	ctx context.Context,
	active bool,
) ([]domain.Market, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	resp := make([]domain.Market, 0)
	for _, v := range m.markets {
		if v.Active == active {
			resp = append(resp, v)
		}
	}

	return resp, nil
}

func (m *inMemoryMarketRepository) GetAllMarketsForFilter(
	ctx context.Context,
	filter []domain.Filter,
	_ domain.Page,
) ([]domain.Market, error) {
	resp := make([]domain.Market, 0)
	for _, v := range m.markets {
		for _, v1 := range filter {
			if v.Url == v1.Url && v.QuoteAsset == v1.QuoteAsset && v.BaseAsset == v1.BaseAsset {
				resp = append(resp, v)
			}
		}
	}

	return resp, nil
}

func (m *inMemoryMarketRepository) ActivateMarket(
	ctx context.Context,
	marketID int,
) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, v := range m.markets {
		if v.ID == marketID {
			v.Active = true
			m.markets[v.ID] = v
		}
	}

	return nil
}

func (m *inMemoryMarketRepository) InactivateMarket(
	ctx context.Context,
	marketID int,
) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, v := range m.markets {
		if v.ID == marketID {
			v.Active = false
			m.markets[v.ID] = v
		}
	}

	return nil
}

func (m *inMemoryMarketRepository) getNextID() int {
	lastID := 0
	for k := range m.markets {
		if k > lastID {
			lastID = k
		}
	}

	return lastID + 1
}
