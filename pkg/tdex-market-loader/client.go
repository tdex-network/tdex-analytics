package tdexmarketloader

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	tdexv1 "tdex-analytics/api-spec/protobuf/gen/tdex/v1"
)

const (
	onionUrlRegex = "onion"
	httpRegex     = "http://"
	httpsRegex    = "https://"
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
	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	reply, err := client.GetMarketBalance(ctx, &tdexv1.GetMarketBalanceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	})
	if err != nil {
		return nil, err
	}

	return &Balance{
		BaseBalance:  int(reply.GetBalance().GetBalance().GetBaseAmount()),
		QuoteBalance: int(reply.GetBalance().GetBalance().GetQuoteAmount()),
	}, nil
}

func (t *tdexMarketLoaderService) FetchPrice(
	ctx context.Context,
	market Market,
) (*Price, error) {
	conn, close, err := t.getConn(market.Url)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	reply, err := client.PreviewTrade(ctx, &tdexv1.PreviewTradeRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		Type:   tdexv1.TradeType_TRADE_TYPE_SELL,
		Amount: uint64(t.priceAmount),
		Asset:  market.BaseAsset,
	})
	if err != nil {
		return nil, err
	}

	basePrices := make([]decimal.Decimal, 0)
	quotePrices := make([]decimal.Decimal, 0)
	for _, v := range reply.GetPreviews() {
		basePrices = append(basePrices, decimal.NewFromFloat(v.GetPrice().GetBasePrice()))
		quotePrices = append(quotePrices, decimal.NewFromFloat(v.GetPrice().GetQuotePrice()))
	}

	basePriceAvg := decimal.Avg(basePrices[0], basePrices[1:]...).Round(8)
	quotePriceAvg := decimal.Avg(quotePrices[0], quotePrices[1:]...).Round(8)

	return &Price{
		BasePrice:  basePriceAvg,
		QuotePrice: quotePriceAvg,
	}, nil
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
	resp := make([]Market, 0)

	conn, close, err := t.getConn(liquidityProvider.Endpoint)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	reply, err := client.ListMarkets(ctx, &tdexv1.ListMarketsRequest{})
	if err != nil {
		return nil, err
	}
	for _, v := range reply.GetMarkets() {
		resp = append(resp, Market{
			QuoteAsset: v.GetMarket().GetQuoteAsset(),
			BaseAsset:  v.GetMarket().GetBaseAsset(),
		})
	}

	return resp, nil
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

type Balance struct {
	BaseBalance  int
	QuoteBalance int
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}
