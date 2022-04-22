package port

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
)

var (
	ErrCurrencyNotFound = errors.New("can't convert currencies, currency not found")
)

type RateService interface {
	// ConvertCurrency converts 1 unit of source currency to target currency
	//e.g. ConvertCurrency("EUR", "USD") would convert 1 EUR to USD (1.12 USD)
	ConvertCurrency(ctx context.Context, source, target string) (decimal.Decimal, error)
}
