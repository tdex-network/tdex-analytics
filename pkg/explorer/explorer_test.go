package explorer

import (
	"context"
	"testing"
)

func Test_explorerService_GetAssetCurrency(t *testing.T) {
	exp := NewExplorerService()
	currency, err := exp.GetAssetCurrency(context.Background(), "b3f5ed2913486826fc58b35e9ae951c1680be6959c6f205c3cbc7fedb36d6b8a")
	if err != nil {
		t.Error(err)
	}
	t.Log(currency)
}
