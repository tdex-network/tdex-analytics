package tdexmarketloader

import "github.com/shopspring/decimal"

type Market struct {
	Url        string
	QuoteAsset string
	BaseAsset  string
}

type LiquidityProvider struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Markets  []Market
}

type Balance struct {
	BaseBalance  decimal.Decimal
	QuoteBalance decimal.Decimal
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}
