//go:build !windows

package main

import (
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// suspendCmd returns a command that suspends the process to the shell using SIGTSTP.
// The terminal will return to the shell on suspend and this process will be stopped.
// Upon resuming (SIGCONT) bubble tea should continue with the model intact.
func suspendCmd() tea.Cmd {
	return func() tea.Msg {
		// send SIGTSTP to ourselves
		syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)
		return nil
	}
}
