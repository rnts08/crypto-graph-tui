package main

import (
	"strings"
	"testing"
	"time"
)

func TestGridForCount(t *testing.T) {
	tests := []struct {
		n        int
		wantRows int
		wantCols int
		desc     string
	}{
		{1, 1, 1, "single item"},
		{2, 1, 2, "two items"},
		{3, 2, 2, "three items (2x2 grid)"},
		{4, 2, 2, "four items"},
		{5, 2, 3, "five items (2x3 grid)"},
		{6, 2, 3, "six items"},
		{7, 3, 3, "seven items (3x3 grid)"},
		{8, 3, 3, "eight items"},
		{9, 3, 3, "nine items"},
		{10, 3, 4, "ten items (3x4 grid)"},
		{0, 0, 0, "zero items"},
		{-1, 0, 0, "negative items"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rows, cols := GridForCount(tt.n)
			if rows != tt.wantRows || cols != tt.wantCols {
				t.Errorf("GridForCount(%d) = (%d, %d), want (%d, %d)",
					tt.n, rows, cols, tt.wantRows, tt.wantCols)
			}
		})
	}
}

func TestRenderCandlesASCII_TooSmall(t *testing.T) {
	candles := []Candle{
		{Time: time.Now(), Open: 100, High: 110, Low: 90, Close: 105},
	}

	// Width too small
	result := RenderCandlesASCII("BTC-USD", candles, 5, 10, "default", "1D", false, false)
	if result != "BTC-USD: not enough space or data" {
		t.Errorf("expected error message for small width, got %s", result)
	}

	// Height too small
	result = RenderCandlesASCII("BTC-USD", candles, 20, 2, "default", "1D", false, false)
	if result != "BTC-USD: not enough space or data" {
		t.Errorf("expected error message for small height, got %s", result)
	}

	// No candles
	result = RenderCandlesASCII("BTC-USD", []Candle{}, 20, 10, "default", "1D", false, false)
	if result != "BTC-USD: not enough space or data" {
		t.Errorf("expected error message for empty candles, got %s", result)
	}
}

func TestRenderCandlesASCII_NoRange(t *testing.T) {
	// All candles have the same price (no range)
	candles := []Candle{
		{Time: time.Now(), Open: 100, High: 100, Low: 100, Close: 100},
		{Time: time.Now().Add(time.Minute), Open: 100, High: 100, Low: 100, Close: 100},
	}

	result := RenderCandlesASCII("BTC-USD", candles, 20, 10, "default", "1D", false, false)
	if result != "BTC-USD: no price range" {
		t.Errorf("expected no range error, got %s", result)
	}
}

func TestRenderCandlesASCII_Valid(t *testing.T) {
	now := time.Now()
	candles := []Candle{
		{Time: now, Open: 100, High: 110, Low: 90, Close: 105},
		{Time: now.Add(time.Minute), Open: 105, High: 115, Low: 100, Close: 110},
		{Time: now.Add(2 * time.Minute), Open: 110, High: 120, Low: 105, Close: 118},
	}

	result := RenderCandlesASCII("BTC-USD", candles, 30, 15, "default", "1D", false, false)

	// Check that it's not empty and contains the symbol
	if result == "" {
		t.Fatal("RenderCandlesASCII returned empty string")
	}
	if !strings.Contains(result, "BTC-USD") {
		t.Errorf("expected symbol BTC-USD in result")
	}

	// Check that it contains the price
	if !strings.Contains(result, "118.00") {
		t.Errorf("expected latest close price 118.00 in result")
	}

	// Check borders exist
	if !strings.Contains(result, "+") || !strings.Contains(result, "-") || !strings.Contains(result, "|") {
		t.Error("expected chart borders in result")
	}
}

func TestRenderCandlesASCII_ColorTags(t *testing.T) {
	now := time.Now()
	candles := []Candle{
		{Time: now, Open: 100, High: 110, Low: 90, Close: 105},
	}

	result := RenderCandlesASCII("BTC-USD", candles, 30, 15, "default", "1D", false, false)

	// Should contain block characters for candle bodies
	if !strings.Contains(result, "█") {
		t.Errorf("expected block characters in result")
	}
}

func TestRenderCandlesASCII_ZoomedTitle(t *testing.T) {
	now := time.Now()
	candles := []Candle{
		{Time: now.Add(-1 * time.Hour), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Time: now, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1200},
	}

	// Render zoomed
	result := RenderCandlesASCII("BTC-USD", candles, 160, 15, "default", "1D", true, true)

	// Check for zoomed title components
	if !strings.Contains(result, "Last: $110.00") {
		t.Errorf("zoomed title missing last price")
	}
	if !strings.Contains(result, "O: 105.00 H: 115.00 L: 100.00 C: 110.00") {
		t.Errorf("zoomed title missing OHLC")
	}
	if !strings.Contains(result, "Chg: +10.00 (+10.00%)") {
		t.Errorf("zoomed title missing change info, got %s", result)
	}

	// Render not zoomed
	result = RenderCandlesASCII("BTC-USD", candles, 80, 15, "default", "1D", false, false)
	if strings.Contains(result, "Last: $110.00") {
		t.Errorf("non-zoomed title should not contain detailed info")
	}
}

func TestPriceToY(t *testing.T) {
	// Test boundary conditions
	minPrice := 100.0
	maxPrice := 110.0
	height := 10
	scale := float64(height-2) / (maxPrice - minPrice)

	// Minimum price should map to bottom (height-2)
	y := priceToY(minPrice, minPrice, scale, height)
	if y != height-2 {
		t.Errorf("priceToY for min price should be %d, got %d", height-2, y)
	}

	// Maximum price should map to top (1)
	y = priceToY(maxPrice, minPrice, scale, height)
	if y != 1 {
		t.Errorf("priceToY for max price should be 1, got %d", y)
	}

	// Mid price should map to mid
	midPrice := (minPrice + maxPrice) / 2
	y = priceToY(midPrice, minPrice, scale, height)
	if y < 1 || y > height-2 {
		t.Errorf("priceToY for mid price out of bounds: %d", y)
	}
}
