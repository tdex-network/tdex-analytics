package domain

import "time"

type MarketPrice struct {
	MarketID   string
	BasePrice  int
	BaseAsset  string
	QuotePrice int
	QuoteAsset string
	Time       time.Time
}
