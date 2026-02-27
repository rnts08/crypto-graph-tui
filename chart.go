package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// colorForTheme returns green/red hex codes for a given theme name.
func colorForTheme(theme string) (green, red string) {
	switch theme {
	case "dark":
		return "#00FF00", "#FF0000"
	case "retro":
		return "#A6E22E", "#F92672"
	default: // default
		return "#00C853", "#FF3D00"
	}
}

// RenderCandlesASCII renders a slice of candles as an ASCII candlestick chart.
// It creates a bordered chart with wicks and bodies, scaled to fit the given width and height.
func RenderCandlesASCII(symbol string, candles []Candle, width, height int, theme string) string {
	if width < 10 || height < 3 || len(candles) == 0 {
		return fmt.Sprintf("%s: not enough space or data", symbol)
	}
	// Limit candles to width-2 for left/right padding.
	maxCandles := width - 2
	if len(candles) > maxCandles {
		candles = candles[len(candles)-maxCandles:]
	}

	var minPrice, maxPrice float64
	minPrice = math.MaxFloat64
	maxPrice = -math.MaxFloat64
	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
	}
	if minPrice == math.MaxFloat64 || maxPrice == -math.MaxFloat64 || maxPrice <= minPrice {
		return fmt.Sprintf("%s: no price range", symbol)
	}

	// Use the latest close as "current price".
	lastClose := candles[len(candles)-1].Close

	startTime := candles[0].Time
	endTime := candles[len(candles)-1].Time

	scale := float64(height-2) / (maxPrice - minPrice)

	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			grid[y][x] = ' '
		}
	}

	// Draw border.
	for x := 0; x < width; x++ {
		grid[0][x] = '-'
		grid[height-1][x] = '-'
	}
	for y := 0; y < height; y++ {
		grid[y][0] = '|'
		grid[y][width-1] = '|'
	}
	grid[0][0], grid[0][width-1] = '+', '+'
	grid[height-1][0], grid[height-1][width-1] = '+', '+'

	// Draw candles.
	offsetX := 1
	for i, c := range candles {
		x := offsetX + i
		if x >= width-1 {
			break
		}

		highY := priceToY(c.High, minPrice, scale, height)
		lowY := priceToY(c.Low, minPrice, scale, height)
		openY := priceToY(c.Open, minPrice, scale, height)
		closeY := priceToY(c.Close, minPrice, scale, height)

		// Wick
		for y := highY; y <= lowY; y++ {
			if y <= 0 || y >= height-1 {
				continue
			}
			grid[y][x] = '|'
		}

		// Body: use marker runes we post-process into colored blocks.
		bodyChar := '▲' // up candle (green)
		if c.Close < c.Open {
			bodyChar = '▼' // down candle (red)
		}
		top := openY
		bottom := closeY
		if closeY < openY {
			top, bottom = closeY, openY
		}
		if top == bottom {
			if top > 0 && top < height-1 {
				grid[top][x] = bodyChar
			}
		} else {
			for y := top; y <= bottom; y++ {
				if y <= 0 || y >= height-1 {
					continue
				}
				grid[y][x] = bodyChar
			}
		}
	}

	lines := make([]string, height)
	// Prepare lipgloss styles for up/down candles based on theme.
	greenColor, redColor := colorForTheme(theme)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color(redColor))

	for i := range grid {
		line := string(grid[i])
		// Replace marker runes with colored blocks using lipgloss styles.
		line = strings.ReplaceAll(line, "▲", green.Render("█"))
		line = strings.ReplaceAll(line, "▼", red.Render("█"))
		lines[i] = line
	}

	timeInfo := fmt.Sprintf("%s – %s",
		startTime.Format("Jan 02 15:04"),
		endTime.Format("Jan 02 15:04"),
	)

	// Format volume with K/M/B suffixes
	lastVol := candles[len(candles)-1].Volume
	volStr := formatVolume(lastVol)

	title := fmt.Sprintf(" %s $%.2f (%.2f - %.2f) Vol: %s [%s] ", symbol, lastClose, minPrice, maxPrice, volStr, timeInfo)
	if len(title) > width-2 {
		title = title[:width-3] + "…"
	}
	runes := []rune(lines[0])
	copy(runes[1:], []rune(title))
	lines[0] = string(runes)

	return strings.Join(lines, "\n")
}

// formatVolume formats a float64 volume into a human-readable string.
func formatVolume(vol float64) string {
	if vol >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", vol/1_000_000_000)
	}
	if vol >= 1_000_000 {
		return fmt.Sprintf("%.1fM", vol/1_000_000)
	}
	if vol >= 1_000 {
		return fmt.Sprintf("%.1fK", vol/1_000)
	}
	return fmt.Sprintf("%.1f", vol)
}

// priceToY converts a price to a Y coordinate on the chart.
func priceToY(price, minPrice, scale float64, height int) int {
	v := (price - minPrice) * scale
	y := (height - 2) - int(math.Round(v)) // invert so higher prices are near top
	if y < 1 {
		return 1
	}
	if y > height-2 {
		return height - 2
	}
	return y
}

// GridForCount computes the rows and columns of a grid to fit n items.
func GridForCount(n int) (rows, cols int) {
	if n <= 0 {
		return 0, 0
	}
	cols = mathCeilSqrt(n)
	rows = (n + cols - 1) / cols
	return
}

// mathCeilSqrt computes the ceiling of the square root of n.
func mathCeilSqrt(n int) int {
	if n <= 1 {
		return 1
	}
	x := 1
	for x*x < n {
		x++
	}
	return x
}
