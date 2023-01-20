package domain

type Market struct {
	ID           int
	ProviderName string
	Url          string
	BaseAsset    string
	QuoteAsset   string
	Active       bool
}

type Filter struct {
	Url        string
	BaseAsset  string
	QuoteAsset string
}
