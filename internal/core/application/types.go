package application

import (
	"encoding/hex"
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/pkg/hexerr"
	"time"
)

const (
	// NIL is added in proto file to recognised when predefined period is passed
	NIL PredefinedPeriod = iota
	LastHour
	LastDay
	LastMonth
	LastThreeMonths
	YearToDate
	All

	StartYear = 2022
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
		Time:         m.Time,
	}, nil
}

type MarketPrice struct {
	MarketID   string
	BasePrice  float32
	BaseAsset  string
	QuotePrice float32
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
		Time:       m.Time,
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
	MarketsBalances map[string][]Balance
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
	MarketsPrices map[string][]Price
}

type Price struct {
	BasePrice  float32
	BaseAsset  string
	QuotePrice float32
	QuoteAsset string
	Time       time.Time
}

type TimeRange struct {
	PredefinedPeriod *PredefinedPeriod
	CustomPeriod     *CustomPeriod
}

func (t *TimeRange) validate() error {
	if t.CustomPeriod == nil && t.PredefinedPeriod == nil {
		return hexerr.NewApplicationLayerError(
			hexerr.InvalidRequest,
			"both PredefinedPeriod period and CustomPeriod cant be null",
		)
	}

	if t.CustomPeriod != nil && t.PredefinedPeriod != nil {
		return hexerr.NewApplicationLayerError(
			hexerr.InvalidRequest,
			"both PredefinedPeriod period and CustomPeriod provided, please provide only one",
		)
	}

	if t.CustomPeriod != nil {
		if err := t.CustomPeriod.validate(); err != nil {
			return err
		}
	}

	if t.PredefinedPeriod != nil {
		if err := t.PredefinedPeriod.validate(); err != nil {
			return err
		}
	}

	return nil
}

type PredefinedPeriod int

func (p *PredefinedPeriod) validate() error {
	if *p > All {
		return hexerr.NewApplicationLayerError(
			hexerr.InvalidRequest,
			fmt.Sprintf("PredefinedPeriod cant be > %v", All),
		)
	}

	return nil
}

type CustomPeriod struct {
	StartDate string
	EndDate   string
}

func (c *CustomPeriod) validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.StartDate, validation.By(validateTimeFormat)),
		validation.Field(&c.EndDate, validation.By(validateTimeFormat)),
	)
}

func (t *TimeRange) getStartAndEndTime(now time.Time) (startTime time.Time, endTime time.Time, err error) {
	if er := t.validate(); er != nil {
		err = er
		return
	}

	if t.CustomPeriod != nil {
		start, _ := time.Parse(time.RFC3339, t.CustomPeriod.StartDate)
		startTime = start

		endTime = now
		if t.CustomPeriod.EndDate != "" {
			end, _ := time.Parse(time.RFC3339, t.CustomPeriod.EndDate)
			endTime = end
		}
		return
	}

	if t.PredefinedPeriod != nil {
		var start time.Time
		switch *t.PredefinedPeriod {
		case LastHour:
			start = now.Add(time.Duration(-60) * time.Minute)
		case LastDay:
			start = now.AddDate(0, 0, -1)
		case LastMonth:
			start = now.AddDate(0, -1, 0)
		case LastThreeMonths:
			start = now.AddDate(0, -3, 0)
		case YearToDate:
			y, _, _ := now.Date()
			start = time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
		case All:
			start = time.Date(StartYear, time.January, 1, 0, 0, 0, 0, time.UTC)
		}

		startTime = start
		endTime = now
	}

	return
}

type MarketProvider struct {
	Url        string
	BaseAsset  string
	QuoteAsset string
}

func (m *MarketProvider) validate() error {
	return validation.ValidateStruct(
		m,
		validation.Field(&m.Url, is.URL),
		validation.Field(&m.BaseAsset, validation.By(validateAssetString)),
		validation.Field(&m.QuoteAsset, validation.By(validateAssetString)),
	)
}

func (m *MarketProvider) toDomain() domain.Filter {
	return domain.Filter{
		Url:        m.Url,
		BaseAsset:  m.BaseAsset,
		QuoteAsset: m.QuoteAsset,
	}
}

type Page domain.Page

func (p *Page) ToDomain() domain.Page {
	return domain.NewPage(p.Number, p.Size)
}

type Market struct {
	ID         int
	Url        string
	BaseAsset  string
	QuoteAsset string
}
