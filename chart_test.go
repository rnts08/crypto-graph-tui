package main

import (
	"testing"
	"time"
)

func TestGridForCount(t *testing.T) {
	tests := []struct {
		n         int
		wantRows  int
		wantCols  int
		desc      string
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
	result := RenderCandlesASCII("BTC-USD", candles, 5, 10, "default")
	if result != "BTC-USD: not enough space or data" {
		t.Errorf("expected error message for small width, got %s", result)
	}

	// Height too small
	result = RenderCandlesASCII("BTC-USD", candles, 20, 2, "default")
	if result != "BTC-USD: not enough space or data" {
		t.Errorf("expected error message for small height, got %s", result)
	}

	// No candles
	result = RenderCandlesASCII("BTC-USD", []Candle{}, 20, 10, "default")
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

	result := RenderCandlesASCII("BTC-USD", candles, 20, 10, "default")
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

	result := RenderCandlesASCII("BTC-USD", candles, 30, 15, "default")

	// Check that it's not empty and contains the symbol
	if result == "" {
		t.Fatal("RenderCandlesASCII returned empty string")
	}
	if !contains(result, "BTC-USD") {
		t.Errorf("expected symbol BTC-USD in result")
	}

	// Check that it contains the price
	if !contains(result, "118.00") {
		t.Errorf("expected latest close price 118.00 in result")
	}

	// Check borders exist
	if !contains(result, "+") || !contains(result, "-") || !contains(result, "|") {
		t.Error("expected chart borders in result")
	}
}

func TestRenderCandlesASCII_ColorTags(t *testing.T) {
	now := time.Now()
	candles := []Candle{
		{Time: now, Open: 100, High: 110, Low: 90, Close: 105},
	}

	result := RenderCandlesASCII("BTC-USD", candles, 30, 15, "default")

	// Should contain block characters for candle bodies
	if !contains(result, "█") {
		t.Errorf("expected block characters in result")
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

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
