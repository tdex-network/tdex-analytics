package domain

type Market struct {
	ID           int
	AccountIndex int
	ProviderName string
	Url          string
	Credentials  string
	BaseAsset    string
	QuoteAsset   string
}
