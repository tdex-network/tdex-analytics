package tdexmarketloader

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetProvidersAndMarkets(t *testing.T) {
	tdexMarketLoaderSvc := NewService("127.0.0.1:9050")
	liquidityProviders, err := tdexMarketLoaderSvc.FetchProvidersMarkets(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, len(liquidityProviders) > 0)
}