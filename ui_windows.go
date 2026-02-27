//go:build windows

package main

import tea "github.com/charmbracelet/bubbletea"

// suspendCmd is a no-op on Windows.
func suspendCmd() tea.Cmd {
	return nil
}
