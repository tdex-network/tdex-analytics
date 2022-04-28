package config

import (
	"reflect"
	"testing"
)

func TestGetAssetCurrencyPair(t *testing.T) {
	tests := []struct {
		name      string
		setEnvVar func()
		want      map[string]string
	}{
		{
			name: "1",
			setEnvVar: func() {
				t.Setenv(
					"TDEXA_ASSET_CURRENCY_PAIRS",
					"0x0000000000000000000000000000000000000000:LBTC,1x0000000000000000000000000000000000000000:USD",
				)
			},
			want: map[string]string{
				"0x0000000000000000000000000000000000000000": "LBTC",
				"1x0000000000000000000000000000000000000000": "USD",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnvVar()
			if got := GetAssetCurrencyPair(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAssetCurrencyPair() = %v, want %v", got, tt.want)
			}
		})
	}
}
