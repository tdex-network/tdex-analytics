package rater

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tdex-network/tdex-analytics/internal/core/port"
	"golang.org/x/time/rate"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/dnaeon/go-vcr/recorder"
)

func TestExchangeRateWrapperGetFiatToFiatRate(t *testing.T) {
	e := &exchangeRateWrapper{
		httpClient: &http.Client{
			Timeout: time.Second * httpTimeout,
		},
	}

	_, err := e.getFiatToFiatRate("eur", "usd")
	assert.NoError(t, err)
}

func TestExchangeRateWrapperIsCryptoSymbol(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	e := &exchangeRateWrapper{
		coinGeckoSvc:  NewCoinGeckoService(httpClient),
		rateLimiter:   rate.NewLimiter(rate.Every(time.Minute), 50),
		exchangeRates: make(map[quoteCurrency]map[baseCurrency]baseRatesInfo),
	}

	isCryptoSymbol, err := e.isCryptoSymbol(context.Background(), time.Second*30, "bitcoin")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, isCryptoSymbol)
}

func TestExchangeRateWrapperGetCryptoToFiatRate(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	e := &exchangeRateWrapper{
		coinGeckoSvc:  NewCoinGeckoService(httpClient),
		rateLimiter:   rate.NewLimiter(rate.Every(time.Minute), 50),
		exchangeRates: make(map[quoteCurrency]map[baseCurrency]baseRatesInfo),
	}

	val, err := e.getCryptoToFiatRate(context.Background(), "BITCOIN", "EUR")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, val.GreaterThan(decimal.Zero))
}

func TestExchangeRateWrapperGetCryptoToFiatRateLimiterWaitInterval(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	e := &exchangeRateWrapper{
		coinGeckoSvc:             NewCoinGeckoService(httpClient),
		rateLimiter:              rate.NewLimiter(rate.Every(time.Minute), 1),
		exchangeRates:            make(map[quoteCurrency]map[baseCurrency]baseRatesInfo),
		coinGeckoRefreshInterval: time.Second * 30,
		coinGeckoWaitDuration:    time.Second * 5,
	}

	val, err := e.getCryptoToFiatRate(context.Background(), "BITCOIN", "EUR")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, val.GreaterThan(decimal.Zero))

	if err := vcrRecorder.Stop(); err != nil {
		t.Error(err)
	}
}

func TestExchangeRateWrapperConvertCurrency(t *testing.T) {
	client, err := NewExchangeRateClient(
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	result, err := client.ConvertCurrency(context.Background(), "EUR", "USD")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, result.GreaterThan(decimal.Zero))
}

func TestExchangeRateWrapperConvertCurrencyNegativeScenario(t *testing.T) {
	client, err := NewExchangeRateClient(
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	_, err = client.ConvertCurrency(context.Background(), "dwdw", "eur")
	assert.Equal(t, port.ErrCurrencyNotFound.Error(), err.Error())
}

func TestCoinGeckoSimplePriceNotInvokedWhenFreshDataInCache(t *testing.T) {
	coinGeckoSvcMock := new(MockCoinGeckoService)

	exchangeRates := make(map[quoteCurrency]map[baseCurrency]baseRatesInfo)
	exchangeRates["bitcoin"] = make(map[baseCurrency]baseRatesInfo)
	exchangeRates["bitcoin"]["eur"] = baseRatesInfo{
		baseRate:         decimal.NewFromFloat(21347),
		refreshTimestamp: time.Now(),
	}

	e := &exchangeRateWrapper{
		coinGeckoSvc:             coinGeckoSvcMock,
		coinGeckoRefreshInterval: time.Minute,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Minute), 50),
		exchangeRates:            exchangeRates,
	}

	val, err := e.getCryptoToFiatRate(context.Background(), "BITCOIN", "EUR")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val.GreaterThan(decimal.Zero))
	coinGeckoSvcMock.AssertNotCalled(t, "SimplePrice")
}

func TestCoinGeckoSimplePriceInvokedWhenStaleDataInCache(t *testing.T) {
	coinGeckoSvcMock := new(MockCoinGeckoService)
	mockResp := map[string]map[string]float32{
		"bitcoin": {
			"eur": 21347,
		},
	}
	coinGeckoSvcMock.On("SimplePrice", []string{"bitcoin"}, []string{"eur"}).
		Return(&mockResp, nil)

	exchangeRates := make(map[quoteCurrency]map[baseCurrency]baseRatesInfo)
	exchangeRates["bitcoin"] = make(map[baseCurrency]baseRatesInfo)
	exchangeRates["bitcoin"]["eur"] = baseRatesInfo{
		baseRate:         decimal.NewFromFloat(21347),
		refreshTimestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	e := &exchangeRateWrapper{
		coinGeckoSvc:             coinGeckoSvcMock,
		coinGeckoRefreshInterval: time.Minute,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Minute), 50),
		exchangeRates:            exchangeRates,
	}

	val, err := e.getCryptoToFiatRate(context.Background(), "BITCOIN", "EUR")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val.GreaterThan(decimal.Zero))
	coinGeckoSvcMock.AssertCalled(t, "SimplePrice", []string{"bitcoin"}, []string{"eur"})
}

