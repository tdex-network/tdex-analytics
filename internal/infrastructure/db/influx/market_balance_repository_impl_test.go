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
				table:     "market_balance",
			},
			want: "(r._measurement == \"market_balance\")",
		},
		{
			name: "one market_id",
			args: args{
				marketIDs: []string{"2"},
				table:     "market_balance",
			},
			want: "(r._measurement == \"market_balance\" and r.market_id==\"2\")",
		},
		{
			name: "multiple market_ids",
			args: args{
				marketIDs: []string{"2", "3", "4"},
				table:     "market_balance",
			},
			want: "(r._measurement == \"market_balance\" and r.market_id==\"2\") or (r._measurement == \"market_balance\" and r.market_id==\"3\") or (r._measurement == \"market_balance\" and r.market_id==\"4\")",
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
