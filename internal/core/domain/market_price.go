package domain

import "time"

type MarketPrice struct {
	MarketID   string
	BasePrice  float32
	BaseAsset  string
	QuotePrice float32
	QuoteAsset string
	Time       time.Time
}
