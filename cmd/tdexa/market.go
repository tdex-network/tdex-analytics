package main

import (
	"context"
	"errors"
	"strings"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/tdexa/v1"

	"github.com/urfave/cli/v2"
)

var marketsCmd = &cli.Command{
	Name:   "markets",
	Usage:  "list market's id's to be passed in prices/balances cmd's",
	Action: listMarkets,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "filter",
			Usage: "(Required) market provider url, base_asset, quote_asset delimited by ,",
		},
		&cli.Uint64Flag{
			Name:  "page_num",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
			Value: 1,
		},
		&cli.Uint64Flag{
			Name:  "page_size",
			Usage: "the size of the page",
			Value: 10,
		},
	},
}

func listMarkets(ctx *cli.Context) error {
	filters := ctx.StringSlice("filter")

	filterReq := make([]*tdexav1.MarketProvider, 0, len(filters))
	for _, v := range filters {
		tmp := strings.Split(strings.TrimSpace(v), ",")
		if len(tmp) != 3 {
			return errors.New("provide url, base_asset, quote_asset")
		}

		filterReq = append(filterReq, &tdexav1.MarketProvider{
			Url:        tmp[0],
			BaseAsset:  tmp[1],
			QuoteAsset: tmp[2],
		})
	}

	client, cleanup, err := getAnalyticsClient()
	if err != nil {
		return err
	}
	defer cleanup()

	pageNum := ctx.Int64("page_num")
	pageSize := ctx.Int64("page_size")
	page := &tdexav1.Page{
		PageNumber: pageNum,
		PageSize:   pageSize,
	}

	req := &tdexav1.ListMarketsRequest{
		MarketProviders: filterReq,
		Page:            page,
	}

	resp, err := client.ListMarkets(context.Background(), req)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
