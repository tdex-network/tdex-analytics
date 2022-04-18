package dbinflux

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"tdex-analytics/internal/core/domain"
	"time"
)

func (i *influxDbService) InsertBalance(
	ctx context.Context,
	balance domain.MarketBalance,
) error {
	writeAPI := i.client.WriteAPI(i.org, i.analyticsBucket)

	p := influxdb2.NewPointWithMeasurement(MarketBalanceTable).
		AddTag(marketTag, balance.MarketID).
		AddField(baseAsset, balance.BaseAsset).
		AddField(baseBalance, balance.BaseBalance).
		AddField(quoteAsset, balance.QuoteAsset).
		AddField(quoteBalance, balance.QuoteBalance).
		SetTime(balance.Time)

	writeAPI.WritePoint(p)

	writeAPI.Flush()

	return nil
}

func (i *influxDbService) GetBalancesForMarkets(
	ctx context.Context,
	startTime time.Time,
	endTime time.Time,
	page domain.Page,
	marketIDs ...string,
) (map[string][]domain.MarketBalance, error) {
	limit := page.Size
	offset := page.Number*page.Size - page.Size
	pagination := fmt.Sprintf("|> limit(n: %v, offset: %v)", limit, offset)
	marketIDsFilter := createMarkedIDsFluxQueryFilter(marketIDs)
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"import \"influxdata/influxdb/schema\" from(bucket:\"%v\")|> range(start: %s, stop: %s)|> filter(fn: (r) => r._measurement == \"%v\" %v) %v |> sort() |> schema.fieldsAsCols()",
		i.analyticsBucket,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		MarketBalanceTable,
		marketIDsFilter,
		pagination,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	response := make(map[string][]domain.MarketBalance)
	for result.Next() {
		marketID := result.Record().ValueByKey(marketTag).(string)
		marketBalance := domain.MarketBalance{
			MarketID:     result.Record().ValueByKey(marketTag).(string),
			BaseBalance:  int(result.Record().ValueByKey(baseBalance).(int64)),
			BaseAsset:    result.Record().ValueByKey(baseAsset).(string),
			QuoteBalance: int(result.Record().ValueByKey(quoteBalance).(int64)),
			QuoteAsset:   result.Record().ValueByKey(quoteAsset).(string),
			Time:         result.Record().Time(),
		}
		val, ok := response[marketID]
		if !ok {
			balances := make([]domain.MarketBalance, 0)
			balances = append(balances, marketBalance)
			response[marketID] = balances
		} else {
			val = append(val, marketBalance)
			response[marketID] = val
		}
	}

	return response, nil
}

func createMarkedIDsFluxQueryFilter(marketIDs []string) string {
	query := ""
	for i, v := range marketIDs {
		if i == 0 {
			query = fmt.Sprintf("and r.market_id==\"%v\"", v)
		} else {
			query = fmt.Sprintf("%v or r.market_id==\"%v\"", query, v)
		}
	}

	return query
}
