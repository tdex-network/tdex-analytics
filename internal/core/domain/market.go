package domain

type Market struct {
	ID           int
	ProviderName string
	Url          string
	BaseAsset    string
	QuoteAsset   string
}

type Filter struct {
	Url        string
	BaseAsset  string
	QuoteAsset string
}
