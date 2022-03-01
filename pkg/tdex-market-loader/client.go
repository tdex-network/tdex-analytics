package tdexmarketloader

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/tdex-network/tdex-protobuf/generated/go/types"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
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
}

func NewService(torProxyUrl, registryUrl string) Service {
	return &tdexMarketLoaderService{
		torProxyUrl: torProxyUrl,
		registryUrl: registryUrl,
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

	client := trade.NewTradeClient(conn)
	reply, err := client.MarketPrice(ctx, &trade.MarketPriceRequest{
		Market: &types.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		Type:   trade.TradeType_SELL,
		Amount: 1000,
		Asset:  market.BaseAsset,
	})
	if err != nil {
		return nil, err
	}

	return &Balance{
		BaseBalance:  int(reply.GetPrices()[0].GetBalance().GetBaseAmount()), //TODO to be update with new spec
		QuoteBalance: int(reply.GetPrices()[0].GetBalance().GetQuoteAmount()),
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

	client := trade.NewTradeClient(conn)
	reply, err := client.MarketPrice(ctx, &trade.MarketPriceRequest{
		Market: &types.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		Type:   trade.TradeType_SELL,
		Amount: 1000,
		Asset:  market.BaseAsset,
	})
	if err != nil {
		return nil, err
	}

	return &Price{
		BasePrice:  int(reply.GetPrices()[0].GetPrice().GetBasePrice()), //TODO to be update with new spec
		QuotePrice: int(reply.GetPrices()[0].GetPrice().GetQuotePrice()),
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

	reg := &Registry{}
	if err = json.Unmarshal(body, reg); err != nil {
		return nil, err
	}

	resp, err = http.Get(reg.DownloadUrl)
	if err != nil {
		return nil, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	client := trade.NewTradeClient(conn)
	reply, err := client.Markets(ctx, &trade.MarketsRequest{})
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

	if strings.Contains(endpoint, onionUrlRegex) {
		c, err := getGrpcConnectionWithTorClient(
			context.Background(),
			strings.ReplaceAll(endpoint, httpRegex, ""),
			t.torProxyUrl,
		)
		if err != nil {
			return nil, nil, err
		}

		conn = c
	} else {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		c, err := grpc.DialContext(
			context.Background(),
			strings.ReplaceAll(endpoint, httpsRegex, ""),
			grpc.WithTransportCredentials(credentials.NewTLS(config)), //TODO if provider doesnt uses TLS this will fail
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

type Registry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	Url         string `json:"url"`
	HtmlUrl     string `json:"html_url"`
	GitUrl      string `json:"git_url"`
	DownloadUrl string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		Html string `json:"html"`
	} `json:"_links"`
}

type LiquidityProvider struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Markets  []Market
}

func getGrpcConnectionWithTorClient(
	ctx context.Context,
	onionUrl, torProxyUrl string,
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
		grpc.WithTransportCredentials(insecure.NewCredentials()), //TODO if provider uses TLS this will fail
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
	BasePrice  int
	QuotePrice int
}
