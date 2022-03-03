package dbpg

import (
	"tdex-analytics/internal/core/domain"
	"testing"
)

func Test_generateQueryAndValues(t *testing.T) {
	type args struct {
		filter []domain.Filter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{
				filter: []domain.Filter{},
			},
			want: "SELECT * FROM market",
		},
		{
			name: "2",
			args: args{
				filter: []domain.Filter{
					{
						Url:        "test1",
						BaseAsset:  "test2",
						QuoteAsset: "test3",
					},
				},
			},
			want: "SELECT * FROM market WHERE (url=$1 AND base_asset=$2 AND quote_asset=$3)",
		},
		{
			name: "3",
			args: args{
				filter: []domain.Filter{
					{
						Url:        "test1",
						BaseAsset:  "test2",
						QuoteAsset: "test3",
					},
					{
						Url:        "test4",
						BaseAsset:  "test5",
						QuoteAsset: "test6",
					},
					{
						Url:        "test7",
						BaseAsset:  "test8",
						QuoteAsset: "test9",
					},
				},
			},
			want: "SELECT * FROM market WHERE (url=$1 AND base_asset=$2 AND quote_asset=$3) OR (url=$4 AND base_asset=$5 AND quote_asset=$6) OR (url=$7 AND base_asset=$8 AND quote_asset=$9)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := generateQueryAndValues(tt.args.filter)
			if got != tt.want {
				t.Errorf("generateQueryAndValues() \ngot = %v, \nwant = %v", got, tt.want)
			}
		})
	}
}
