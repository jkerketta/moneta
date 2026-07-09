package data

import "github.com/jkerketta/stocktui/internal/models"

// Provider defines the interface for data sources.
type Provider interface {
	Name() string
	GetQuotes(symbols []string) ([]models.Quote, error)
	GetHistory(symbol string, tr models.TimeRange) ([]models.Candle, error)
}

// NewProvider returns the requested provider implementation.
func NewProvider(name string) (Provider, error) {
	return NewYahoo(), nil
}
