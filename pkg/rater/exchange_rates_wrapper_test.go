package rater

import (
	"context"
	"net/http"
	"tdex-analytics/internal/core/port"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	coingecko "github.com/superoo7/go-gecko/v3"
)

func TestExchangeRateWrapperGetFiatToFiatRate(t *testing.T) {
	e := &exchangeRateWrapper{
		httpClient: &http.Client{
			Timeout: time.Second * httpTimeout,
		},
	}
	_, err := e.getFiatToFiatRate(context.Background(), "ETH", "EUR")
	assert.Error(t, err)

	_, err = e.getFiatToFiatRate(context.Background(), "EUR", "USD")
	assert.NoError(t, err)

	_, err = e.getFiatToFiatRate(context.Background(), "eur", "usd")
	assert.NoError(t, err)
}

func TestExchangeRateWrapperIsCryptoSymbol(t *testing.T) {
	httpClient := &http.Client{
		Timeout: time.Second * httpTimeout,
	}
	e := &exchangeRateWrapper{
		coinGeckoClient: coingecko.NewClient(httpClient),
	}

	isCryptoSymbol, err := e.isCryptoSymbol("EUR")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, false, isCryptoSymbol)

	isCryptoSymbol, err = e.isCryptoSymbol("bitcoin")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, isCryptoSymbol)
}

func TestExchangeRateWrapperGetCryptoToFiatRate(t *testing.T) {
	httpClient := &http.Client{
		Timeout: time.Second * httpTimeout,
	}
	e := &exchangeRateWrapper{
		coinGeckoClient: coingecko.NewClient(httpClient),
	}

	rate, err := e.getCryptoToFiatRate(context.Background(), "BITCOIN", "EUR")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, rate.GreaterThan(decimal.Zero))
}

func TestExchangeRateWrapperConvertCurrency(t *testing.T) {
	client := NewExchangeRateClient(nil)
	result, err := client.ConvertCurrency(context.Background(), "EUR", "USD")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))

	result, err = client.ConvertCurrency(context.Background(), "USD", "EUR")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))

	result, err = client.ConvertCurrency(context.Background(), "bitcoin", "USD")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))

	result, err = client.ConvertCurrency(context.Background(), "bitcoin", "USD")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))

	result, err = client.ConvertCurrency(context.Background(), "btc", "USD")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))
}

func TestExchangeRateWrapperConvertCurrencyNegativeScenario(t *testing.T) {
	client := NewExchangeRateClient(nil)
	_, err := client.ConvertCurrency(context.Background(), "dwdw", "eur")
	assert.Equal(t, port.ErrCurrencyNotFound.Error(), err.Error())

	_, err = client.ConvertCurrency(context.Background(), "btc", "dwdw")
	assert.Equal(t, port.ErrCurrencyNotFound.Error(), err.Error())
}
