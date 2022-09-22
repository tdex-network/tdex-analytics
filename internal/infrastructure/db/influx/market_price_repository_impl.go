package dbinflux

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"time"
)

func (i *influxDbService) InsertPrice(
	ctx context.Context,
	price domain.MarketPrice,
) error {
	writeAPI := i.client.WriteAPI(i.org, i.analyticsBucket)

	basePriceF, _ := price.BasePrice.BigFloat().Float64()
	quotePriceF, _ := price.QuotePrice.BigFloat().Float64()

	p := influxdb2.NewPointWithMeasurement(MarketPriceTable).
		AddTag(marketTag, price.MarketID).
		AddField(basePrice, basePriceF).
		AddField(quotePrice, quotePriceF).
		SetTime(price.Time)

	writeAPI.WritePoint(p)

	writeAPI.Flush()

	return nil
}

func (i *influxDbService) GetPricesForMarkets(
	ctx context.Context,
	startTime time.Time,
	endTime time.Time,
	page domain.Page,
	groupBy string,
	marketIDs ...string,
) (map[string][]domain.MarketPrice, error) {
	limit := page.Size
	offset := page.Number*page.Size - page.Size
	pagination := fmt.Sprintf("|> limit(n: %v, offset: %v)", limit, offset)
	marketIDsFilter := createMarkedIDsFluxQueryFilter(marketIDs, MarketPriceTable)
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"import \"influxdata/influxdb/schema\" from(bucket:\"%v\")"+
			"|> range(start: %s, stop: %s)"+
			"|> filter(fn: (r) => %v)"+
			"|> filter(fn: (r) => r[\"_field\"] == \"base_price\" or r[\"_field\"] == \"quote_price\")"+
			"|> aggregateWindow(every: %s, fn: mean)"+
			"|> sort() "+
			"|> schema.fieldsAsCols()"+
			"%v",
		i.analyticsBucket,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		marketIDsFilter,
		groupBy,
		pagination,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	response := make(map[string][]domain.MarketPrice)
	for result.Next() {
		marketID := result.Record().ValueByKey(marketTag).(string)
		marketPrice := domain.MarketPrice{
			MarketID:   result.Record().ValueByKey(marketTag).(string),
			BasePrice:  decimal.NewFromFloat(result.Record().ValueByKey(basePrice).(float64)),
			QuotePrice: decimal.NewFromFloat(result.Record().ValueByKey(quotePrice).(float64)),
			Time:       result.Record().Time(),
		}
		val, ok := response[marketID]
		if !ok {
			prices := make([]domain.MarketPrice, 0)
			prices = append(prices, marketPrice)
			response[marketID] = prices
		} else {
			val = append(val, marketPrice)
			response[marketID] = val
		}
	}

	return response, nil
}