func TestCoinGeckoCoinsListNotInvokedWhenFreshDataInCache(t *testing.T) {
	coinGeckoSvcMock := new(MockCoinGeckoService)

	e := &exchangeRateWrapper{
		coinGeckoSvc:             coinGeckoSvcMock,
		coinGeckoRefreshInterval: time.Minute,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Minute), 50),
		coins: coinListInfo{
			refreshTimestamp: time.Now(),
			coinList:         map[string]string{"bitcoin": "btc"},
		},
	}

	val, err := e.isCryptoSymbol(context.Background(), 0, "bitcoin")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val)
	coinGeckoSvcMock.AssertNotCalled(t, "CoinsList")
}

func TestCoinGeckoCoinsListInvokedWhenStaleDataInCache(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	coinGeckoSvc := NewCoinGeckoService(httpClient)
	list, err := coinGeckoSvc.CoinsList()
	if err != nil {
		t.Error(err)
	}

	coinGeckoSvcMock := new(MockCoinGeckoService)
	coinGeckoSvcMock.On("CoinsList").Return(list, nil)

	e := &exchangeRateWrapper{
		coinGeckoSvc:             coinGeckoSvcMock,
		coinGeckoRefreshInterval: time.Minute,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Minute), 50),
		coins: coinListInfo{
			refreshTimestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			coinList:         map[string]string{"bitcoin": "btc"},
		},
	}

	val, err := e.isCryptoSymbol(context.Background(), 0, "bitcoin")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val)
	coinGeckoSvcMock.AssertCalled(t, "CoinsList")
}

func TestGetCryptoToFiatRateLimiter(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	exchangeRates := make(map[quoteCurrency]map[baseCurrency]baseRatesInfo)

	e := &exchangeRateWrapper{
		coinGeckoSvc:             NewCoinGeckoService(httpClient),
		coinGeckoRefreshInterval: time.Minute,
		coinGeckoWaitDuration:    time.Second,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Second*5), 1),
		exchangeRates:            exchangeRates,
	}

	val, err := e.getCryptoToFiatRate(context.Background(), "bitcoin", "eur")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val.GreaterThan(decimal.Zero))

	//reset exchange rates so coin gecko is called again, and we can test the limiter
	newData := make(map[quoteCurrency]map[baseCurrency]baseRatesInfo)
	e.exchangeRates = newData

	_, err = e.getCryptoToFiatRate(context.Background(), "bitcoin", "eur")
	assert.Equal(t, err, ErrCoinGeckoWaitDuration)

	//reset wait duration so that req is waiting limiter to allow call
	e.coinGeckoWaitDuration = time.Second * 10

	val, err = e.getCryptoToFiatRate(context.Background(), "bitcoin", "eur")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, val.GreaterThan(decimal.Zero))
}

func TestIsCryptoSymbolLimiter(t *testing.T) {
	vcrRecorder, err := recorder.New(fmt.Sprintf("fixtures/coingecko_%v", t.Name()))
	if err != nil {
		t.Error(err)
	}
	defer vcrRecorder.Stop()

	httpClient := &http.Client{
		Transport: vcrRecorder,
		Timeout:   time.Second * httpTimeout,
	}

	exchangeRates := make(map[quoteCurrency]map[baseCurrency]baseRatesInfo)

	e := &exchangeRateWrapper{
		coinGeckoSvc:             NewCoinGeckoService(httpClient),
		coinGeckoRefreshInterval: time.Minute,
		rateLimiter:              rate.NewLimiter(rate.Every(time.Second*5), 1),
		exchangeRates:            exchangeRates,
	}

	waitDuration := time.Second * 1

	val, err := e.isCryptoSymbol(context.Background(), waitDuration, "bitcoin")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, val)

	//reset coins so coin gecko is called again, and we can test the limiter
	e.coins = coinListInfo{
		coinList: nil,
	}

	_, err = e.isCryptoSymbol(context.Background(), 0, "bitcoin")
	assert.Equal(t, err, ErrCoinGeckoWaitDuration)

	//reset wait duration so that req is waiting limiter to allow call
	waitDuration = time.Second * 10

	val, err = e.isCryptoSymbol(context.Background(), waitDuration, "bitcoin")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, true, val)
}
