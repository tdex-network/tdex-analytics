package application

import (
	"encoding/hex"
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/pkg/hexerr"
	"time"
)

type MarketBalance struct {
	MarketID     string
	BaseBalance  int
	BaseAsset    string
	QuoteBalance int
	QuoteAsset   string
	Time         time.Time
}

func (m *MarketBalance) validate() error {
	return validation.ValidateStruct(
		m,
		validation.Field(&m.MarketID, validation.Required),
		validation.Field(&m.BaseAsset, validation.By(validateAssetString)),
		validation.Field(&m.QuoteAsset, validation.By(validateAssetString)),
		validation.Field(&m.Time, validation.By(validateTimeFormat)),
	)
}

func (m *MarketBalance) toDomain() (*domain.MarketBalance, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &domain.MarketBalance{
		MarketID:     m.MarketID,
		BaseBalance:  m.BaseBalance,
		BaseAsset:    m.BaseAsset,
		QuoteBalance: m.QuoteBalance,
		QuoteAsset:   m.QuoteAsset,
	}, nil
}

type MarketPrice struct {
	MarketID   string
	BasePrice  int
	BaseAsset  string
	QuotePrice int
	QuoteAsset string
	Time       time.Time
}

func (m *MarketPrice) validate() error {
	return validation.ValidateStruct(
		m,
		validation.Field(&m.MarketID, validation.Required),
		validation.Field(&m.BaseAsset, validation.By(validateAssetString)),
		validation.Field(&m.QuoteAsset, validation.By(validateAssetString)),
	)
}

func (m *MarketPrice) toDomain() (*domain.MarketPrice, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &domain.MarketPrice{
		MarketID:   m.MarketID,
		BasePrice:  m.BasePrice,
		BaseAsset:  m.BaseAsset,
		QuotePrice: m.QuotePrice,
		QuoteAsset: m.QuoteAsset,
	}, nil
}

func validateAssetString(asset interface{}) error {
	a, ok := asset.(string)
	if !ok {
		return errors.New("must be a valid asset string")
	}

	buf, err := hex.DecodeString(a)
	if err != nil {
		return errors.New("asset is not in hex format")
	}

	if len(buf) != 32 {
		return errors.New("asset length is invalid")
	}

	return nil
}

func validateTimeFormat(t interface{}) error {
	tm, ok := t.(string)
	if !ok {
		return errors.New("must be a valid time.Time")
	}

	if _, err := time.Parse(time.RFC3339, tm); err != nil {
		return hexerr.NewApplicationLayerError(
			hexerr.InvalidRequest,
			ErrInvalidTimeFormat.Error(),
		)
	}

	return nil
}

type MarketsBalances struct {
	//market_id and its Balances
	MarketsBalances map[int][]Balance
}

type Balance struct {
	BaseBalance  int
	BaseAsset    string
	QuoteBalance int
	QuoteAsset   string
	Time         time.Time
}

type MarketsPrices struct {
	//market_id and its Prices
	MarketsPrices map[int][]Price
}

type Price struct {
	BasePrice  int
	BaseAsset  string
	QuotePrice int
	QuoteAsset string
	Time       time.Time
}
