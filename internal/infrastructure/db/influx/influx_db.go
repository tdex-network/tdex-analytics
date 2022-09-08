package dbinflux

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
)

const (
	marketTag          = "market_id"
	MarketPriceTable   = "market_price"
	MarketBalanceTable = "market_balance"
	baseAsset          = "base_asset"
	baseBalance        = "base_balance"
	basePrice          = "base_price"
	quoteAsset         = "quote_asset"
	quoteBalance       = "quote_balances"
	quotePrice         = "quote_price"
)

type Config struct {
	Org             string
	AuthToken       string
	DbUrl           string
	AnalyticsBucket string
}

type Service interface {
	domain.MarketBalanceRepository
	domain.MarketPriceRepository
	Close()
}

type influxDbService struct {
	org             string
	analyticsBucket string
	client          influxdb2.Client
}

func New(config Config) (Service, error) {
	//TODO check if there should be one writeApi and if it should be sync or async
	client := influxdb2.NewClient(config.DbUrl, config.AuthToken)

	return &influxDbService{
		org:             config.Org,
		analyticsBucket: config.AnalyticsBucket,
		client:          client,
	}, nil
}

func (i *influxDbService) Close() {
	i.client.Close()
}
