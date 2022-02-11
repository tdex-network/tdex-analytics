package domain

import "time"

type MarketBalance struct {
	MarketID     string
	BaseBalance  int
	BaseAsset    string
	QuoteBalance int
	QuoteAsset   string
	Time         time.Time
}
