package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	coingeckoBaseURL = "https://api.coingecko.com/api/v3"
	coinbaseBaseURL  = "https://api.exchange.coinbase.com"
)

// PriceProvider is an interface for fetching OHLCV data.
type PriceProvider interface {
	FetchOHLC(symbol, view string, count int) ([]Candle, error)
}

// HTTPClient interface for testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CoinGeckoProvider fetches data from CoinGecko's free API.
type CoinGeckoProvider struct {
	client HTTPClient
}

// CoinbaseProvider fetches data from Coinbase's public API.
type CoinbaseProvider struct {
	client HTTPClient
}

// NewCoinGeckoProvider creates a new CoinGecko provider.
func NewCoinGeckoProvider() *CoinGeckoProvider {
	return &CoinGeckoProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewCoinbaseProvider creates a new Coinbase provider.
func NewCoinbaseProvider() *CoinbaseProvider {
	return &CoinbaseProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchOHLC fetches OHLCV data from CoinGecko.
func (cg *CoinGeckoProvider) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	if view == "" {
		view = "1D"
	}

	coinID := coingeckoIDForSymbol(symbol)
	if coinID == "" {
		return nil, fmt.Errorf("unsupported symbol for CoinGecko: %s", symbol)
	}

	days := daysForView(view)
	debug := os.Getenv("CMC_DEBUG") == "1"

	v := url.Values{}
	v.Set("vs_currency", "usd")
	v.Set("days", fmt.Sprintf("%d", days))

	u := fmt.Sprintf("%s/coins/%s/ohlc?%s", coingeckoBaseURL, coinID, v.Encode())
	if debug {
		log.Printf("[CG] Requesting OHLC: %s", u)
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := cg.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if debug {
			log.Printf("[CG] Non-200 response: status=%s body=%s", resp.Status, string(body))
		}
		return nil, fmt.Errorf("CoinGecko API error: status=%s", resp.Status)
	}

	var raw [][]float64
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var candles []Candle
	for _, entry := range raw {
		if len(entry) < 5 {
			continue
		}
		ts := time.UnixMilli(int64(entry[0]))
		candles = append(candles, Candle{
			Time:   ts,
			Open:   entry[1],
			High:   entry[2],
			Low:    entry[3],
			Close:  entry[4],
			Volume: 0,
		})
	}

	return candles, nil
}

// FetchOHLC fetches OHLCV data from Coinbase.
func (cb *CoinbaseProvider) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	if view == "" {
		view = "1D"
	}

	product := strings.ToUpper(strings.TrimSpace(symbol))
	if !strings.Contains(product, "-") {
		product = product + "-USD"
	}

	debug := os.Getenv("CMC_DEBUG") == "1"

	end := time.Now().UTC()
	start := end.Add(-durationForView(view))
	gran := granularityForView(view)
	maxSpan := time.Duration(gran*300) * time.Second

	var all []Candle
	for cur := start; cur.Before(end); {
		curEnd := cur.Add(maxSpan)
		if curEnd.After(end) {
			curEnd = end
		}

		v := url.Values{}
		v.Set("granularity", fmt.Sprintf("%d", gran))
		v.Set("start", cur.Format(time.RFC3339))
		v.Set("end", curEnd.Format(time.RFC3339))

		u := fmt.Sprintf("%s/products/%s/candles?%s", coinbaseBaseURL, url.PathEscape(product), v.Encode())
		if debug {
			log.Printf("[CB] Requesting OHLC: %s", u)
		}

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := cb.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if debug {
				log.Printf("[CB] Non-200 response: status=%s", resp.Status)
			}
			return nil, fmt.Errorf("Coinbase API error: status=%s", resp.Status)
		}

		var raw [][]float64
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, err
		}

		for _, entry := range raw {
			if len(entry) < 6 {
				continue
			}
			ts := time.Unix(int64(entry[0]), 0).UTC()
			all = append(all, Candle{
				Time:   ts,
				Open:   entry[3],
				High:   entry[2],
				Low:    entry[1],
				Close:  entry[4],
				Volume: entry[5],
			})
		}

		cur = curEnd
		time.Sleep(150 * time.Millisecond)
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("no candles returned")
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Time.Before(all[j].Time) })

	// Remove duplicates
	out := all[:0]
	var last time.Time
	for _, cndl := range all {
		if !last.IsZero() && cndl.Time.Equal(last) {
			continue
		}
		out = append(out, cndl)
		last = cndl.Time
	}

	return out, nil
}

// CoinMarketCapProvider fetches data using a paid API key.
type CoinMarketCapProvider struct {
	client HTTPClient
	apiKey string
}

func NewCoinMarketCapProvider(apiKey string) *CoinMarketCapProvider {
	return &CoinMarketCapProvider{
		client: &http.Client{Timeout: 10 * time.Second},
		apiKey: apiKey,
	}
}

