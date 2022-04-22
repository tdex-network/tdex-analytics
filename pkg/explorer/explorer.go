package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tdex-analytics/internal/core/port"
	"time"
)

const (
	httpTimeout = 10
)

type explorerService struct {
	explorerUrl string
	httpClient  *http.Client
}

func NewExplorerService(explorerUrl string) port.ExplorerService {
	return &explorerService{
		explorerUrl: explorerUrl,
		httpClient: &http.Client{
			Timeout: time.Second * httpTimeout,
		},
	}
}

func (e explorerService) GetAssetCurrency(
	ctx context.Context,
	assetId string,
) (string, error) {
	url := fmt.Sprintf("%s/asset/%s", e.explorerUrl, assetId)

	resp, err := e.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf(
			"unexpected status code: %d, error: %v",
			resp.StatusCode,
			resp.Body,
		)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Ticker, nil
}

type Response struct {
	AssetId      string `json:"asset_id"`
	IssuanceTxin struct {
		Txid string `json:"txid"`
		Vin  int    `json:"vin"`
	} `json:"issuance_txin"`
	IssuancePrevout struct {
		Txid string `json:"txid"`
		Vout int    `json:"vout"`
	} `json:"issuance_prevout"`
	ReissuanceToken string `json:"reissuance_token"`
	ContractHash    string `json:"contract_hash"`
	Status          struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int    `json:"block_time"`
	} `json:"status"`
	ChainStats struct {
		TxCount                int  `json:"tx_count"`
		IssuanceCount          int  `json:"issuance_count"`
		IssuedAmount           int  `json:"issued_amount"`
		BurnedAmount           int  `json:"burned_amount"`
		HasBlindedIssuances    bool `json:"has_blinded_issuances"`
		ReissuanceTokens       int  `json:"reissuance_tokens"`
		BurnedReissuanceTokens int  `json:"burned_reissuance_tokens"`
	} `json:"chain_stats"`
	MempoolStats struct {
		TxCount                int         `json:"tx_count"`
		IssuanceCount          int         `json:"issuance_count"`
		IssuedAmount           int         `json:"issued_amount"`
		BurnedAmount           int         `json:"burned_amount"`
		HasBlindedIssuances    bool        `json:"has_blinded_issuances"`
		ReissuanceTokens       interface{} `json:"reissuance_tokens"`
		BurnedReissuanceTokens int         `json:"burned_reissuance_tokens"`
	} `json:"mempool_stats"`
	Contract struct {
		Entity struct {
			Domain string `json:"domain"`
		} `json:"entity"`
		IssuerPubkey string `json:"issuer_pubkey"`
		Name         string `json:"name"`
		Nft          struct {
			Domain string `json:"domain"`
			Hash   string `json:"hash"`
		} `json:"nft"`
		Precision int    `json:"precision"`
		Ticker    string `json:"ticker"`
		Version   int    `json:"version"`
	} `json:"contract"`
	Entity struct {
		Domain string `json:"domain"`
	} `json:"entity"`
	Precision int    `json:"precision"`
	Name      string `json:"name"`
	Ticker    string `json:"ticker"`
}
