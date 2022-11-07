package dbinflux

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"time"
)

func (i *influxDbService) InsertBalance(
	ctx context.Context,
	balance domain.MarketBalance,
) error {
	writeAPI := i.client.WriteAPI(i.org, i.analyticsBucket)

	bBalance, _ := balance.BaseBalance.Float64()
	qBalance, _ := balance.QuoteBalance.Float64()

	p := influxdb2.NewPointWithMeasurement(MarketBalanceTable).
		AddTag(marketTag, balance.MarketID).
		AddField(baseBalance, bBalance).
		AddField(quoteBalance, qBalance).
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
			"|> schema.fieldsAsCols()"+
			"%v"+
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

	response := make(map[string][]domain.MarketBalance)
	for result.Next() {
		marketID := result.Record().ValueByKey(marketTag).(string)
		bBalance := decimal.NewFromFloat(0)
		if result.Record().ValueByKey(baseBalance) != nil {
			bBalance = decimal.NewFromFloat(result.Record().ValueByKey(baseBalance).(float64))
		}
		qBalance := decimal.NewFromFloat(0)
		if result.Record().ValueByKey(quoteBalance) != nil {
			qBalance = decimal.NewFromFloat(result.Record().ValueByKey(quoteBalance).(float64))
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
	fieldsFilter := "and (r._field == \"base_price\" or r._field == \"quote_price\")"
	if table == MarketBalanceTable {
		fieldsFilter = "and (r._field == \"base_balance\" or r._field == \"quote_balance\")"
	}
	for i, v := range marketIDs {
		if i == 0 {
			query = fmt.Sprintf("%v and r.market_id==\"%v\" %v)", query, v, fieldsFilter)
		} else {
			query = fmt.Sprintf("%v or (r._measurement == \"%v\" and r.market_id==\"%v\" %v)", query, table, v, fieldsFilter)
		}
	}

	if len(marketIDs) == 0 {
		query = fmt.Sprintf("%v %v)", query, fieldsFilter)
	}

	return query
}
