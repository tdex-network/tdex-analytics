package domain

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil"
)

type Market struct {
	ID           int
	ProviderName string
	Url          string
	BaseAsset    string
	QuoteAsset   string
	Active       bool
}

func (m Market) Key() string {
	key := btcutil.Hash160(
		[]byte(fmt.Sprintf("%s%s%s", m.Url, m.BaseAsset, m.QuoteAsset)),
	)
	return hex.EncodeToString(key)
}

type Filter struct {
	Url        string
	BaseAsset  string
	QuoteAsset string
}
