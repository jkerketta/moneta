package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	now := time.Now()
	quotes := make([]models.Quote, 0, len(symbols))
	var firstErr error

	for _, symbol := range symbols {
		q, err := y.fetchQuoteFromChart(symbol)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		q.LastUpdated = now
		quotes = append(quotes, q)
	}

	if len(quotes) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return quotes, nil
}

// fetchQuoteFromChart uses the Yahoo chart endpoint (still publicly
// reachable) because the dedicated /v7/finance/quote API now returns
// Unauthorized for unauthenticated clients.
func (y *Yahoo) fetchQuoteFromChart(symbol string) (models.Quote, error) {
	baseURL := "https://query1.finance.yahoo.com/v8/finance/chart/" + url.PathEscape(symbol)
	params := url.Values{}
	params.Set("interval", "1d")
	params.Set("range", "5d")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	body, err := fetch(ctx, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return models.Quote{}, err
	}

	var resp struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol                   string   `json:"symbol"`
					RegularMarketPrice       float64  `json:"regularMarketPrice"`
					RegularMarketChangePct   *float64 `json:"regularMarketChangePercent"`
					ChartPreviousClose       float64  `json:"chartPreviousClose"`
					PreviousClose            float64  `json:"previousClose"`
				} `json:"meta"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return models.Quote{}, fmt.Errorf("parse error: %w", err)
	}
	if resp.Chart.Error != nil {
		return models.Quote{}, fmt.Errorf("yahoo: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return models.Quote{}, fmt.Errorf("no data for %s", symbol)
	}

	meta := resp.Chart.Result[0].Meta
	if meta.RegularMarketPrice == 0 {
		return models.Quote{}, fmt.Errorf("no price for %s", symbol)
	}

	changePct := 0.0
	if meta.RegularMarketChangePct != nil {
		changePct = *meta.RegularMarketChangePct
	} else {
		prev := meta.ChartPreviousClose
		if prev == 0 {
			prev = meta.PreviousClose
		}
		if prev > 0 {
			changePct = (meta.RegularMarketPrice - prev) / prev * 100
		}
	}

	sym := meta.Symbol
	if sym == "" {
		sym = symbol
	}

	return models.Quote{
		Symbol:    sym,
		Price:     meta.RegularMarketPrice,
		ChangePct: changePct,
	}, nil
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

// FetchNews fetches news for a single ticker symbol from Yahoo Finance's
// search endpoint — no auth required.
func (y *Yahoo) FetchNews(symbol string, count int) ([]models.NewsItem, error) {
	baseURL := "https://query1.finance.yahoo.com/v1/finance/search"
	params := url.Values{}
	params.Set("q", symbol)
	params.Set("quotesCount", "0")
	params.Set("newsCount", fmt.Sprintf("%d", count))

	fullURL := baseURL + "?" + params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	body, err := fetch(ctx, fullURL, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		News []struct {
			Title              string `json:"title"`
			Link               string `json:"link"`
			Publisher          string `json:"publisher"`
			ProviderPublishTime int64  `json:"providerPublishTime"`
			Summary            string `json:"summary"`
		} `json:"news"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("yahoo news parse error: %w", err)
	}

	items := make([]models.NewsItem, 0, len(resp.News))
	for _, n := range resp.News {
		items = append(items, models.NewsItem{
			Headline: n.Title,
			URL:      n.Link,
			Source:   n.Publisher,
			Datetime: n.ProviderPublishTime,
			Summary:  n.Summary,
			Related:  symbol,
		})
	}
	return items, nil
}

// FetchAllNews fetches news for multiple symbols with deduplication.
func (y *Yahoo) FetchAllNews(symbols []string) ([]models.NewsItem, error) {
	var all []models.NewsItem
	seen := make(map[string]bool)

	for _, sym := range symbols {
		items, err := y.FetchNews(sym, 5)
		if err != nil {
			continue
		}
		for _, item := range items {
			key := item.URL
			if key == "" {
				key = strings.ToLower(item.Headline)
			}
			if !seen[key] {
				seen[key] = true
				all = append(all, item)
			}
		}
	}
	return all, nil
}

// changePctRe extracts the daily change percentage from a Yahoo Finance quote
// page header, e.g. "(+0.38%)" or "(-0.07%)".
var changePctRe = regexp.MustCompile(`\(([+-]\d+\.\d+)%\)`)

// ScrapeMarketChange fetches a Yahoo Finance quote page and extracts the daily
// change percentage for the given symbol — exactly as shown on yahoo.com.
func ScrapeMarketChange(symbol string) (float64, error) {
	path := url.PathEscape(symbol)
	u := "https://finance.yahoo.com/quote/" + path

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "stock-tui/1.0")
	req.Header.Set("Accept", "text/html")

	resp, err := defaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	m := changePctRe.FindSubmatch(body)
	if len(m) < 2 {
		return 0, fmt.Errorf("no change percentage found for %s", symbol)
	}

	var pct float64
	if _, err := fmt.Sscanf(string(m[1]), "%f", &pct); err != nil {
		return 0, fmt.Errorf("parse change percentage for %s: %w", symbol, err)
	}
	return pct, nil
}
