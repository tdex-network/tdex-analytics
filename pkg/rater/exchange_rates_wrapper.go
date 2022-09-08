package rater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tdex-network/tdex-analytics/internal/core/port"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const (
	// httpTimeout is the timeout for http requests
	httpTimeout = 10
	// exchangeRateApiUrl is the url of the exchange rate wrapper
	exchangeRateApiUrl = "https://api.exchangerate.host"
	coinGeckoBtcID     = "bitcoin"
	btcSymbol          = "btc"
	lbtcSymbol         = "lbtc"

	// defaultCoinGeckoRefreshInterval is the default interval for refreshing the coin
	//list and exchange rates fetched from Coin Gecko
	defaultCoinGeckoRefreshInterval = time.Minute * 5
	// defaultNumOfCallsPerMin is the default number of calls per minute for the rate limiter
	defaultCoinGeckoNumOfCallsPerMin = 50
	// defaultCoinGeckoWaitDuration is the default duration for waiting for the
	//rate limiter to allow call to coin gecko
	defaultCoinGeckoWaitDuration = time.Second * httpTimeout
)

var (
	ErrCoinGeckoWaitDuration = errors.New("coin gecko wait duration exceeded")
)

type baseCurrency string
type quoteCurrency string
type baseRatesInfo struct {
	baseRate         decimal.Decimal
	refreshTimestamp time.Time
}
type coinListInfo struct {
	coinList         map[string]string
	refreshTimestamp time.Time
}

type exchangeRateWrapper struct {
	httpClient *http.Client

	coinGeckoSvc CoinGeckoService
	// coinGeckoRefreshInterval is the interval for refreshing the coin list and
	//exchange rates fetched from Coin Gecko
	coinGeckoRefreshInterval time.Duration
	// coinGeckoWaitDuration is the duration for waiting for the rate limiter
	//to allow call to coin gecko
	coinGeckoWaitDuration time.Duration

	assetCurrencySymbolPair map[string]string

	// exchangeRatesMtx is the mutex for the exchange rates map
	exchangeRatesMtx sync.RWMutex
	// exchangeRates is a cache of exchange rates fetched from Coin Gecko,
	//it is refreshed every coinGeckoRefreshInterval
	exchangeRates map[quoteCurrency]map[baseCurrency]baseRatesInfo

	// coinListMtx is a mutex for the coin list
	coinListMtx sync.RWMutex
	// coins is a cache of the coin list from coin gecko, it is refreshed every
	//coinGeckoRefreshInterval
	coins coinListInfo

	// rateLimiter is the rate limiter(token bucket) for the coin gecko api
	rateLimiter *rate.Limiter
}

func NewExchangeRateClient(
	assetCurrencySymbolPair map[string]string,
	coinGeckoNumOfCallsPerMin *int,
	coinGeckoRefreshInterval *time.Duration,
	coinGeckoWaitDuration *time.Duration,
) port.RateService {
	httpClient := &http.Client{
		Timeout: time.Second * httpTimeout,
	}

	coinGeckoSvc := NewCoinGeckoService(httpClient)

	numOfCallsPerMin := defaultCoinGeckoNumOfCallsPerMin
	if coinGeckoNumOfCallsPerMin != nil {
		numOfCallsPerMin = *coinGeckoNumOfCallsPerMin
	}

	refreshInterval := defaultCoinGeckoRefreshInterval
	if coinGeckoRefreshInterval != nil {
		refreshInterval = *coinGeckoRefreshInterval
	}

	waitDuration := defaultCoinGeckoWaitDuration
	if coinGeckoWaitDuration != nil {
		waitDuration = *coinGeckoWaitDuration
	}

	return &exchangeRateWrapper{
		httpClient:               httpClient,
		coinGeckoSvc:             coinGeckoSvc,
		coinGeckoRefreshInterval: refreshInterval,
		coinGeckoWaitDuration:    waitDuration,
		assetCurrencySymbolPair:  assetCurrencySymbolPair,
		exchangeRates:            make(map[quoteCurrency]map[baseCurrency]baseRatesInfo),
		exchangeRatesMtx:         sync.RWMutex{},
		coins: coinListInfo{
			coinList: make(map[string]string),
		},
		coinListMtx: sync.RWMutex{},
		rateLimiter: rate.NewLimiter(rate.Every(time.Minute), numOfCallsPerMin),
	}
}

