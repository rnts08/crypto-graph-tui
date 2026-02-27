package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the persistent application state.
type Config struct {
	Symbols         []string `json:"symbols"`
	View            string   `json:"view"`
	RefreshInterval int      `json:"refresh_interval_secs"`
	APIKey          string   `json:"api_key,omitempty"`
	Theme           string   `json:"theme,omitempty"`
	WindowWidth     int      `json:"window_width,omitempty"`
	WindowHeight    int      `json:"window_height,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Symbols:         []string{"BTC-USD", "ETH-USD", "LTC-USD", "SOL-USD"},
		View:            "1D",
		RefreshInterval: 60,
		APIKey:          "",
		Theme:           "default",
		WindowWidth:     0,
		WindowHeight:    0,
	}
}

// ConfigPath returns the path to the config file.
// It prefers ${XDG_CONFIG_HOME}/graph-watcher/config.json,
// falling back to ~/.graph-watcher.json.
func ConfigPath() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "graph-watcher", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".graph-watcher.json"), nil
}

// LoadConfig reads the configuration from disk.
// If the file does not exist, it returns the default config.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Ensure defaults for empty fields.
	if len(cfg.Symbols) == 0 {
		cfg.Symbols = DefaultConfig().Symbols
	}
	// Sanitize any persisted symbols (old config may contain junk/flags).
	cfg.Symbols = sanitizeSymbols(cfg.Symbols)

	if cfg.View == "" {
		cfg.View = "1D"
	}
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = 60
	}
	if cfg.Theme == "" {
		cfg.Theme = DefaultConfig().Theme
	}
	// keep window dimensions if present
	// zero means unknown/not yet set

	return &cfg, nil
}

// SaveConfig writes the configuration to disk.
func SaveConfig(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
