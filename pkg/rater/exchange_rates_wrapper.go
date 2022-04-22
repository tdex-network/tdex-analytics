package rater

import (
	"context"
	"github.com/shopspring/decimal"
	"net/http"
	"tdex-analytics/internal/core/port"
	"time"

	coingecko "github.com/superoo7/go-gecko/v3"
)

const (
	// httpTimeout is the timeout for http requests
	httpTimeout = 10
)

type exchangeRateWrapper struct {
	client *coingecko.Client
}

func NewExchangeRateClient() port.RateService {
	httpClient := &http.Client{
		Timeout: time.Second * httpTimeout,
	}

	client := coingecko.NewClient(httpClient)

	return &exchangeRateWrapper{
		client: client,
	}
}

func (e *exchangeRateWrapper) ConvertCurrency(
	ctx context.Context,
	source string,
	target string,
) (decimal.Decimal, error) {
	panic("implement me")
}
