package dbinflux

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"time"
)

func (i *influxDbService) InsertBalance(
	ctx context.Context,
	balance domain.MarketBalance,
) error {
	writeAPI := i.client.WriteAPI(i.org, i.analyticsBucket)

	p := influxdb2.NewPointWithMeasurement(MarketBalanceTable).
		AddTag(marketTag, balance.MarketID).
		AddField(baseBalance, balance.BaseBalance).
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
	groupBy string,
	marketIDs ...string,
) (map[string][]domain.MarketBalance, error) {
	limit := page.Size
	offset := page.Number*page.Size - page.Size
	pagination := fmt.Sprintf("|> limit(n: %v, offset: %v)", limit, offset)
	marketIDsFilter := createMarkedIDsFluxQueryFilter(marketIDs, MarketBalanceTable)
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"import \"influxdata/influxdb/schema\" from(bucket:\"%v\")"+
			"|> range(start: %s, stop: %s)"+
			"|> filter(fn: (r) => %v)"+
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

	response := make(map[string][]domain.MarketBalance)
	for result.Next() {
		marketID := result.Record().ValueByKey(marketTag).(string)
		bBalance := 0
		if result.Record().ValueByKey(baseBalance) != nil {
			bBalance = int(result.Record().ValueByKey(baseBalance).(int64))
		}
		qBalance := 0
		if result.Record().ValueByKey(quoteBalance) != nil {
			qBalance = int(result.Record().ValueByKey(quoteBalance).(int64))
		}

		marketBalance := domain.MarketBalance{
			MarketID:     marketID,
			BaseBalance:  bBalance,
			QuoteBalance: qBalance,
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

func createMarkedIDsFluxQueryFilter(marketIDs []string, table string) string {
	query := fmt.Sprintf("(r._measurement == \"%v\"", table)
	for i, v := range marketIDs {
		if i == 0 {
			query = fmt.Sprintf("%v and r.market_id==\"%v\")", query, v)
		} else {
			query = fmt.Sprintf("%v or (r._measurement == \"%v\" and r.market_id==\"%v\")", query, table, v)
		}
	}

	if len(marketIDs) == 0 {
		query = fmt.Sprintf("%v)", query)
	}

	return query
}
