package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCoinGeckoIDForSymbol(t *testing.T) {
	tests := []struct {
		symbol   string
		expected string
	}{
		{"BTC", "bitcoin"},
		{"BTC-USD", "bitcoin"},
		{"ETH", "ethereum"},
		{"ETH-USD", "ethereum"},
		{"SOL", "solana"},
		{"LTC", "litecoin"},
		{"DOGE", "dogecoin"},
		{"XRP", "ripple"},
		{"UNKNOWN", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			id := coingeckoIDForSymbol(tt.symbol)
			if id != tt.expected {
				t.Errorf("coingeckoIDForSymbol(%s) = %s, want %s", tt.symbol, id, tt.expected)
			}
		})
	}
}

func TestDaysForView(t *testing.T) {
	tests := []struct {
		view     string
		expected int
	}{
		{"1D", 1},
		{"WTD", 7},
		{"MTD", 31},
		{"YTD", 365},
		{"", 1},
		{"INVALID", 1},
	}

	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			days := daysForView(tt.view)
			if days != tt.expected {
				t.Errorf("daysForView(%s) = %d, want %d", tt.view, days, tt.expected)
			}
		})
	}
}

func TestGranularityForView(t *testing.T) {
	tests := []struct {
		view     string
		expected int
	}{
		{"1D", 300},
		{"WTD", 3600},
		{"MTD", 21600},
		{"YTD", 86400},
		{"", 300},
	}

	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			gran := granularityForView(tt.view)
			if gran != tt.expected {
				t.Errorf("granularityForView(%s) = %d, want %d", tt.view, gran, tt.expected)
			}
		})
	}
}

func TestDurationForView(t *testing.T) {
	tests := []struct {
		view     string
		expected time.Duration
	}{
		{"1D", 24 * time.Hour},
		{"WTD", 7 * 24 * time.Hour},
		{"MTD", 31 * 24 * time.Hour},
		{"YTD", 365 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			d := durationForView(tt.view)
			if d != tt.expected {
				t.Errorf("durationForView(%s) = %v, want %v", tt.view, d, tt.expected)
			}
		})
	}
}

type mockSuccessProvider struct {
	candles []Candle
}

func (m *mockSuccessProvider) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	return m.candles, nil
}

type mockFailProvider struct{}

func (m *mockFailProvider) FetchOHLC(symbol, view string, count int) ([]Candle, error) {
	return nil, fmt.Errorf("mock failure")
}

func TestClientFallback(t *testing.T) {
	client := NewClient()

	// Inject mock providers - one fails, one succeeds
	failProvider := &mockFailProvider{}
	successProvider := &mockSuccessProvider{
		candles: []Candle{
			{Time: time.Now(), Open: 100, High: 110, Low: 90, Close: 105},
		},
	}

	client.providers = []PriceProvider{failProvider, successProvider}

	candles, err := client.FetchOHLC("BTC", "1D", 60)
	if err != nil {
		t.Fatalf("FetchOHLC should succeed with fallback: %v", err)
	}

	if len(candles) != 1 {
		t.Errorf("expected 1 candle, got %d", len(candles))
	}
}

func TestClientAllProvidersFail(t *testing.T) {
	client := NewClient()
	client.providers = []PriceProvider{
		&mockFailProvider{},
		&mockFailProvider{},
	}

	_, err := client.FetchOHLC("BTC", "1D", 60)
	if err == nil {
		t.Error("expected error when all providers fail")
	}

	if !strings.Contains(err.Error(), "all providers failed") {
		t.Errorf("expected 'all providers failed' in error, got %v", err)
	}
}

// --- HTTP-backed provider tests ------------------------------------------------

func TestCoinGeckoProvider_HTTP(t *testing.T) {
    // create test server that mimics CoinGecko /coins/{id}/ohlc
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // successful path only
        w.WriteHeader(http.StatusOK)
        // return valid array plus an entry with too few fields to test skipping
        fmt.Fprint(w, `[[1630000000000,100,110,90,105],[1630000000001,1,2]]`)
    }))
    defer srv.Close()

    old := coingeckoBaseURL
    coingeckoBaseURL = srv.URL
    defer func() { coingeckoBaseURL = old }()

    prov := NewCoinGeckoProvider()
    prov.client = srv.Client()

    candles, err := prov.FetchOHLC("BTC", "1D", 60)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(candles) != 1 {
        t.Fatalf("expected 1 candle, got %d", len(candles))
    }
    if candles[0].Open != 100 {
        t.Errorf("unexpected open value: %v", candles[0].Open)
    }
}

func TestCoinGeckoProvider_HTTPErrors(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusTooManyRequests)
        fmt.Fprint(w, `rate limit`)
    }))
    defer srv.Close()

    old := coingeckoBaseURL
    coingeckoBaseURL = srv.URL
    defer func() { coingeckoBaseURL = old }()

    prov := NewCoinGeckoProvider()
    prov.client = srv.Client()

    if _, err := prov.FetchOHLC("BTC", "1D", 60); err == nil {
        t.Error("expected error for 429 status")
    }
}

func TestCoinGeckoProvider_MalformedJSON(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `not a json`)
    }))
    defer srv.Close()

    old := coingeckoBaseURL
    coingeckoBaseURL = srv.URL
    defer func() { coingeckoBaseURL = old }()

    prov := NewCoinGeckoProvider()
    prov.client = srv.Client()

    if _, err := prov.FetchOHLC("BTC", "1D", 60); err == nil {
        t.Error("expected json parse error")
    }
}

func TestCoinbaseProvider_HTTP(t *testing.T) {
    // simple single-page response with one candle
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        // matching [time, low, high, open, close, volume]
        fmt.Fprint(w, `[[1630000000,90,110,100,105,42]]`)
    }))
    defer srv.Close()

    old := coinbaseBaseURL
    coinbaseBaseURL = srv.URL
    defer func() { coinbaseBaseURL = old }()

    prov := NewCoinbaseProvider()
    prov.client = srv.Client()

    candles, err := prov.FetchOHLC("BTC", "1D", 60)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(candles) != 1 {
        t.Fatalf("expected 1 candle, got %d", len(candles))
    }
    if candles[0].Volume != 42 {
        t.Errorf("expected volume 42, got %v", candles[0].Volume)
    }
}

func TestCoinbaseProvider_HTTPErrors(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusTooManyRequests)
        fmt.Fprint(w, `wait`)
    }))
    defer srv.Close()

    old := coinbaseBaseURL
    coinbaseBaseURL = srv.URL
    defer func() { coinbaseBaseURL = old }()

    prov := NewCoinbaseProvider()
    prov.client = srv.Client()

    if _, err := prov.FetchOHLC("BTC", "1D", 60); err == nil {
        t.Error("expected error for 429 status")
    }
}

func TestCoinbaseProvider_MalformedJSON(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `[]junk[]`)
    }))
    defer srv.Close()

    old := coinbaseBaseURL
    coinbaseBaseURL = srv.URL
    defer func() { coinbaseBaseURL = old }()

    prov := NewCoinbaseProvider()
    prov.client = srv.Client()

    if _, err := prov.FetchOHLC("BTC", "1D", 60); err == nil {
        t.Error("expected json parse error")
    }
}

