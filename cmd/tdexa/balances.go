package main

import (
	"context"
	"github.com/urfave/cli/v2"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/v1"
)

var listBalancesCmd = &cli.Command{
	Name:   "balances",
	Usage:  "list all balances",
	Action: listBalancesAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from_time",
			Usage:    "fetch balances from specific time in the past til now",
			Required: true,
		},
	},
}

func listBalancesAction(ctx *cli.Context) error {
	fromTime := ctx.String("from_time")

	req := &tdexav1.MarketsBalancesRequest{
		FromTime: fromTime,
	}

	client, cleanup, err := getAnalyticsClient()
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.MarketsBalances(context.Background(), req)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
