package main

import (
	"context"
	"github.com/urfave/cli/v2"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/v1"
)

var listPricesCmd = &cli.Command{
	Name:   "prices",
	Usage:  "list all prices",
	Action: listPricesAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from_time",
			Usage:    "fetch prices from specific time in the past til now",
			Required: true,
		},
	},
}

func listPricesAction(ctx *cli.Context) error {
	fromTime := ctx.String("from_time")

	req := &tdexav1.MarketsPricesRequest{
		FromTime: fromTime,
	}

	client, cleanup, err := getAnalyticsClient()
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.MarketsPrices(context.Background(), req)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
