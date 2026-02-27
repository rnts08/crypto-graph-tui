package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Symbols) == 0 {
		t.Error("DefaultConfig should have symbols")
	}
	if cfg.View != "1D" {
		t.Errorf("expected View=1D, got %s", cfg.View)
	}
	if cfg.RefreshInterval != 60 {
		t.Errorf("expected RefreshInterval=60, got %d", cfg.RefreshInterval)
	}
}

func TestLoadConfigDefault(t *testing.T) {
	// Create a temp directory and override XDG_CONFIG_HOME
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Loading should return default config when file doesn't exist
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(cfg.Symbols) == 0 {
		t.Error("LoadConfig should return default symbols when file missing")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a custom config
	cfg := &Config{
		Symbols:         []string{"BTC-USD", "ETH-USD"},
		View:            "WTD",
		RefreshInterval: 120,
		APIKey:          "test-key",
	}

	// Save it
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load it back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	if len(loaded.Symbols) != 2 || loaded.Symbols[0] != "BTC-USD" {
		t.Errorf("symbols mismatch: %v", loaded.Symbols)
	}
	if loaded.View != "WTD" {
		t.Errorf("expected View=WTD, got %s", loaded.View)
	}
	if loaded.RefreshInterval != 120 {
		t.Errorf("expected RefreshInterval=120, got %d", loaded.RefreshInterval)
	}
	if loaded.APIKey != "test-key" {
		t.Errorf("APIKey mismatch: %s", loaded.APIKey)
	}
}

func TestConfigPath(t *testing.T) {
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Test with XDG_CONFIG_HOME set
	os.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}
	expected := filepath.Join("/tmp/xdg", "graph-watcher", "config.json")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}

	// Test without XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", "")
	path, err = ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("path should be absolute, got %s", path)
	}
	if !strings.HasSuffix(path, ".graph-watcher.json") {
		t.Errorf("expected path to end with .graph-watcher.json, got %s", path)
	}
}

func TestLoadConfigMergeDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Save a config with empty/zero values
	cfg := &Config{
		Symbols: nil,
		View:    "",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load should merge with defaults
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(loaded.Symbols) == 0 {
		t.Error("LoadConfig should merge empty symbols with defaults")
	}
	if loaded.View != "1D" {
		t.Errorf("LoadConfig should merge empty view with default, got %s", loaded.View)
	}
	if loaded.RefreshInterval != 60 {
		t.Errorf("LoadConfig should have default RefreshInterval, got %d", loaded.RefreshInterval)
	}
}

// Ensure that invalid or flag-like symbols stored in config are sanitized on load.
func TestLoadConfigSanitizesInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Save a config containing a bogus entry
	cfg := &Config{
		Symbols: []string{"--HELP", "BTC"},
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load should drop the flag-like entry and normalize the good one
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(loaded.Symbols) != 1 || loaded.Symbols[0] != "BTC-USD" {
		t.Errorf("unexpected symbols after sanitization: %v", loaded.Symbols)
	}
}
