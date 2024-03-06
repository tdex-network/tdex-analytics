package tdexmarketloader

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	tdexv1 "github.com/tdex-network/tdex-analytics/api-spec/protobuf/gen/tdex/v1"
	tdexv2 "github.com/tdex-network/tdex-analytics/api-spec/protobuf/gen/tdex/v2"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	onionUrlRegex = "onion"
	httpRegex     = "http://"
	httpsRegex    = "https://"

	listMarketsUrlRegexV2             = "%s/v2/markets"
	fetchMarketBalanceUrlRegexV2      = "%s/v2/market/balance"
	fetchMarketPriceUrlRegexV2        = "%s/v2/market/price"
	fetchMarketTradePreviewUrlRegexV2 = "%s/v2/trade/preview"
	listMarketsUrlRegexV1             = "%s/v1/markets"
	fetchMarketBalanceUrlRegexV1      = "%s/v1/market/balance"
	fetchMarketPriceUrlRegexV1        = "%s/v1/market/price"
	fetchMarketTradePreviewUrlRegexV1 = "%s/v1/trade/preview"
)

type Service interface {
	FetchProvidersMarkets(ctx context.Context) ([]LiquidityProvider, error)
	FetchBalance(ctx context.Context, market Market) (*Balance, error)
	FetchPrice(ctx context.Context, market Market) (*Price, error)
}

type tdexMarketLoaderService struct {
	torProxyUrl string
	registryUrl string
	priceAmount int
}

func NewService(torProxyUrl, registryUrl string, priceAmount int) Service {
	return &tdexMarketLoaderService{
		torProxyUrl: torProxyUrl,
		registryUrl: registryUrl,
		priceAmount: priceAmount,
	}
}

func (t *tdexMarketLoaderService) FetchProvidersMarkets(
	ctx context.Context,
) ([]LiquidityProvider, error) {
	res := make([]LiquidityProvider, 0)
	liquidityProviders, err := t.fetchLiquidityProviders()
	if err != nil {
		return nil, err
	}

	for _, v := range liquidityProviders {
		markets, err := t.fetchLiquidityProviderMarkets(ctx, v)
		if err != nil {
			log.Errorf(
				"error while trying to fetch markets for liquidity provider: %v, err: %v\n",
				v.Name,
				err.Error(),
			)

			continue
		}
		res = append(res, LiquidityProvider{
			Name:     v.Name,
			Endpoint: v.Endpoint,
			Markets:  markets,
		})
	}

	return res, nil
}

func (t *tdexMarketLoaderService) FetchBalance(
	ctx context.Context,
	market Market,
) (*Balance, error) {
	balance, err := t.getBalanceV2(ctx, market)
	if err == nil {
		return balance, nil
	}

	return t.getBalanceV1(ctx, market)
}

func (t *tdexMarketLoaderService) FetchPrice(
	ctx context.Context,
	market Market,
) (*Price, error) {
	price, err := t.getPriceV2(ctx, market)
	if err == nil {
		return price, nil
	}
	return t.getPriceV1(ctx, market)
}

func (t *tdexMarketLoaderService) previewPrice(
	ctx context.Context,
	client tdexv1.TradeServiceClient,
	market Market,
) (decimal.Decimal, decimal.Decimal, error) {
	req := &tdexv1.PreviewTradeRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		Type:   tdexv1.TradeType_TRADE_TYPE_SELL,
		Amount: uint64(t.priceAmount),
		Asset:  market.BaseAsset,
	}
	// Try HTTP/2 endpoint.
	reply, err := client.PreviewTrade(ctx, req)
	if err != nil {
		requestData, _ := protojson.Marshal(req)
		// Fallback to HTTP/1 endpoint.
		r, err := http1Req(
			fmt.Sprintf(fetchMarketTradePreviewUrlRegexV1, market.Url),
			t.torProxyUrl,
			"POST",
			requestData,
		)
		if err != nil {
			return decimal.Zero, decimal.Zero, err
		}

		reply = &tdexv1.PreviewTradeResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return decimal.Zero, decimal.Zero, err
		}
	}

	basePrices := make([]decimal.Decimal, 0)
	quotePrices := make([]decimal.Decimal, 0)
	for _, v := range reply.GetPreviews() {
		basePrices = append(basePrices, decimal.NewFromFloat(v.GetPrice().GetBasePrice()))
		quotePrices = append(quotePrices, decimal.NewFromFloat(v.GetPrice().GetQuotePrice()))
	}

	basePrice := decimal.Avg(basePrices[0], basePrices[1:]...).Round(8)
	quotePrice := decimal.Avg(quotePrices[0], quotePrices[1:]...).Round(8)
	return basePrice, quotePrice, nil
}

func (t *tdexMarketLoaderService) fetchLiquidityProviders() ([]LiquidityProvider, error) {
	resp, err := http.Get(t.registryUrl)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %v, err: %v", resp.StatusCode, string(body))
	}

	var liquidityProviders []LiquidityProvider
	if err = json.Unmarshal(body, &liquidityProviders); err != nil {
		return nil, err
	}

	return liquidityProviders, nil
}

