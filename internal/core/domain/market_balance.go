package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type MarketBalance struct {
	MarketID     string
	BaseBalance  decimal.Decimal
	BaseAsset    string
	QuoteBalance decimal.Decimal
	QuoteAsset   string
	Time         time.Time
}
