package dbinflux

import (
	"context"
	"errors"
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

func (i *influxDbService) GetBalancesForMarket(
	ctx context.Context,
	marketID string,
	fromTime time.Time,
) ([]domain.MarketBalance, error) {
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"from(bucket:\"%v\")|> range(start: %s)|> filter(fn: (r) => r._measurement == \"%v\" and r.market_id==\"%v\")|> sort()",
		i.analyticsBucket,
		fromTime.Format(time.RFC3339),
		MarketBalanceTable,
		marketID,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	resultPoints := make(map[time.Time]domain.MarketBalance)
	for result.Next() {
		val, ok := resultPoints[result.Record().Time()]
		if !ok {
			val = domain.MarketBalance{
				MarketID: result.Record().ValueByKey(marketTag).(string),
			}
		}

		switch field := result.Record().Field(); field {
		case baseAsset:
			val.BaseAsset = result.Record().Value().(string)
		case baseBalance:
			val.BaseBalance = int(result.Record().Value().(int64))
		case quoteAsset:
			val.QuoteAsset = result.Record().Value().(string)
		case quoteBalance:
			val.QuoteBalance = int(result.Record().Value().(int64))
		default:
			return nil, errors.New(fmt.Sprintf("unrecognized field %v", field))
		}

		resultPoints[result.Record().Time()] = val
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	response := make([]domain.MarketBalance, 0)
	for k, v := range resultPoints {
		response = append(response, domain.MarketBalance{
			MarketID:     v.MarketID,
			BaseBalance:  v.BaseBalance,
			BaseAsset:    v.BaseAsset,
			QuoteBalance: v.QuoteBalance,
			QuoteAsset:   v.QuoteAsset,
			Time:         k,
		})
	}

	return response, nil
}

func (i *influxDbService) GetBalancesForAllMarkets(
	ctx context.Context,
	fromTime time.Time,
) (map[string][]domain.MarketBalance, error) {
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"from(bucket:\"%v\")|> range(start: %s)|> filter(fn: (r) => r._measurement == \"%v\")|> sort()",
		i.analyticsBucket,
		fromTime.Format(time.RFC3339),
		MarketBalanceTable,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	resultPoints := make(map[time.Time]domain.MarketBalance)
	for result.Next() {
		val, ok := resultPoints[result.Record().Time()]
		if !ok {
			val = domain.MarketBalance{
				MarketID: result.Record().ValueByKey(marketTag).(string),
			}
		}

		switch field := result.Record().Field(); field {
		case baseAsset:
			val.BaseAsset = result.Record().Value().(string)
		case baseBalance:
			val.BaseBalance = int(result.Record().Value().(int64))
		case quoteAsset:
			val.QuoteAsset = result.Record().Value().(string)
		case quoteBalance:
			val.QuoteBalance = int(result.Record().Value().(int64))
		default:
			return nil, errors.New(fmt.Sprintf("unrecognized field %v", field))
		}

		resultPoints[result.Record().Time()] = val
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	response := make(map[string][]domain.MarketBalance)
	for k, v := range resultPoints {
		if v1, ok := response[v.MarketID]; !ok {
			balances := make([]domain.MarketBalance, 0)
			balances = append(balances, domain.MarketBalance{
				MarketID:     v.MarketID,
				BaseBalance:  v.BaseBalance,
				BaseAsset:    v.BaseAsset,
				QuoteBalance: v.QuoteBalance,
				QuoteAsset:   v.QuoteAsset,
				Time:         k,
			})
			response[v.MarketID] = balances
		} else {
			v1 = append(v1, domain.MarketBalance{
				MarketID:     v.MarketID,
				BaseBalance:  v.BaseBalance,
				BaseAsset:    v.BaseAsset,
				QuoteBalance: v.QuoteBalance,
				QuoteAsset:   v.QuoteAsset,
				Time:         k,
			})
			response[v.MarketID] = v1
		}
	}

	return response, nil
}