func (cmc *CoinMarketCapProvider) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	if cmc.apiKey == "" {
		return nil, fmt.Errorf("no API key provided for CoinMarketCap")
	}

	debug := os.Getenv("CMC_DEBUG") == "1"
	// Map view to CMC interval
	interval := "daily"
	switch strings.ToUpper(view) {
	case "1D":
		interval = "hourly"
	case "WTD", "MTD":
		interval = "daily"
	case "YTD":
		interval = "weekly"
	}

	// Endpoint: /v2/cryptocurrency/ohlcv/historical
	// Note: CMC requires 'symbol' and 'interval'
	u := "https://pro-api.coinmarketcap.com/v2/cryptocurrency/ohlcv/historical"
	v := url.Values{}
	v.Set("symbol", strings.Split(symbol, "-")[0])
	v.Set("interval", interval)
	v.Set("count", fmt.Sprintf("%d", count))

	reqUrl := fmt.Sprintf("%s?%s", u, v.Encode())
	if debug {
		log.Printf("[CMC] Requesting OHLC: %s", reqUrl)
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", cmc.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := cmc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinMarketCap API error: status=%s", resp.Status)
	}

	// CMC response structure is complex, we target the quotes
	var result struct {
		Data map[string][]struct {
			Quotes []struct {
				Quote struct {
					USD struct {
						Open      float64   `json:"open"`
						High      float64   `json:"high"`
						Low       float64   `json:"low"`
						Close     float64   `json:"close"`
						Volume    float64   `json:"volume"`
						Timestamp time.Time `json:"timestamp"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"quotes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var candles []Candle
	for _, symData := range result.Data {
		for _, q := range symData {
			for _, item := range q.Quotes {
				val := item.Quote.USD
				candles = append(candles, Candle{
					Time:   val.Timestamp,
					Open:   val.Open,
					High:   val.High,
					Low:    val.Low,
					Close:  val.Close,
					Volume: val.Volume,
				})
			}
		}
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles returned from CMC")
	}

	return candles, nil
}

// Client tries multiple providers and returns the first successful result.
type Client struct {
	providers []PriceProvider
}

// NewClient creates a new client with default providers.
func NewClient() *Client {
	return &Client{
		providers: []PriceProvider{
			NewCoinGeckoProvider(),
			NewCoinbaseProvider(),
		},
	}
}

// AddProvider adds a provider to the front of the list (higher priority).
func (c *Client) AddProvider(p PriceProvider) {
	c.providers = append([]PriceProvider{p}, c.providers...)
}

// FetchOHLC tries each provider in order and returns the first successful result.
func (c *Client) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	var lastErr error
	for i, provider := range c.providers {
		candles, err := provider.FetchOHLC(symbol, view, count)
		if err == nil && len(candles) > 0 {
			return candles, nil
		}
		lastErr = err
		if i < len(c.providers)-1 {
			time.Sleep(250 * time.Millisecond)
		}
	}
	return nil, fmt.Errorf("all providers failed for %s (%s): %v", symbol, view, lastErr)
}

// coingeckoIDForSymbol maps a ticker symbol to a CoinGecko coin ID.
func coingeckoIDForSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if parts := strings.SplitN(s, "-", 2); len(parts) > 1 {
		s = parts[0]
	}
	switch s {
	case "BTC":
		return "bitcoin"
	case "ETH":
		return "ethereum"
	case "SOL":
		return "solana"
	case "LTC":
		return "litecoin"
	case "ADA":
		return "cardano"
	case "AVAX":
		return "avalanche-2"
	case "MATIC":
		return "matic-network"
	case "DOGE":
		return "dogecoin"
	case "XRP":
		return "ripple"
	case "XMR":
		return "monero"
	}
	return strings.ToLower(s)
}

// daysForView returns the number of days for a given view.
func daysForView(view string) int {
	switch strings.ToUpper(view) {
	case "1D":
		return 1
	case "WTD":
		return 7
	case "MTD":
		return 31
	case "YTD":
		return 365
	default:
		return 1
	}
}

// durationForView returns the duration for a given view.
func durationForView(view string) time.Duration {
	switch strings.ToUpper(view) {
	case "1D":
		return 24 * time.Hour
	case "WTD":
		return 7 * 24 * time.Hour
	case "MTD":
		return 31 * 24 * time.Hour
	case "YTD":
		return 365 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// granularityForView returns the Coinbase granularity in seconds.
func granularityForView(view string) int {
	switch strings.ToUpper(view) {
	case "1D":
		return 300 // 5 minutes
	case "WTD":
		return 3600 // 1 hour
	case "MTD":
		return 21600 // 6 hours
	case "YTD":
		return 86400 // 1 day
	default:
		return 300
	}
}
