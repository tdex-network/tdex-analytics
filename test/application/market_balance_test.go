package applicationtest

import (
	"context"
	"errors"
	"github.com/tdex-network/tdex-analytics/internal/core/application"
	"math"
	"time"
)

func (a *AppSvcTestSuit) TestGetMarketBalance() {
	type args struct {
		ctx       context.Context
		timeRange application.TimeRange
		marketIDs []string
		timeFrame application.TimeFrame
	}
	tests := []struct {
		name             string
		args             args
		validateResponse func(balances *application.MarketsBalances) error
		wantErr          bool
	}{
		{
			name: "both PredefinedPeriod period and CustomPeriod cant be null",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: nil,
					CustomPeriod:     nil,
				},
				marketIDs: nil,
				timeFrame: application.TimeFrameFourHours,
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
				marketIDs: nil,
				timeFrame: application.TimeFrameFourHours,
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
				marketIDs: nil,
				timeFrame: application.TimeFrameFourHours,
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
				marketIDs: nil,
				timeFrame: application.TimeFrameFourHours,
			},
			wantErr: true,
		},
		{
			name: "fetch balances for one market for last hour",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastHourPp,
					CustomPeriod:     nil,
				},
				marketIDs: []string{"1"},
				timeFrame: application.TzNil,
			},
			validateResponse: func(balances *application.MarketsBalances) error {
				if len(balances.MarketsBalances) != 1 {
					return errors.New("expected balances for 1 markets")
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch balances for one market for last day",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastDayPp,
					CustomPeriod:     nil,
				},
				marketIDs: []string{"1"},
				timeFrame: application.TzNil,
			},
			validateResponse: func(balances *application.MarketsBalances) error {
				if len(balances.MarketsBalances) != 1 {
					return errors.New("expected balances for 1 markets")
				}

				//loaded balances are sorted in asc order, validate that first one is from yesterday
				for _, v := range balances.MarketsBalances {
					if math.Round(v[0].Time.Sub(time.Now()).Hours()) != -24 {
						return errors.New("expected that first balance is from yesterday")
					}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch balances for one market for last month",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastMonthPp,
					CustomPeriod:     nil,
				},
				marketIDs: []string{"1"},
				timeFrame: application.TimeFrameFourHours,
			},
			validateResponse: func(balances *application.MarketsBalances) error {
				if len(balances.MarketsBalances) != 1 {
					return errors.New("expected balances for 1 markets")
				}

				//loaded balances are sorted in asc order, validate that first one is from prev month
				for _, v := range balances.MarketsBalances {
					if !isPreviousMonth(v[0].Time) {
						return errors.New("expected that first balance is from last month")
					}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "fetch balances for one market for last 3 month",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					PredefinedPeriod: &lastMonthPp,
					CustomPeriod:     nil,
				},
				marketIDs: []string{"1"},
				timeFrame: application.TimeFrameFourHours,
			},
			validateResponse: func(balances *application.MarketsBalances) error {
				if len(balances.MarketsBalances) != 1 {
					return errors.New("expected balances for 1 markets")
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "time frame nil for custom period",
			args: args{
				ctx: ctx,
				timeRange: application.TimeRange{
					CustomPeriod: &application.CustomPeriod{
						StartDate: fourHoursAgo,
						EndDate:   now,
					},
				},
				marketIDs: []string{"1"},
				timeFrame: application.TzNil,
			},
			validateResponse: func(balances *application.MarketsBalances) error {
				if len(balances.MarketsBalances) != 1 {
					return errors.New("expected balances for 1 markets")
				}

				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		a.Run(tt.name, func() {
			got, err := marketBalanceSvc.GetBalances(
				tt.args.ctx,
				tt.args.timeRange,
				application.Page{
					Number: 1,
					Size:   10000000,
				},
				tt.args.timeFrame,
				tt.args.marketIDs...)
			if (err != nil) != tt.wantErr {
				a.T().Errorf("GetBalances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				if err := tt.validateResponse(got); err != nil {
					a.T().Logf("debug got %v", got)
					a.T().Logf("debug got length %v", len(got.MarketsBalances))
					for k, v := range got.MarketsBalances {
						a.T().Logf("market %s", k)
						for _, p := range v {
							a.T().Logf("balance %v", p)
						}
					}
					a.T().Error(err)
				}
			}
		})
	}
}
