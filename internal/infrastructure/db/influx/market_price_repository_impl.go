package dbinflux

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"strings"
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
	if groupBy != "" {
		groupBy = fmt.Sprintf("|> aggregateWindow(every: %s, fn: mean)", groupBy)
	}
	query := fmt.Sprintf(
		"import \"influxdata/influxdb/schema\" from(bucket:\"%s\")"+
			"|> range(start: %s, stop: %s)"+
			"|> filter(fn: (r) => %s)"+
			"%v"+
			"|> schema.fieldsAsCols()"+
			"%s"+
			"|> sort()",
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
		bPrice := decimal.NewFromInt(0)
		if result.Record().ValueByKey(basePrice) != nil {
			bPrice = decimal.NewFromFloat(result.Record().ValueByKey(basePrice).(float64))
		}
		qPrice := decimal.NewFromInt(0)
		if result.Record().ValueByKey(quotePrice) != nil {
			qPrice = decimal.NewFromFloat(result.Record().ValueByKey(quotePrice).(float64))
		}

		marketPrice := domain.MarketPrice{
			MarketID:   marketID,
			BasePrice:  bPrice,
			QuotePrice: qPrice,
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

// CalculateVWAP calculates the Volume Weighted Average Price (VWAP) for the given market IDs within the specified time range.
func (i *influxDbService) CalculateVWAP(
	ctx context.Context,
	aggregationWindow string,
	startTime time.Time,
	endTime time.Time,
	marketIDs ...string,
) (decimal.Decimal, error) {
	marketIdsFiler := fmt.Sprintf(
		`["%s"]`, strings.Join(marketIDs, `","`),
	)

	// The FluxDB query below calculates the sum of products of price and
	//balance for each market and each time window
	priceBalanceProductSumTemplate := `
	market_ids = %s

	balanceStream = from(bucket: "analytics")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) =>
		r._measurement == "market_balance" and
		r._field == "quote_balance" and
		contains(value: r.market_id, set: market_ids)
	)
	|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
	|> sort()

	priceStream = from(bucket: "analytics")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) =>
		r._measurement == "market_price" and
		r._field == "base_price" and
		contains(value: r.market_id, set: market_ids)
	)
	|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
	|> sort()

	join(tables: {balance: balanceStream, price: priceStream}, on: ["market_id", "_time"])
	|> map(fn: (r) => ({
		_time: r._time,
		market_id: r.market_id,
		VWAP: r._value_price * r._value_balance
	})
	)
	|> group()
	|> sum(column: "VWAP")`

	priceBalanceProductSumQuery := fmt.Sprintf(
		priceBalanceProductSumTemplate,
		marketIdsFiler,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		aggregationWindow,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		aggregationWindow,
	)

	// The FluxDB query below calculates the total balance for each market
	balanceSumTemplate := `
	market_ids = %s

	totalBalanceStream = from(bucket: "analytics")
    |> range(start: %s, stop: %s)
    |> filter(fn: (r) =>
    	r._measurement == "market_balance" and
    	r._field == "base_balance" and
    	contains(value: r.market_id, set: market_ids)
  	)
  	|> group()
	|> sum()
	|> yield()`

	balanceSumQuery := fmt.Sprintf(
		balanceSumTemplate,
		marketIdsFiler,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	priceBalanceProductSumResult, err := i.client.QueryAPI(i.org).Query(
		ctx,
		priceBalanceProductSumQuery,
	)
	if err != nil {
		return decimal.Zero, err
	}

	balanceSumResult, err := i.client.QueryAPI(i.org).Query(
		ctx,
		balanceSumQuery,
	)
	if err != nil {
		return decimal.Zero, err
	}

	var vwapSum decimal.Decimal
	var balanceSum decimal.Decimal

	for priceBalanceProductSumResult.Next() {
		vwapSum = decimal.NewFromFloat(priceBalanceProductSumResult.Record().ValueByKey("VWAP").(float64))
	}
	for balanceSumResult.Next() {
		balanceSum = decimal.NewFromFloat(balanceSumResult.Record().Value().(float64))
	}

	if vwapSum.IsZero() || balanceSum.IsZero() {
		return decimal.Zero, nil
	}

	return vwapSum.Div(balanceSum), nil
}
