package rater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/port"
	"golang.org/x/time/rate"
)

const (
	// httpTimeout is the timeout for http requests
	httpTimeout = 10
	// exchangeRateApiUrl is the url of the exchange rate wrapper
	exchangeRateApiUrl = "https://open.er-api.com/v6/latest"
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

type ratesCache struct {
	rates      map[string]decimal.Decimal
	lastUpdate time.Time
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

	// symbols is a cache of the fiat symbols supported, it is fetched only once
	// at startup
	symbols map[string]struct{}

	// cache for fiat rates
	ratesCache map[string]ratesCache
	ratesLock  *sync.Mutex

	// rateLimiter is the rate limiter(token bucket) for the coin gecko api
	rateLimiter *rate.Limiter
}

func NewExchangeRateClient(
	assetCurrencySymbolPair map[string]string,
	coinGeckoNumOfCallsPerMin *int,
	coinGeckoRefreshInterval *time.Duration,
	coinGeckoWaitDuration *time.Duration,
) (port.RateService, error) {
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

	symbols, err := fetchSymbols(httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch symbols: %s", err)
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
		symbols:     symbols,
		ratesCache:  make(map[string]ratesCache),
		ratesLock:   &sync.Mutex{},
		rateLimiter: rate.NewLimiter(rate.Every(time.Minute), numOfCallsPerMin),
	}, nil
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

	isFiatSymbol, err := e.IsFiatSymbolSupported(target)
	if err != nil {
		return decimal.Zero, err
	}

	if isFiatSymbol {
		return e.getFiatToFiatRate(source, target)
	}

	isCryptoSymbol, _ := e.isCryptoSymbol(ctx, e.coinGeckoWaitDuration, source)
	if !isCryptoSymbol {
		return decimal.Zero, fmt.Errorf("%s is not a supported fiat nor crypto symbol", source)
	}

	return e.getCryptoToFiatRate(ctx, source, target)
}

func (e *exchangeRateWrapper) IsFiatSymbolSupported(symbol string) (bool, error) {
	_, ok := e.symbols[strings.ToLower(symbol)]
	return ok, nil
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
// data are fetched from coin gecko and in order to prevent rate limit errors, the
// rates are cached and reloaded every coinGeckoRefreshInterval
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

		quotePerBase := e.exchangeRates[quote][base]
		return quotePerBase.baseRate, nil
	}

	return v.baseRate, nil
}

func (e *exchangeRateWrapper) getFiatToFiatRate(
	source string,
	target string,
) (decimal.Decimal, error) {
	e.ratesLock.Lock()
	defer e.ratesLock.Unlock()
	// Update cache once a day
	if cache, ok := e.ratesCache[target]; !ok || time.Since(cache.lastUpdate).Hours() >= 24 {
		data, err := fetchRates(e.httpClient, target)
		if err != nil {
			if !ok {
				return decimal.Zero, err
			}
			return cache.rates[source], nil
		}
		e.ratesCache[target] = ratesCache{
			rates:      data.rates,
			lastUpdate: time.Now(),
		}
	}

	return e.ratesCache[target].rates[source], nil
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

func fetchSymbols(httpClient *http.Client) (map[string]struct{}, error) {
	data, err := fetchRates(httpClient, "usd")
	if err != nil {
		return nil, err
	}

	symbols := make(map[string]struct{})
	for symbol := range data.rates {
		symbols[strings.ToLower(symbol)] = struct{}{}
	}

	return symbols, nil
}

type rateResponse struct {
	base  string
	date  string
	rates map[string]decimal.Decimal
}

func fetchRates(httpClient *http.Client, base string) (*rateResponse, error) {
	base = strings.ToUpper(base)
	url := fmt.Sprintf("%s/%s", exchangeRateApiUrl, base)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"unexpected status code: %d, error: %v",
			resp.StatusCode,
			resp.Body,
		)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := make(map[string]interface{})
	if err := json.Unmarshal(buf, &body); err != nil {
		return nil, err
	}

	base, ok := body["base_code"].(string)
	if !ok {
		return nil, fmt.Errorf("base code not found")
	}
	date, ok := body["time_last_update_utc"].(string)
	if !ok {
		return nil, fmt.Errorf("date not found")
	}
	r, ok := body["rates"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("rates list not found")
	}
	rates := make(map[string]decimal.Decimal)
	for symbol, rate := range r {
		rates[strings.ToLower(symbol)] = decimal.NewFromFloat(rate.(float64))
	}

	return &rateResponse{
		base:  strings.ToLower(base),
		date:  date,
		rates: rates,
	}, nil
}
