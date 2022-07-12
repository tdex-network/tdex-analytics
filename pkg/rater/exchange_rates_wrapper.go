package rater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"tdex-analytics/internal/core/port"
	"time"

	"github.com/shopspring/decimal"

	coingecko "github.com/superoo7/go-gecko/v3"
)

const (
	// httpTimeout is the timeout for http requests
	httpTimeout = 10
	// exchangeRateApiUrl is the url of the exchange rate wrapper
	exchangeRateApiUrl = "https://api.exchangerate.host"
	coinGeckoBtcID     = "bitcoin"
	btcSymbol          = "btc"
	lbtcSymbol         = "lbtc"
)

type exchangeRateWrapper struct {
	httpClient              *http.Client
	coinGeckoClient         *coingecko.Client
	assetCurrencySymbolPair map[string]string
}

func NewExchangeRateClient(assetCurrencySymbolPair map[string]string) port.RateService {
	httpClient := &http.Client{
		Timeout: time.Second * httpTimeout,
	}

	client := coingecko.NewClient(httpClient)

	return &exchangeRateWrapper{
		httpClient:              httpClient,
		coinGeckoClient:         client,
		assetCurrencySymbolPair: assetCurrencySymbolPair,
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

	isCryptoSymbol, err := e.isCryptoSymbol(source)
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

func (e *exchangeRateWrapper) isCryptoSymbol(symbol string) (bool, error) {
	symbol = strings.ToLower(symbol)

	list, err := e.coinGeckoClient.CoinsList()
	if err != nil {
		return false, err
	}

	for _, coin := range *list {
		if coin.ID == symbol {
			return true, nil
		}
	}

	return false, err
}

func (e *exchangeRateWrapper) getCryptoToFiatRate(
	ctx context.Context,
	source string,
	target string,
) (decimal.Decimal, error) {
	source = strings.ToLower(source)
	target = strings.ToLower(target)

	price, err := e.coinGeckoClient.SimplePrice(
		[]string{source},
		[]string{target},
	)
	if err != nil {
		return decimal.Zero, err
	}

	fValue := float64((*price)[source][target])
	if fValue == 0 {
		return decimal.Zero, port.ErrCurrencyNotFound
	}

	return decimal.NewFromFloat(fValue), nil
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
