package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jkerketta/stocktui/internal/models"
)

type Yahoo struct{}

func NewYahoo() *Yahoo {
	return &Yahoo{}
}

func (y *Yahoo) Name() string { return "Yahoo Finance" }

func (y *Yahoo) GetQuotes(symbols []string) ([]models.Quote, error) {
	baseURL := "https://query1.finance.yahoo.com/v7/finance/quote"
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))
	params.Set("fields", "symbol,regularMarketPrice,regularMarketChangePercent")

	fullURL := baseURL + "?" + params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	body, err := fetch(ctx, fullURL, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		QuoteResponse struct {
			Result []struct {
				Symbol                     string  `json:"symbol"`
				RegularMarketPrice         float64 `json:"regularMarketPrice"`
				RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"quoteResponse"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if resp.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("yahoo: %s", resp.QuoteResponse.Error.Description)
	}

	now := time.Now()
	quotes := make([]models.Quote, 0, len(resp.QuoteResponse.Result))
	for _, r := range resp.QuoteResponse.Result {
		if r.RegularMarketPrice == 0 {
			continue
		}
		quotes = append(quotes, models.Quote{
			Symbol:      r.Symbol,
			Price:       r.RegularMarketPrice,
			ChangePct:   r.RegularMarketChangePercent,
			LastUpdated: now,
		})
	}

	return quotes, nil
}

func (y *Yahoo) GetHistory(symbol string, tr models.TimeRange) ([]models.Candle, error) {
	var interval, rangeVal string
	switch tr {
	case models.Range1H:
		interval = "2m"
		rangeVal = "1d"
	case models.Range24H:
		interval = "5m"
		rangeVal = "1d"
	case models.Range7D:
		interval = "15m"
		rangeVal = "5d"
	case models.Range30D:
		interval = "1h"
		rangeVal = "1mo"
	default:
		interval = "5m"
		rangeVal = "1d"
	}

	baseURL := "https://query1.finance.yahoo.com/v8/finance/chart/" + url.PathEscape(symbol)
	params := url.Values{}
	params.Set("interval", interval)
	params.Set("range", rangeVal)
	params.Set("includePrePost", "false")

	fullURL := baseURL + "?" + params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	body, err := fetch(ctx, fullURL, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Chart struct {
			Result []struct {
				Timestamp  []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Open   []*float64 `json:"open"`
						High   []*float64 `json:"high"`
						Low    []*float64 `json:"low"`
						Close  []*float64 `json:"close"`
						Volume []*float64 `json:"volume"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo: %s", resp.Chart.Error.Description)
	}

	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	result := resp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 || len(result.Timestamp) == 0 {
		return nil, fmt.Errorf("no quote data for %s", symbol)
	}

	q := result.Indicators.Quote[0]
	candles := make([]models.Candle, 0, len(result.Timestamp))

	for i, ts := range result.Timestamp {
		// Skip null values (market closed periods)
		if i >= len(q.Close) || q.Close[i] == nil {
			continue
		}

		closeVal := *q.Close[i]
		if closeVal == 0 {
			continue
		}

		var openVal, highVal, lowVal, volVal float64
		if i < len(q.Open) && q.Open[i] != nil {
			openVal = *q.Open[i]
		} else {
			openVal = closeVal
		}
		if i < len(q.High) && q.High[i] != nil {
			highVal = *q.High[i]
		} else {
			highVal = closeVal
		}
		if i < len(q.Low) && q.Low[i] != nil {
			lowVal = *q.Low[i]
		} else {
			lowVal = closeVal
		}
		if i < len(q.Volume) && q.Volume[i] != nil {
			volVal = *q.Volume[i]
		}

		candles = append(candles, models.Candle{
			Timestamp: time.Unix(ts, 0),
			Open:      openVal,
			High:      highVal,
			Low:       lowVal,
			Close:     closeVal,
			Volume:    volVal,
		})
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no valid candles for %s", symbol)
	}

	return candles, nil
}
