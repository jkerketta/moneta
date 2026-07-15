package models

import "time"

// TimeRange represents the chart history range.
type TimeRange string

const (
	Range1H  TimeRange = "1H"
	Range24H TimeRange = "24H"
	Range7D  TimeRange = "7D"
	Range30D TimeRange = "30D"
)

// Holding represents a single portfolio position.
type Holding struct {
	Symbol   string  `yaml:"symbol"`
	Shares   float64 `yaml:"shares"`
	AvgPrice float64 `yaml:"avg_price"`
	Currency string  `yaml:"currency"`
}

// Portfolio is the top-level portfolio YAML structure.
type Portfolio struct {
	Holdings []Holding `yaml:"holdings"`
}

// NewsItem from Yahoo Finance search API.
type NewsItem struct {
	Headline string `json:"title"`
	URL      string `json:"link"`
	Source   string `json:"publisher"`
	Datetime int64  `json:"providerPublishTime"`
	Summary  string `json:"summary"`
	Related  string `json:"-"` // populated by the fetch code
}

// Quote represents a snapshot of an asset's price.
type Quote struct {
	Symbol      string
	Price       float64
	ChangePct   float64
	LastUpdated time.Time
}

// Candle represents a single data point in a historical chart.
type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// AppConfig holds the complete run configuration.
type AppConfig struct {
	Symbols         []string      `mapstructure:"symbols"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	Provider        string        `mapstructure:"provider"`
	DefaultRange    string        `mapstructure:"default_range"`
}