func (t *tdexMarketLoaderService) fetchLiquidityProviderMarkets(
	ctx context.Context,
	liquidityProvider LiquidityProvider,
) ([]Market, error) {
	markets, err := t.getMarketsV2(ctx, liquidityProvider)
	if err == nil {
		return markets, nil
	}

	return t.getMarketsV1(ctx, liquidityProvider)
}

func (t *tdexMarketLoaderService) getConn(endpoint string) (*grpc.ClientConn, func(), error) {
	var conn *grpc.ClientConn

	url := strings.ReplaceAll(endpoint, httpRegex, "")
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	if strings.Contains(endpoint, httpsRegex) { //if https skip TLS cert, it can be trusted
		url = strings.ReplaceAll(endpoint, httpsRegex, "")
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		creds = grpc.WithTransportCredentials(credentials.NewTLS(config))
	}

	if strings.Contains(endpoint, onionUrlRegex) {
		c, err := getGrpcConnectionWithTorClient(
			context.Background(),
			url,
			t.torProxyUrl,
			creds,
		)
		if err != nil {
			return nil, nil, err
		}

		conn = c
	} else {
		c, err := grpc.DialContext(
			context.Background(),
			url,
			creds,
		)
		if err != nil {
			return nil, nil, err
		}

		conn = c
	}

	cleanup := func() { conn.Close() }

	return conn, cleanup, nil
}

func (t *tdexMarketLoaderService) getMarketsV2(
	ctx context.Context,
	liquidityProvider LiquidityProvider,
) ([]Market, error) {
	conn, close, err := t.getConn(liquidityProvider.Endpoint)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv2.NewTradeServiceClient(conn)
	req := &tdexv2.ListMarketsRequest{}
	// Try HTTP/2 endpoint.
	reply, err := client.ListMarkets(ctx, req)
	if err != nil {
		requestData := []byte("{}")
		// Fallback to HTTP/1 endpoint.
		r, err := http1Req(fmt.Sprintf(
			listMarketsUrlRegexV2,
			liquidityProvider.Endpoint),
			t.torProxyUrl,
			"POST",
			requestData,
		)
		if err != nil {
			return nil, err
		}

		reply = &tdexv2.ListMarketsResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}

	resp := make([]Market, 0, len(reply.GetMarkets()))
	for _, v := range reply.GetMarkets() {
		resp = append(resp, Market{
			QuoteAsset: v.GetMarket().GetQuoteAsset(),
			BaseAsset:  v.GetMarket().GetBaseAsset(),
		})
	}

	return resp, nil
}

func (t *tdexMarketLoaderService) getMarketsV1(
	ctx context.Context,
	liquidityProvider LiquidityProvider,
) ([]Market, error) {

	conn, close, err := t.getConn(liquidityProvider.Endpoint)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	req := &tdexv1.ListMarketsRequest{}
	// Try HTTP/2 endpoint.
	reply, err := client.ListMarkets(ctx, req)
	if err != nil {
		requestData := []byte("{}")
		// Fallback to HTTP/1 endpoint.
		r, err := http1Req(fmt.Sprintf(
			listMarketsUrlRegexV1,
			liquidityProvider.Endpoint),
			t.torProxyUrl,
			"POST",
			requestData,
		)
		if err != nil {
			return nil, err
		}

		reply = &tdexv1.ListMarketsResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}

	resp := make([]Market, 0, len(reply.GetMarkets()))
	for _, v := range reply.GetMarkets() {
		resp = append(resp, Market{
			QuoteAsset: v.GetMarket().GetQuoteAsset(),
			BaseAsset:  v.GetMarket().GetBaseAsset(),
		})
	}

	return resp, nil
}

func (t *tdexMarketLoaderService) getBalanceV2(
	ctx context.Context,
	market Market,
) (*Balance, error) {
	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv2.NewTradeServiceClient(conn)
	req := &tdexv2.GetMarketBalanceRequest{
		Market: &tdexv2.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	}

	// Try HTTP/2 endpoint.
	reply, err := client.GetMarketBalance(ctx, req)
	if err != nil {
		requestData, _ := protojson.Marshal(req)
		// Fallback to HTTP/1 endpoint.
		r, err := http1Req(
			fmt.Sprintf(fetchMarketBalanceUrlRegexV2, market.Url),
			t.torProxyUrl, "POST",
			requestData,
		)
		if err != nil {
			return nil, err
		}

		reply = &tdexv2.GetMarketBalanceResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}
	return &Balance{
		BaseBalance:  decimal.NewFromInt(int64(reply.GetBalance().GetBaseAmount())),
		QuoteBalance: decimal.NewFromInt(int64(reply.GetBalance().GetQuoteAmount())),
	}, nil
}

