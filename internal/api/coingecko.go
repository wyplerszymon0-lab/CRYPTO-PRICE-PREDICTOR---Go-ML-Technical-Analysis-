package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.coingecko.com/api/v3"

type CoinGeckoClient struct {
	http *http.Client
}

func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

type marketChartResponse struct {
	Prices [][]float64 `json:"prices"`
}

func (c *CoinGeckoClient) FetchPrices(coinID string, days int) ([]float64, error) {
	url := fmt.Sprintf(
		"%s/coins/%s/market_chart?vs_currency=usd&days=%d&interval=daily",
		baseURL, coinID, days,
	)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data marketChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	prices := make([]float64, len(data.Prices))
	for i, p := range data.Prices {
		prices[i] = p[1]
	}
	return prices, nil
}