func (e *exchangeRateWrapper) ConvertCurrency(
	ctx context.Context,
	source string,
	target string,
) (decimal.Decimal, error) {
	source = strings.ToLower(source)
	target = strings.ToLower(target)

	if source == btcSymbol || source == lbtcSymbol {
		source = coinGeckoBtcID
	}

	if source == target {
		return decimal.NewFromFloat(1), nil
	}

	isCryptoSymbol, err := e.isCryptoSymbol(ctx, e.coinGeckoWaitDuration, source)
	if err != nil {
		return decimal.Zero, err
	}

	if isCryptoSymbol {
		return e.getCryptoToFiatRate(ctx, source, target)
	}

	sourceValidFiatSymbol, err := e.IsFiatSymbolSupported(target)
	if err != nil {
		return decimal.Zero, err
	}

	if !isCryptoSymbol && !sourceValidFiatSymbol {
		return decimal.Zero, fmt.Errorf("%s is not a supported fiat nor crypto symbol", source)
	}

	return e.getFiatToFiatRate(ctx, source, target)
}

func (e *exchangeRateWrapper) IsFiatSymbolSupported(symbol string) (bool, error) {
	symbol = strings.ToLower(symbol)

	urlSymbols := fmt.Sprintf("%s/%s", exchangeRateApiUrl, "symbols")
	resp, err := e.httpClient.Get(urlSymbols)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf(
			"unexpected status code: %d, error: %v",
			resp.StatusCode,
			resp.Body,
		)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return false, err
	}
	respStr := buf.String()

	return strings.Contains(respStr, fmt.Sprintf("\"%s\"", strings.ToUpper(symbol))), nil
}

func (e *exchangeRateWrapper) GetAssetCurrency(
	assetId string,
) (string, error) {
	currency, ok := e.assetCurrencySymbolPair[assetId]
	if !ok {
		return "", fmt.Errorf("asset %s not found", assetId)
	}

	return currency, nil
}

type CryptoCoin struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func (e *exchangeRateWrapper) isCryptoSymbol(
	ctx context.Context,
	waitDuration time.Duration,
	symbol string,
) (bool, error) {
	symbol = strings.ToLower(symbol)

	if len(e.coins.coinList) == 0 {
		if err := e.reloadCoinList(ctx, waitDuration); err != nil {
			return false, err
		}
	}

	_, ok := e.coins.coinList[symbol]
	if e.coins.refreshTimestamp.Add(e.coinGeckoRefreshInterval).Before(time.Now()) || !ok {
		if err := e.reloadCoinList(ctx, waitDuration); err != nil {
			return false, err
		}

		_, ok1 := e.coins.coinList[symbol]
		return ok1, nil
	}

	return ok, nil
}

// getCryptoToFiatRate returns the rate of one source crypt coin to the target fiat
//data are fetched from coin gecko and in order to prevent rate limit errors, the
//rates are cached and reloaded every coinGeckoRefreshInterval
func (e *exchangeRateWrapper) getCryptoToFiatRate(
	ctx context.Context,
	source string,
	target string,
) (decimal.Decimal, error) {
	source = strings.ToLower(source)
	target = strings.ToLower(target)

	quote := quoteCurrency(source)
	base := baseCurrency(target)

	v, ok := e.exchangeRates[quote][base]
	// if the rate is not found or data are old, reload the exchange rates, else return from cache
	if v.refreshTimestamp.Add(e.coinGeckoRefreshInterval).Before(time.Now()) || !ok {
		if err := e.reloadQuoteBasePair(ctx, e.coinGeckoWaitDuration, quote, base); err != nil {
			return decimal.Decimal{}, err
		}

		quotePerBase, _ := e.exchangeRates[quote][base]
		return quotePerBase.baseRate, nil
	}

	return v.baseRate, nil
}

