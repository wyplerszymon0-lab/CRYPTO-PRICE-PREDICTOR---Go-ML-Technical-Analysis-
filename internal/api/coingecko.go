package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"
)

const (
	baseURL    = "https://api.coingecko.com/api/v3"
	maxRetries = 3
)

// CoinGeckoClient is a rate-limit-aware HTTP client for the CoinGecko public API.
type CoinGeckoClient struct {
	http *http.Client
}

// NewCoinGeckoClient creates a client with a 20-second per-request timeout.
func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

type marketChartResponse struct {
	Prices [][]float64 `json:"prices"`
}

// FetchPrices retrieves daily closing prices for coinID in USD.
// On transient failures it retries up to 3 times with exponential backoff (1s, 2s, 4s).
func (c *CoinGeckoClient) FetchPrices(ctx context.Context, coinID string, days int) ([]float64, error) {
	url := fmt.Sprintf(
		"%s/coins/%s/market_chart?vs_currency=usd&days=%d&interval=daily",
		baseURL, coinID, days,
	)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			slog.Debug("retry backoff", "coin", coinID, "attempt", attempt, "wait", backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		prices, err := c.fetchOnce(ctx, url)
		if err == nil {
			slog.Debug("fetched prices", "coin", coinID, "days", days, "points", len(prices))
			return prices, nil
		}
		lastErr = err
		slog.Debug("fetch failed", "coin", coinID, "attempt", attempt+1, "err", err)
	}
	return nil, fmt.Errorf("%s: all %d attempts failed: %w", coinID, maxRetries, lastErr)
}

func (c *CoinGeckoClient) fetchOnce(ctx context.Context, url string) ([]float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limited (429) — reduce --workers or wait before retrying")
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("unexpected HTTP %d", resp.StatusCode)
	}

	var data marketChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(data.Prices) == 0 {
		return nil, fmt.Errorf("empty price series returned by API")
	}

	prices := make([]float64, len(data.Prices))
	for i, p := range data.Prices {
		if len(p) < 2 {
			return nil, fmt.Errorf("malformed price entry at index %d", i)
		}
		prices[i] = p[1]
	}
	return prices, nil
}
