package main

import (
	"time"
)

// Candle represents OHLCV data for a single time period.
type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}
