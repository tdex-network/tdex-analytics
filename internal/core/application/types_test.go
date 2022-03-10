package application

import (
	"reflect"
	"testing"
	"time"
)

func TestMarketBalanceValidate(t *testing.T) {
	type fields struct {
		MarketID     string
		BaseBalance  int
		BaseAsset    string
		QuoteBalance int
		QuoteAsset   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "invalid market_id and assets",
			fields: fields{
				MarketID:     "0",
				BaseBalance:  0,
				BaseAsset:    "",
				QuoteBalance: 0,
				QuoteAsset:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid market_id",
			fields: fields{
				MarketID:     "0",
				BaseBalance:  0,
				BaseAsset:    "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
				QuoteBalance: 0,
				QuoteAsset:   "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			},
			wantErr: false,
		},
		{
			name: "happy path",
			fields: fields{
				MarketID:     "1",
				BaseBalance:  0,
				BaseAsset:    "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
				QuoteBalance: 0,
				QuoteAsset:   "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MarketBalance{
				MarketID:     tt.fields.MarketID,
				BaseBalance:  tt.fields.BaseBalance,
				BaseAsset:    tt.fields.BaseAsset,
				QuoteBalance: tt.fields.QuoteBalance,
				QuoteAsset:   tt.fields.QuoteAsset,
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeRange_getStartAndEndTime(t *testing.T) {
	lastHour := LastHour
	lastDay := LastDay
	lastMonth := LastMonth
	lastThreeMonths := LastThreeMonths
	yearToDate := YearToDate
	now := time.Date(StartYear, 2, 1, 15, 0, 0, 0, time.UTC)
	startOfYear := time.Date(StartYear, time.January, 1, 0, 0, 0, 0, time.UTC)

	tm, err := time.Parse(time.RFC3339, "2022-02-08T14:34:40+01:00")
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		PredefinedPeriod *PredefinedPeriod
		CustomPeriod     *CustomPeriod
	}
	type args struct {
		now time.Time
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantStartTime time.Time
		wantEndTime   time.Time
		wantErr       bool
	}{
		{
			name: "Last Hour",
			fields: fields{
				PredefinedPeriod: &lastHour,
				CustomPeriod:     nil,
			},
			args: args{
				now: now,
			},
			wantStartTime: now.Add(-time.Hour),
			wantEndTime:   now,
		},
		{
			name: "Last Day",
			fields: fields{
				PredefinedPeriod: &lastDay,
				CustomPeriod:     nil,
			},
			args: args{
				now: now,
			},
			wantStartTime: now.Add(-time.Hour * 24),
			wantEndTime:   now,
		},
		{
			name: "Last Month",
			fields: fields{
				PredefinedPeriod: &lastMonth,
				CustomPeriod:     nil,
			},
			args: args{
				now: now,
			},
			wantStartTime: now.AddDate(0, -1, 0),
			wantEndTime:   now,
		},
		{
			name: "Last 3 Month",
			fields: fields{
				PredefinedPeriod: &lastThreeMonths,
				CustomPeriod:     nil,
			},
			args: args{
				now: now,
			},
			wantStartTime: now.AddDate(0, -3, 0),
			wantEndTime:   now,
		},
		{
			name: "Year to Date",
			fields: fields{
				PredefinedPeriod: &yearToDate,
				CustomPeriod:     nil,
			},
			args: args{
				now: now,
			},
			wantStartTime: startOfYear,
			wantEndTime:   now,
		},
		{
			name: "Custom start/end date not provided",
			fields: fields{
				PredefinedPeriod: nil,
				CustomPeriod:     &CustomPeriod{},
			},
			args: args{
				now: now,
			},
			wantErr:       true,
			wantStartTime: time.Time{},
			wantEndTime:   time.Time{},
		},
		{
			name: "Custom start/end date not RFC3339 format",
			fields: fields{
				PredefinedPeriod: nil,
				CustomPeriod: &CustomPeriod{
					StartDate: "Mon, 02 Jan 2006 15:04:05 MST",
					EndDate:   "Mon, 02 Jan 2006 15:04:05 MST",
				},
			},
			args: args{
				now: now,
			},
			wantErr:       true,
			wantStartTime: time.Time{},
			wantEndTime:   time.Time{},
		},
		{
			name: "Custom start/end date valid RFC3339 format",
			fields: fields{
				PredefinedPeriod: nil,
				CustomPeriod: &CustomPeriod{
					StartDate: "2022-02-08T14:34:40+01:00",
					EndDate:   "2022-02-08T14:34:40+01:00",
				},
			},
			args: args{
				now: now,
			},
			wantErr:       false,
			wantStartTime: tm,
			wantEndTime:   tm,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := &TimeRange{
				PredefinedPeriod: tt.fields.PredefinedPeriod,
				CustomPeriod:     tt.fields.CustomPeriod,
			}
			gotStartTime, gotEndTime, err := t.getStartAndEndTime(tt.args.now)
			if (err != nil) != tt.wantErr {
				t1.Errorf("getStartAndEndTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotStartTime, tt.wantStartTime) {
				t1.Errorf("getStartAndEndTime() gotStartTime = %v, want %v", gotStartTime, tt.wantStartTime)
			}
			if !reflect.DeepEqual(gotEndTime, tt.wantEndTime) {
				t1.Errorf("getStartAndEndTime() gotEndTime = %v, want %v", gotEndTime, tt.wantEndTime)
			}
		})
	}
}

func TestMarketRequest_validate(t *testing.T) {
	type fields struct {
		Url        string
		BaseAsset  string
		QuoteAsset string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid regular url",
			fields: fields{
				Url:        "https://provider.tdex.network:9945",
				BaseAsset:  "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
				QuoteAsset: "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a",
			},
			wantErr: false,
		},
		{
			name: "valid onion url",
			fields: fields{
				Url:        "http://d7y3mzol3eo2tneqw5oytj23knm3734npwml4jzazrzzpy32e56lrxqd.onion:80",
				BaseAsset:  "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
				QuoteAsset: "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MarketProvider{
				Url:        tt.fields.Url,
				BaseAsset:  tt.fields.BaseAsset,
				QuoteAsset: tt.fields.QuoteAsset,
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
