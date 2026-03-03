package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.1.1"

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Command-line flags (symbols, api-key, interval)
	var flagSymbols string
	var flagAPIKey string
	var flagInterval int
	var flagTheme string
	flag.StringVar(&flagSymbols, "symbols", "", "comma-separated list of symbols (e.g. BTC-USD,ETH-USD)")
	flag.StringVar(&flagAPIKey, "api-key", "", "API key for CoinMarketCap or other provider")
	flag.IntVar(&flagInterval, "interval", 0, "refresh interval in seconds (overrides config)")
	flag.StringVar(&flagTheme, "theme", "", "color theme (default, dark, retro)")
	flag.Parse()

	// Sanitize the loaded symbols first (in case the config was dirty)
	symbols := sanitizeSymbols(cfg.Symbols)

	if flagSymbols != "" {
		symbols = parseSymbols(flagSymbols)
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}
	if flagInterval > 0 {
		cfg.RefreshInterval = flagInterval
	}
	if flagTheme != "" {
		cfg.Theme = flagTheme
	}
	cfg.Symbols = symbols

	// Create client
	client := NewClient()

	// Create and run model
	m := NewModel(cfg, client)
	// restore previous window size if available (will be overwritten by WindowSizeMsg)
	if cfg.WindowWidth > 0 && cfg.WindowHeight > 0 {
		m.width = cfg.WindowWidth
		m.height = cfg.WindowHeight
	}

	// Setup signal handlers
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	p := tea.NewProgram(m, tea.WithAltScreen())

	// Handle signals in background
	go func() {
		sig := <-sigCh
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			// Save and quit
			cfg.Symbols = m.symbols
			SaveConfig(cfg)
			p.Quit()
		}
	}()

	_, err = p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Save config on exit
	cfg.Symbols = m.symbols
	cfg.WindowWidth = m.width
	cfg.WindowHeight = m.height
	if err := SaveConfig(cfg); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

// defaultSymbols returns the hardcoded starting symbols.
func defaultSymbols() []string {
	return []string{"BTC-USD", "ETH-USD", "LTC-USD", "SOL-USD"}
}

// sanitizeSymbols normalizes and filters a list of raw symbol strings.
// It uppercases, adds "-USD" if missing, and drops entries that look like
// flags or are empty. If the resulting slice is empty, it returns defaults.
func sanitizeSymbols(raw []string) []string {
	var out []string
	for _, p := range raw {
		s := strings.TrimSpace(strings.ToUpper(p))
		if s == "" || strings.HasPrefix(s, "-") {
			continue
		}
		if !strings.Contains(s, "-") {
			s = s + "-USD"
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return defaultSymbols()
	}
	return out
}

// parseSymbols parses a comma-separated list of symbols.  If the argument
// appears to be a flag (starts with '-'), it returns the default set instead.
func parseSymbols(arg string) []string {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return defaultSymbols()
	}
	// treat something like "--help" as no symbols
	if strings.HasPrefix(arg, "-") {
		return defaultSymbols()
	}
	parts := strings.Split(arg, ",")
	return sanitizeSymbols(parts)
}
