package port

import "context"

type ExplorerService interface {
	GetAssetCurrency(ctx context.Context, assetId string) (string, error)
}
