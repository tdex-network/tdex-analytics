package dbinflux

import "testing"

func Test_createMarkedIDsFluxQueryFilter(t *testing.T) {
	type args struct {
		marketIDs []string
		table     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no market_id",
			args: args{
				marketIDs: []string{},
				table:     "market_price",
			},
			want: "(r._measurement == \"market_price\" and (r._field == \"base_price\" or r._field == \"quote_price\"))",
		},
		{
			name: "one market_id",
			args: args{
				marketIDs: []string{"2"},
				table:     "market_balance",
			},
			want: "(r._measurement == \"market_balance\" and r.market_id==\"2\" and (r._field == \"base_balance\" or r._field == \"quote_balance\"))",
		},
		{
			name: "multiple market_ids",
			args: args{
				marketIDs: []string{"2", "3", "4"},
				table:     "market_balance",
			},
			want: "(r._measurement == \"market_balance\" and r.market_id==\"2\" and (r._field == \"base_balance\" or r._field == \"quote_balance\")) or (r._measurement == \"market_balance\" and r.market_id==\"3\" and (r._field == \"base_balance\" or r._field == \"quote_balance\")) or (r._measurement == \"market_balance\" and r.market_id==\"4\" and (r._field == \"base_balance\" or r._field == \"quote_balance\"))",
		},
		{
			name: "multiple market_ids, price table",
			args: args{
				marketIDs: []string{"2", "3", "4"},
				table:     "market_price",
			},
			want: "(r._measurement == \"market_price\" and r.market_id==\"2\" and (r._field == \"base_price\" or r._field == \"quote_price\")) or (r._measurement == \"market_price\" and r.market_id==\"3\" and (r._field == \"base_price\" or r._field == \"quote_price\")) or (r._measurement == \"market_price\" and r.market_id==\"4\" and (r._field == \"base_price\" or r._field == \"quote_price\"))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMarkedIDsFluxQueryFilter(tt.args.marketIDs, tt.args.table); got != tt.want {
				t.Errorf("createMarkedIDsFluxQueryFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
