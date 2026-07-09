package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jkerketta/stocktui/internal/models"
)

const finnhubBase = "https://finnhub.io/api/v1"

func FinnhubAPIKey() string {
	return os.Getenv("FINNHUB_API_KEY")
}

// FetchNews gets company news for a symbol from Finnhub.
// Uses the last 3 days as date range.
func FetchNews(symbol string) ([]models.NewsItem, error) {
	key := FinnhubAPIKey()
	if key == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not set")
	}

	now := time.Now()
	from := now.AddDate(0, 0, -3).Format("2006-01-02")
	to := now.Format("2006-01-02")

	url := fmt.Sprintf("%s/company-news?symbol=%s&from=%s&to=%s&token=%s",
		finnhubBase, symbol, from, to, key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Finnhub API error: %s", resp.Status)
	}

	var news []models.NewsItem
	if err := json.NewDecoder(resp.Body).Decode(&news); err != nil {
		return nil, err
	}

	if news == nil {
		return []models.NewsItem{}, nil
	}
	return news, nil
}

// FetchAllNews fetches news for multiple symbols.
func FetchAllNews(symbols []string) ([]models.NewsItem, error) {
	var all []models.NewsItem
	seen := make(map[string]bool)

	for _, sym := range symbols {
		items, err := FetchNews(sym)
		if err != nil {
			continue
		}
		for _, item := range items {
			key := fmt.Sprintf("%s-%d-%s", item.Related, item.Datetime, item.Headline)
			if !seen[key] {
				seen[key] = true
				all = append(all, item)
			}
		}
	}

	return all, nil
}
