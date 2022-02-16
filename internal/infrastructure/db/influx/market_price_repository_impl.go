package dbinflux

import (
	"context"
	"errors"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"tdex-analytics/internal/core/domain"
	"time"
)

func (i *influxDbService) InsertPrice(
	ctx context.Context,
	price domain.MarketPrice,
) error {
	writeAPI := i.client.WriteAPI(i.org, i.analyticsBucket)

	p := influxdb2.NewPointWithMeasurement(MarketPriceTable).
		AddTag(marketTag, price.MarketID).
		AddField(baseAsset, price.BaseAsset).
		AddField(basePrice, price.BasePrice).
		AddField(quoteAsset, price.QuoteAsset).
		AddField(quotePrice, price.QuotePrice).
		SetTime(price.Time)

	writeAPI.WritePoint(p)

	writeAPI.Flush()

	return nil
}

func (i *influxDbService) GetPricesForMarket(
	ctx context.Context,
	marketID string,
	fromTime time.Time,
) ([]domain.MarketPrice, error) {
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"from(bucket:\"%v\")|> range(start: %s)|> filter(fn: (r) => r._measurement == \"%v\" and r.market_id==\"%v\")|> sort()",
		i.analyticsBucket,
		fromTime.Format(time.RFC3339),
		MarketPriceTable,
		marketID,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	resultPoints := make(map[time.Time]domain.MarketPrice)
	for result.Next() {
		val, ok := resultPoints[result.Record().Time()]
		if !ok {
			val = domain.MarketPrice{
				MarketID: result.Record().ValueByKey(marketTag).(string),
			}
		}

		switch field := result.Record().Field(); field {
		case baseAsset:
			val.BaseAsset = result.Record().Value().(string)
		case basePrice:
			val.BasePrice = int(result.Record().Value().(int64))
		case quoteAsset:
			val.QuoteAsset = result.Record().Value().(string)
		case quotePrice:
			val.QuotePrice = int(result.Record().Value().(int64))
		default:
			return nil, errors.New(fmt.Sprintf("unrecognized field %v", field))
		}

		resultPoints[result.Record().Time()] = val
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	response := make([]domain.MarketPrice, 0)
	for k, v := range resultPoints {
		response = append(response, domain.MarketPrice{
			MarketID:   v.MarketID,
			BasePrice:  v.BasePrice,
			BaseAsset:  v.BaseAsset,
			QuotePrice: v.QuotePrice,
			QuoteAsset: v.QuoteAsset,
			Time:       k,
		})
	}

	return response, nil
}

func (i *influxDbService) GetPricesForAllMarkets(
	ctx context.Context,
	fromTime time.Time,
) (map[string][]domain.MarketPrice, error) {
	queryAPI := i.client.QueryAPI(i.org)
	query := fmt.Sprintf(
		"from(bucket:\"%v\")|> range(start: %s)|> filter(fn: (r) => r._measurement == \"%v\")|> sort()",
		i.analyticsBucket,
		fromTime.Format(time.RFC3339),
		MarketPriceTable,
	)
	result, err := queryAPI.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	resultPoints := make(map[time.Time]domain.MarketPrice)
	for result.Next() {
		val, ok := resultPoints[result.Record().Time()]
		if !ok {
			val = domain.MarketPrice{
				MarketID: result.Record().ValueByKey(marketTag).(string),
			}
		}

		switch field := result.Record().Field(); field {
		case baseAsset:
			val.BaseAsset = result.Record().Value().(string)
		case basePrice:
			val.BasePrice = int(result.Record().Value().(int64))
		case quoteAsset:
			val.QuoteAsset = result.Record().Value().(string)
		case quotePrice:
			val.QuotePrice = int(result.Record().Value().(int64))
		default:
			return nil, errors.New(fmt.Sprintf("unrecognized field %v", field))
		}

		resultPoints[result.Record().Time()] = val
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	response := make(map[string][]domain.MarketPrice)
	for k, v := range resultPoints {
		if v1, ok := response[v.MarketID]; !ok {
			prices := make([]domain.MarketPrice, 0)
			prices = append(prices, domain.MarketPrice{
				MarketID:   v.MarketID,
				BasePrice:  v.BasePrice,
				BaseAsset:  v.BaseAsset,
				QuotePrice: v.QuotePrice,
				QuoteAsset: v.QuoteAsset,
				Time:       k,
			})
			response[v.MarketID] = prices
		} else {
			v1 = append(v1, domain.MarketPrice{
				MarketID:   v.MarketID,
				BasePrice:  v.BasePrice,
				BaseAsset:  v.BaseAsset,
				QuotePrice: v.QuotePrice,
				QuoteAsset: v.QuoteAsset,
				Time:       k,
			})
			response[v.MarketID] = v1
		}
	}

	return response, nil
}
