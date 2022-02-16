package application

import "testing"

func TestMarketBalance_validate(t *testing.T) {
	type fields struct {
		MarketID     int
		BaseBalance  int64
		BaseAsset    string
		QuoteBalance int64
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
				MarketID:     0,
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
				MarketID:     0,
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
				MarketID:     1,
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