func (e *exchangeRateWrapper) getFiatToFiatRate(
	ctx context.Context,
	source string,
	target string,
) (decimal.Decimal, error) {
	source = strings.ToUpper(source)
	target = strings.ToUpper(target)

	params := map[string]string{
		"base":    source,
		"symbols": target,
		"date":    time.Now().Format("2006-01-02"),
	}

	rates, err := e.fetchRates(params)
	if err != nil {
		return decimal.Zero, err
	}

	// current api provider returns target conversion to eur if source is not supported
	if rates.Base != source {
		return decimal.Zero, port.ErrCurrencyNotFound
	}

	return decimal.NewFromFloat(rates.Rates[target]), nil
}

type RateResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func (e *exchangeRateWrapper) fetchRates(params map[string]string) (RateResponse, error) {
	var result RateResponse

	urlRates, err := url.Parse(fmt.Sprintf("%s/%s", exchangeRateApiUrl, params["date"]))
	if err != nil {
		return result, err
	}

	urlRates.Path = "latest"
	if len(params) > 0 {
		q := urlRates.Query()
		if params["base"] != "" {
			q.Set("base", strings.ToUpper(params["base"]))
		}
		if params["symbols"] != "" {
			q.Set("symbols", strings.ToUpper(params["symbols"]))
		}
		urlRates.RawQuery = q.Encode()
	}

	resp, err := e.httpClient.Get(urlRates.String())
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return result, fmt.Errorf(
			"unexpected status code: %d, error: %v",
			resp.StatusCode,
			resp.Body,
		)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// reloadCoinList reloads the coin list from coin gecko
func (e *exchangeRateWrapper) reloadCoinList(
	ctx context.Context,
	waitTimeout time.Duration,
) error {
	e.coinListMtx.Lock()
	defer e.coinListMtx.Unlock()

	//check if allowed number of requests is exceeded, if yes wait for the next period
	//but wait for waitDuration interval at least
	if !e.rateLimiter.Allow() {
		ctx, cancel := context.WithTimeout(ctx, waitTimeout)
		defer cancel()

		if err := e.rateLimiter.Wait(ctx); err != nil {
			return ErrCoinGeckoWaitDuration
		}
	}

	list, err := e.coinGeckoSvc.CoinsList()
	if err != nil {
		return err
	}

	if list == nil {
		return fmt.Errorf("coin list returned empty list")
	}

	e.coins.coinList = make(map[string]string)
	for _, coin := range *list {
		e.coins.coinList[coin.ID] = coin.Symbol
	}

	e.coins.refreshTimestamp = time.Now()

	return nil
}

// reloadQuoteBasePair reloads the quote per base rate from coinGecko APIr.
func (e *exchangeRateWrapper) reloadQuoteBasePair(
	ctx context.Context,
	waitDuration time.Duration,
	quote quoteCurrency,
	base baseCurrency,
) error {
	e.exchangeRatesMtx.Lock()
	defer e.exchangeRatesMtx.Unlock()

	//check if allowed number of requests is exceeded, if yes wait for the next period
	//but wait for waitDuration interval at least
	if !e.rateLimiter.Allow() {
		ctx, cancel := context.WithTimeout(ctx, waitDuration)
		defer cancel()

		if err := e.rateLimiter.Wait(ctx); err != nil {
			return ErrCoinGeckoWaitDuration
		}
	}

	price, err := e.coinGeckoSvc.SimplePrice(
		[]string{string(quote)},
		[]string{string(base)},
	)
	if err != nil {
		return err
	}

	fValue := float64((*price)[string(quote)][string(base)])
	if fValue == 0 {
		return port.ErrCurrencyNotFound
	}

	if _, ok := e.exchangeRates[quote]; !ok {
		e.exchangeRates[quote] = make(map[baseCurrency]baseRatesInfo)
	}

	e.exchangeRates[quote][base] = baseRatesInfo{
		baseRate:         decimal.NewFromFloat(fValue),
		refreshTimestamp: time.Now(),
	}

	return nil
}
