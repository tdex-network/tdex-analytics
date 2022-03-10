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
			Name:  "start",
			Usage: "fetch balances from specific time in the past, please provide end flag also",
		},
		&cli.StringFlag{
			Name:  "end",
			Usage: "fetch balances from specific time in the past til end date, use with start flag",
		},
		&cli.StringSliceFlag{
			Name:  "market_id",
			Usage: "market_id to fetch balances for",
		},
		&cli.IntFlag{
			Name: "predefined_period",
			Usage: "time predefined periods:\n" +
				"       1 -> last hour\n" +
				"       2 -> last day\n" +
				"       3 -> last month\n" +
				"       4 -> last 3 months\n" +
				"       5 -> year to date\n" +
				"       6 -> all",
			Value: 2,
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

func listPricesAction(ctx *cli.Context) error {
	marketIDs := ctx.StringSlice("market_id")

	var customPeriod *tdexav1.CustomPeriod
	start := ctx.String("start")
	end := ctx.String("end")
	if start != "" && end != "" {
		customPeriod = &tdexav1.CustomPeriod{
			StartDate: start,
			EndDate:   end,
		}
	}

	var predefinedPeriod tdexav1.PredefinedPeriod
	pp := ctx.Int("predefined_period")
	if pp > 0 {
		predefinedPeriod = tdexav1.PredefinedPeriod(pp)
	}

	pageNum := ctx.Int64("page_num")
	pageSize := ctx.Int64("page_size")
	page := &tdexav1.Page{
		PageNumber: pageNum,
		PageSize:   pageSize,
	}

	req := &tdexav1.MarketsPricesRequest{
		TimeRange: &tdexav1.TimeRange{
			PredefinedPeriod: predefinedPeriod,
			CustomPeriod:     customPeriod,
		},
		MarketIds: marketIDs,
		Page:      page,
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
