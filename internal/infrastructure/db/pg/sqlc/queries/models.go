// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package queries

import (
	"database/sql"
)

type Market struct {
	MarketID     sql.NullInt32
	ProviderName string
	Url          string
	BaseAsset    string
	QuoteAsset   string
	Active       sql.NullBool
}
