package main

import (
	"context"
	"errors"
	"github.com/urfave/cli/v2"
	"strings"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/v1"
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
	},
}

func listMarkets(ctx *cli.Context) error {
	filters := ctx.StringSlice("filter")

	filterReq := make([]*tdexav1.MarketRequest, 0, len(filters))
	for _, v := range filters {
		tmp := strings.Split(strings.TrimSpace(v), ",")
		if len(tmp) != 3 {
			return errors.New("provide url, base_asset, quote_asset")
		}

		filterReq = append(filterReq, &tdexav1.MarketRequest{
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

	req := &tdexav1.ListMarketIDsRequest{
		MarketsRequest: filterReq,
	}

	resp, err := client.ListMarketIDs(context.Background(), req)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
