package rater

import (
	coingecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"net/http"
)

type CoinGeckoService interface {
	CoinsList() (*types.CoinList, error)
	SimplePrice(ids []string, vsCurrencies []string) (*map[string]map[string]float32, error)
}

type coinGeckoService struct {
	coinGeckoClient *coingecko.Client
}

func NewCoinGeckoService(httpClient *http.Client) CoinGeckoService {
	client := coingecko.NewClient(httpClient)

	return &coinGeckoService{
		coinGeckoClient: client,
	}
}

func (c *coinGeckoService) CoinsList() (*types.CoinList, error) {
	return c.coinGeckoClient.CoinsList()
}

func (c *coinGeckoService) SimplePrice(ids []string, vsCurrencies []string) (*map[string]map[string]float32, error) {
	return c.coinGeckoClient.SimplePrice(ids, vsCurrencies)
}
