package dbpg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/lib/pq"
	"tdex-analytics/internal/core/domain"
	"tdex-analytics/internal/infrastructure/db/pg/sqlc/queries"
)

const (
	uniqueViolation = "23505"
)

func (p *postgresDbService) InsertMarket(
	ctx context.Context,
	market domain.Market,
) error {
	if _, err := p.querier.InsertMarket(ctx, queries.InsertMarketParams{
		ProviderName: market.ProviderName,
		Url:          market.Url,
		BaseAsset:    market.BaseAsset,
		QuoteAsset:   market.QuoteAsset,
	}); err != nil {
		if pqErr := err.(*pq.Error); pqErr != nil {
			if pqErr.Code == uniqueViolation {
				return nil
			}
		}
	}

	return nil
}

func (p *postgresDbService) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	markets, err := p.querier.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Market, 0, len(markets))
	for _, v := range markets {
		res = append(res, domain.Market{
			ID:           int(v.MarketID.Int32),
			ProviderName: v.ProviderName,
			Url:          v.Url,
			BaseAsset:    v.BaseAsset,
			QuoteAsset:   v.QuoteAsset,
		})
	}

	return res, nil
}

func (p *postgresDbService) GetAllMarketsForFilter(
	ctx context.Context,
	filter []domain.Filter,
	page domain.Page,
) ([]domain.Market, error) {
	res := make([]domain.Market, 0)
	limit := page.Size
	offset := page.Number*page.Size - page.Size
	pagination := fmt.Sprintf(
		" ORDER by market.market_id DESC LIMIT %v OFFSET %v",
		limit,
		offset,
	)
	query, values := generateQueryAndValues(filter)
	queryWithPagination := fmt.Sprintf("%v %v", query, pagination)

	rows, err := p.db.QueryContext(ctx, queryWithPagination, values...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		var providerName string
		var url string
		var baseAsset string
		var quoteAsset string

		if err = rows.Scan(&id, &providerName, &url, &baseAsset, &quoteAsset); err != nil {
			return nil, err
		}

		res = append(res, domain.Market{
			ID:           id,
			ProviderName: providerName,
			Url:          url,
			BaseAsset:    baseAsset,
			QuoteAsset:   quoteAsset,
		})
	}

	return res, nil
}

func generateQueryAndValues(filter []domain.Filter) (string, []interface{}) {
	query := bytes.NewBuffer([]byte("SELECT * FROM market"))
	queryCondition, values := parseFilter(filter)
	if queryCondition != "" {
		query.WriteString(" ")
		query.WriteString(queryCondition)
	}

	return query.String(), values
}

func parseFilter(filter []domain.Filter) (string, []interface{}) {
	var values []interface{}
	queryCondition := bytes.NewBuffer([]byte(""))

	if len(filter) > 0 {
		values = make([]interface{}, 0)
		j := 3
		for i, v := range filter {
			if i == 0 {
				queryCondition.WriteString("WHERE (url=$1 AND base_asset=$2 AND quote_asset=$3)")
			} else {
				queryCondition.WriteString(fmt.Sprintf(" OR (url=$%v AND base_asset=$%v AND quote_asset=$%v)", j+1, j+2, j+3))
				j = j + 3
			}

			values = append(values, v.Url)
			values = append(values, v.BaseAsset)
			values = append(values, v.QuoteAsset)
		}
	}

	return queryCondition.String(), values
}
