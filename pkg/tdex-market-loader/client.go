package tdexmarketloader

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	tdexv1 "github.com/tdex-network/tdex-analytics/api-spec/protobuf/gen/tdex/v1"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const (
	onionUrlRegex = "onion"
	httpRegex     = "http://"
	httpsRegex    = "https://"

	listMarketsUrlRegex             = "%s/v1/markets"
	fetchMarketBalanceUrlRegex      = "%s/v1/market/balance"
	fetchMarketPriceUrlRegex        = "%s/v1/market/price"
	fetchMarketTradePreviewUrlRegex = "%s/v1/trade/preview"
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
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unavailable {
				requestData, err := json.Marshal(MarketReq{
					MarketInfo: struct {
						BaseAsset  string `json:"baseAsset"`
						QuoteAsset string `json:"quoteAsset"`
					}{
						BaseAsset:  market.BaseAsset,
						QuoteAsset: market.QuoteAsset,
					},
				})
				if err != nil {
					return nil, err
				}

				r, err := http1Req(
					fmt.Sprintf(fetchMarketBalanceUrlRegex, market.Url),
					t.torProxyUrl, "POST",
					requestData,
				)
				if err != nil {
					return nil, err
				}

				fetchBalanceResponse := FetchBalanceResp{}
				if err := json.Unmarshal(r, &fetchBalanceResponse); err != nil {
					return nil, err
				}

				baseAmount, err := decimal.NewFromString(fetchBalanceResponse.BalanceInfo.Balance.BaseAmount)
				if err != nil {
					return nil, err
				}

				quoteAmount, err := decimal.NewFromString(fetchBalanceResponse.BalanceInfo.Balance.QuoteAmount)
				if err != nil {
					return nil, err
				}

				return &Balance{
					BaseBalance:  baseAmount,
					QuoteBalance: quoteAmount,
				}, nil
			}
		}
		return nil, err
	}

	return &Balance{
		BaseBalance:  decimal.NewFromInt(int64(reply.GetBalance().GetBalance().GetBaseAmount())),
		QuoteBalance: decimal.NewFromInt(int64(reply.GetBalance().GetBalance().GetQuoteAmount())),
	}, nil
}

func (t *tdexMarketLoaderService) FetchPrice(
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
	reply, err := client.GetMarketPrice(ctx, &tdexv1.GetMarketPriceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
	})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unavailable {
				requestData, err := json.Marshal(MarketReq{
					MarketInfo: struct {
						BaseAsset  string `json:"baseAsset"`
						QuoteAsset string `json:"quoteAsset"`
					}{
						BaseAsset:  market.BaseAsset,
						QuoteAsset: market.QuoteAsset,
					},
				})
				if err != nil {
					return nil, err
				}

				r, err := http1Req(
					fmt.Sprintf(fetchMarketPriceUrlRegex, market.Url),
					t.torProxyUrl,
					"POST",
					requestData,
				)
				if err != nil {
					r, err = http1Req(
						fmt.Sprintf(fetchMarketTradePreviewUrlRegex, market.Url),
						t.torProxyUrl,
						"POST",
						requestData,
					)
					if err != nil {
						return nil, err
					}

					fetchMarketTradePreview := FetchMarketTradePreviewResp{}
					if err := json.Unmarshal(r, &fetchMarketTradePreview); err != nil {
						return nil, err
					}

					basePrices := make([]decimal.Decimal, 0)
					quotePrices := make([]decimal.Decimal, 0)
					for _, v := range fetchMarketTradePreview.Previews {
						basePrices = append(basePrices, decimal.NewFromFloat(v.PriceInfo.BasePrice))
						quotePrices = append(quotePrices, decimal.NewFromInt(int64(v.PriceInfo.QuotePrice)))
					}

					bp := decimal.Avg(basePrices[0], basePrices[1:]...).Round(8)
					qp := decimal.Avg(quotePrices[0], quotePrices[1:]...).Round(8)

					return &Price{
						BasePrice:  bp,
						QuotePrice: qp,
					}, nil
				}

				fetchMarketPriceResponse := FetchMarketPriceResp{}
				if err := json.Unmarshal(r, &fetchMarketPriceResponse); err != nil {
					return nil, err
				}

				minAmt, err := decimal.NewFromString(fetchMarketPriceResponse.MinTradableAmount)
				if err := json.Unmarshal(r, &fetchMarketPriceResponse); err != nil {
					return nil, err
				}

				qp := decimal.NewFromFloat(fetchMarketPriceResponse.SpotPrice)
				bp := decimal.NewFromFloat(1).Div(minAmt)

				return &Price{
					BasePrice:  bp,
					QuotePrice: qp,
				}, nil
			}
		}

		bp, qp, err := t.previewPrice(ctx, client, market)
		if err != nil {
			return nil, err
		}

		basePrice = bp
		quotePrice = qp
	} else {
		quotePrice = decimal.NewFromFloat(reply.GetSpotPrice())
		basePrice = decimal.NewFromFloat(1).Div(quotePrice)
	}

	return &Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}, nil
}

func (t *tdexMarketLoaderService) previewPrice(
	ctx context.Context,
	client tdexv1.TradeServiceClient,
	market Market,
) (decimal.Decimal, decimal.Decimal, error) {
	var (
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	)
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
		return decimal.Zero, decimal.Zero, err
	}

	basePrices := make([]decimal.Decimal, 0)
	quotePrices := make([]decimal.Decimal, 0)
	for _, v := range reply.GetPreviews() {
		basePrices = append(basePrices, decimal.NewFromFloat(v.GetPrice().GetBasePrice()))
		quotePrices = append(quotePrices, decimal.NewFromFloat(v.GetPrice().GetQuotePrice()))
	}

	basePrice = decimal.Avg(basePrices[0], basePrices[1:]...).Round(8)
	quotePrice = decimal.Avg(quotePrices[0], quotePrices[1:]...).Round(8)

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
	resp := make([]Market, 0)

	conn, close, err := t.getConn(liquidityProvider.Endpoint)
	if err != nil {
		return nil, err
	}
	defer close()

	client := tdexv1.NewTradeServiceClient(conn)
	reply, err := client.ListMarkets(ctx, &tdexv1.ListMarketsRequest{})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unavailable {
				requestData := []byte("{}")
				r, err := http1Req(fmt.Sprintf(
					listMarketsUrlRegex,
					liquidityProvider.Endpoint),
					t.torProxyUrl,
					"POST",
					requestData,
				)
				if err != nil {
					return nil, err
				}

				listMarketsResponse := ListMarketsResp{}
				if err := json.Unmarshal(r, &listMarketsResponse); err != nil {
					return nil, err
				}

				for _, v := range listMarketsResponse.Markets {
					resp = append(resp, Market{
						QuoteAsset: v.MarketInfo.QuoteAsset,
						BaseAsset:  v.MarketInfo.BaseAsset,
					})
				}

				return resp, nil
			}
		}
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