func (t *tdexMarketLoaderService) getBalanceV1(
	ctx context.Context,
	market Market,
) (*Balance, error) {
	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	req := &tdexv1.GetMarketBalanceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	}

	// Try HTTP/2 endpoint.
	reply, err := client.GetMarketBalance(ctx, req)
	if err != nil {
		requestData, _ := protojson.Marshal(req)
		// Fallback to HTTP/1 endpoint.
		r, err := http1Req(
			fmt.Sprintf(fetchMarketBalanceUrlRegexV1, market.Url),
			t.torProxyUrl, "POST",
			requestData,
		)
		if err != nil {
			return nil, err
		}

		reply = &tdexv1.GetMarketBalanceResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}
	return &Balance{
		BaseBalance:  decimal.NewFromInt(int64(reply.GetBalance().GetBalance().GetBaseAmount())),
		QuoteBalance: decimal.NewFromInt(int64(reply.GetBalance().GetBalance().GetQuoteAmount())),
	}, nil
}

func (t *tdexMarketLoaderService) getPriceV2(
	ctx context.Context,
	market Market,
) (*Price, error) {
	var (
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	)

	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv2.NewTradeServiceClient(conn)
	req := &tdexv2.GetMarketPriceRequest{
		Market: &tdexv2.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	}
	// Try HTTP/2 market price endpoint.
	reply, err := client.GetMarketPrice(ctx, req)
	if err != nil {
		requestData, _ := protojson.Marshal(req)
		// Fallback to HTTP/1 market price endpoint.
		r, err := http1Req(
			fmt.Sprintf(fetchMarketPriceUrlRegexV2, market.Url),
			t.torProxyUrl,
			"POST",
			requestData,
		)

		if err != nil {
			return nil, err
		}

		reply = &tdexv2.GetMarketPriceResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}
	quotePrice = decimal.NewFromFloat(reply.GetSpotPrice())
	basePrice = decimal.NewFromFloat(1).Div(quotePrice)

	return &Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}, nil
}

func (t *tdexMarketLoaderService) getPriceV1(
	ctx context.Context,
	market Market,
) (*Price, error) {
	var (
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	)

	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	req := &tdexv1.GetMarketPriceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	}
	// Try HTTP/2 market price endpoint.
	reply, err := client.GetMarketPrice(ctx, req)
	if err != nil {
		requestData, _ := protojson.Marshal(req)
		// Fallback to HTTP/1 market price endpoint.
		r, err := http1Req(
			fmt.Sprintf(fetchMarketPriceUrlRegexV1, market.Url),
			t.torProxyUrl,
			"POST",
			requestData,
		)
		if err != nil {
			// Fallback to trade preview endpoint.
			bp, qp, err := t.previewPrice(ctx, client, market)
			if err != nil {
				return nil, err
			}

			return &Price{
				BasePrice:  bp,
				QuotePrice: qp,
			}, nil
		}

		reply = &tdexv1.GetMarketPriceResponse{}
		if err := protojson.Unmarshal(r, reply); err != nil {
			return nil, err
		}
	}
	quotePrice = decimal.NewFromFloat(reply.GetSpotPrice())
	basePrice = decimal.NewFromFloat(1).Div(quotePrice)

	return &Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}, nil
}

func getGrpcConnectionWithTorClient(
	ctx context.Context,
	onionUrl, torProxyUrl string,
	creds grpc.DialOption,
) (*grpc.ClientConn, error) {
	result := make(chan interface{}, 1)

	writeResult := func(res interface{}) {
		select {
		case result <- res:
		default:
		}
	}

	dialer := func(ctx context.Context, address string) (net.Conn, error) {
		torDialer, err := proxy.SOCKS5("tcp", torProxyUrl, nil, proxy.Direct)
		if err != nil {
			writeResult(err)
		}

		conn, err := torDialer.Dial("tcp", address)
		if err != nil {
			writeResult(err)
		}
		return conn, err
	}

	opts := []grpc.DialOption{
		creds,
		grpc.WithContextDialer(dialer),
	}

	go func(ou string, o ...grpc.DialOption) {
		conn, err := grpc.DialContext(context.Background(), ou, o...)
		if err != nil {
			writeResult(err)
		} else {
			writeResult(conn)
		}
	}(onionUrl, opts...)

	select {
	case res := <-result:
		if conn, ok := res.(*grpc.ClientConn); ok {
			return conn, nil
		}
		return nil, res.(error)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func http1Req(url string, socksUrl string, method string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpTransport := &http.Transport{}
	if strings.Contains(url, onionUrlRegex) {
		socksDialer, err := proxy.SOCKS5("tcp", socksUrl, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		httpTransport = &http.Transport{
			DialContext: socksDialer.(proxy.ContextDialer).DialContext,
		}
	}

	client := &http.Client{Transport: httpTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"unexpected status code: %d, error: %v",
			resp.StatusCode,
			string(body),
		)
	}

	return body, nil
}
