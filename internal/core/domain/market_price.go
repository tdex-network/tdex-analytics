package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type MarketPrice struct {
	MarketID   string
	BasePrice  decimal.Decimal
	BaseAsset  string
	QuotePrice decimal.Decimal
	QuoteAsset string
	Time       time.Time
}
