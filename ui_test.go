package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func makeTestModel() *Model {
	cfg := &Config{Symbols: defaultSymbols(), View: "1D", Theme: "light"}
	client := &Client{} // not used in this test
	m := NewModel(cfg, client)
	return m
}

func TestCustomSymbolEntry(t *testing.T) {
	m := makeTestModel()
	var res tea.Model
	// start in main
	if m.mode != ModeMain {
		t.Fatalf("expected initial mode main, got %v", m.mode)
	}

	// simulate pressing 'a' to open add-symbol modal
	res, _ = m.handleMainKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = res.(*Model)
	if m.mode != ModeAddSymbol {
		t.Fatalf("expected mode addSymbol after pressing a, got %v", m.mode)
	}

	// press 'c' to start custom entry
	res, _ = m.handleAddSymbolKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = res.(*Model)
	if m.mode != ModeCustomSymbol {
		t.Fatalf("expected mode customSymbol after pressing c, got %v", m.mode)
	}

	// type unique token ABC and press Enter (not in defaults)
	example := "ABC"
	for _, r := range example {
		res, _ = m.handleCustomKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = res.(*Model)
	}
	// press enter
	res, _ = m.handleCustomKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(*Model)

	// should return to main
	if m.mode != ModeMain {
		t.Fatalf("expected mode main after enter, got %v", m.mode)
	}
	// symbol list should contain ABC-USD
	found := false
	for _, s := range m.symbols {
		if s == "ABC-USD" {
			found = true
		}
	}
	if !found {
		t.Error("expected ABC-USD to be added to symbols")
	}

	// config should have same symbol
	if len(m.config.Symbols) == 0 || m.config.Symbols[len(m.config.Symbols)-1] != "ABC-USD" {
		t.Error("config not updated with custom symbol")
	}
}

func TestCustomEntryCancel(t *testing.T) {
	m := makeTestModel()
	var res tea.Model
	m.mode = ModeCustomSymbol
	m.customInput = "XXX"
	// press esc
	res, _ = m.handleCustomKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(*Model)
	if m.mode != ModeMain {
		t.Errorf("expected mode main after esc, got %v", m.mode)
	}
	// symbol list should not contain the cancelled symbol
	for _, s := range m.symbols {
		if s == "XXX-USD" {
			t.Error("should not add symbol when cancelled")
		}
	}
}
