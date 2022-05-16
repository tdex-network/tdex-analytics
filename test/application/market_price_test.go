package influxdbtest

import (
	"context"
	"errors"
	"math"
	"tdex-analytics/internal/core/application"
	"time"
)

func (a *AppSvcTestSuit) TestGetMarketPrice() {
	type args struct {
		ctx               context.Context
		timeRange         application.TimeRange
		referenceCurrency string
		marketIDs         []string
	}
	tests := []struct {
		name             string
		args             args
		validateResponse func(prices *application.MarketsPrices) error
		wantErr          bool
	}{
		{
			name: "missing reference currency",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: nil,
					CustomPeriod:     nil,
				},
				referenceCurrency: "",
				marketIDs:         nil,
			},
			wantErr: true,
		},
		{
			name: "both PredefinedPeriod period and CustomPeriod cant be null",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: nil,
					CustomPeriod:     nil,
				},
				referenceCurrency: "EUR",
				marketIDs:         nil,
			},
			wantErr: true,
		},
		{
			name: "both PredefinedPeriod period and CustomPeriod provided",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &nilPp,
					CustomPeriod: &application.CustomPeriod{
						StartDate: "",
						EndDate:   "",
					},
				},
				referenceCurrency: "EUR",
				marketIDs:         nil,
			},
			wantErr: true,
		},
		{
			name: "empty custom period",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: nil,
					CustomPeriod: &application.CustomPeriod{
						StartDate: "",
						EndDate:   "",
					},
				},
				referenceCurrency: "EUR",
				marketIDs:         nil,
			},
			wantErr: true,
		},
		{
			name: "invalid custom period time format",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: nil,
					CustomPeriod: &application.CustomPeriod{
						StartDate: "Mon, 02 Jan 2006 15:04:05 -0700",
						EndDate:   "Mon, 02 Jan 2006 15:04:05 -0700",
					},
				},
				referenceCurrency: "EUR",
				marketIDs:         nil,
			},
			wantErr: true,
		},
		{
			name: "fetch prices for one market for last hour",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastHourPp,
					CustomPeriod:     nil,
				},
				referenceCurrency: "EUR",
				marketIDs:         []string{"1"},
			},
			validateResponse: func(prices *application.MarketsPrices) error {
				if len(prices.MarketsPrices) != 1 {
					return errors.New("expected prices for 1 markets")
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch prices for one market for last day",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastDayPp,
					CustomPeriod:     nil,
				},
				referenceCurrency: "EUR",
				marketIDs:         []string{"1"},
			},
			validateResponse: func(prices *application.MarketsPrices) error {
				if len(prices.MarketsPrices) != 1 {
					return errors.New("expected prices for 1 markets")
				}

				//loaded prices are sorted in asc order, validate that first one is from yesterday
				for _, v := range prices.MarketsPrices {
					if math.Round(v[0].Time.Sub(time.Now()).Hours()) != -24 {
						return errors.New("expected that first price is from yesterday")
					}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch prices for one market for last month",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastMonthPp,
					CustomPeriod:     nil,
				},
				referenceCurrency: "EUR",
				marketIDs:         []string{"1"},
			},
			validateResponse: func(prices *application.MarketsPrices) error {
				if len(prices.MarketsPrices) != 1 {
					return errors.New("expected prices for 1 markets")
				}

				//loaded prices are sorted in asc order, validate that first one is from prev month
				for _, v := range prices.MarketsPrices {
					if v[0].Time.Month() != time.Now().Month()-1 {
						return errors.New("expected that first price is from last month")
					}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch prices for one market for last 3 month",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastThreeMonthsPp,
					CustomPeriod:     nil,
				},
				referenceCurrency: "EUR",
				marketIDs:         []string{"1"},
			},
			validateResponse: func(prices *application.MarketsPrices) error {
				if len(prices.MarketsPrices) != 1 {
					return errors.New("expected prices for 1 markets")
				}

				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		a.Run(tt.name, func() {
			got, err := marketPriceSvc.GetPrices(
				tt.args.ctx,
				tt.args.timeRange,
				application.Page{
					Number: 1,
					Size:   10000000,
				},
				tt.args.referenceCurrency,
				tt.args.marketIDs...)
			if (err != nil) != tt.wantErr {
				a.T().Errorf("GetPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				if err := tt.validateResponse(got); err != nil {
					a.T().Error(err)
				}
			}
		})
	}
}
