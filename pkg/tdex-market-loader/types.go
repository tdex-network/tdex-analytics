package tdexmarketloader

import "github.com/shopspring/decimal"

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

type Balance struct {
	BaseBalance  decimal.Decimal
	QuoteBalance decimal.Decimal
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

type MarketReq struct {
	MarketInfo struct {
		BaseAsset  string `json:"baseAsset"`
		QuoteAsset string `json:"quoteAsset"`
	} `json:"market"`
}

type MarketPreviewReq struct {
	MarketInfo struct {
		BaseAsset  string `json:"baseAsset"`
		QuoteAsset string `json:"quoteAsset"`
	} `json:"market"`
	Type   int    `json:"type"`
	Amount int    `json:"amount"`
	Asset  string `json:"asset"`
}

type ListMarketsResp struct {
	Markets []struct {
		MarketInfo struct {
			BaseAsset  string `json:"baseAsset"`
			QuoteAsset string `json:"quoteAsset"`
		} `json:"market"`
		FeeInfo struct {
			BasisPoint string `json:"basisPoint"`
			Fixed      struct {
				BaseFee  string `json:"baseFee"`
				QuoteFee string `json:"quoteFee"`
			} `json:"fixed"`
		} `json:"fee"`
	} `json:"markets"`
}

type FetchBalanceResp struct {
	BalanceInfo struct {
		Balance struct {
			BaseAmount  string `json:"baseAmount"`
			QuoteAmount string `json:"quoteAmount"`
		} `json:"balance"`
		Fee struct {
			BasisPoint string `json:"basisPoint"`
			Fixed      struct {
				BaseFee  string `json:"baseFee"`
				QuoteFee string `json:"quoteFee"`
			} `json:"fixed"`
		} `json:"fee"`
	} `json:"balance"`
}

type FetchMarketPriceResp struct {
	SpotPrice         float64 `json:"spotPrice"`
	MinTradableAmount string  `json:"minTradableAmount"`
}

type FetchMarketTradePreviewResp struct {
	Previews []struct {
		PriceInfo struct {
			BasePrice  float64 `json:"basePrice"`
			QuotePrice int     `json:"quotePrice"`
		} `json:"price"`
		FeeInfo struct {
			BasisPoint string `json:"basisPoint"`
			Fixed      struct {
				BaseFee  string `json:"baseFee"`
				QuoteFee string `json:"quoteFee"`
			} `json:"fixed"`
		} `json:"fee"`
		Amount      string `json:"amount"`
		Asset       string `json:"asset"`
		BalanceInfo struct {
			BaseAmount  string `json:"baseAmount"`
			QuoteAmount string `json:"quoteAmount"`
		} `json:"balance"`
	} `json:"previews"`
}
