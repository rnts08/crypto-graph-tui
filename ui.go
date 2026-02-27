package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the application state
// UIMode indicates which screen is active
// (declared before constants for readability)
type UIMode string

type Model struct {
	// UI state
	width           int
	height          int
	mode            UIMode // main, addSymbol, confirmQuit
	selectedSymbols map[string]bool
	symbolList      []string

	// Data
	client  *Client
	charts  map[string]*ChartData
	view    string
	config  *Config
	symbols []string

	// Refresh timing
	nextRefresh time.Time

	// Navigation (for symbol list)
	symbolListIndex int
	scrollOffset    int

	// custom entry buffer
	customInput string
}

const (
	ModeMain         UIMode = "main"
	ModeAddSymbol    UIMode = "addSymbol"
	ModeConfirmQuit  UIMode = "confirmQuit"
	ModeCustomSymbol UIMode = "customSymbol"
)

type ChartData struct {
	Symbol  string
	Candles []Candle
	Error   string
	Updated time.Time
}

// Messages for bubbletea framework
type tickMsg time.Time
type refreshMsg struct {
	symbol string
	data   []Candle
	err    error
}

// NewModel creates a new application model
func NewModel(cfg *Config, client *Client) *Model {
	return &Model{
		client:          client,
		charts:          make(map[string]*ChartData),
		view:            cfg.View,
		config:          cfg,
		symbols:         cfg.Symbols,
		selectedSymbols: make(map[string]bool),
		symbolList: []string{
			"BTC-USD", "ETH-USD", "SOL-USD", "LTC-USD", "ADA-USD", "AVAX-USD",
			"MATIC-USD", "DOGE-USD", "XRP-USD", "XMR-USD", "PEPE-USD", "UNI-USD",
			"USDT-USD", "BNB-USD", "BCH-USD", "ZBCN-USD", "BABYDOGE-USD",
			"NOBODY-USD", "SUI-USD", "HYPE-USD", "AZERO-USD", "LINK-USD",
			"TRUMP-USD", "TOWNS-USD", "SNX-USD", "MON-USD",
		},
		mode:        ModeMain,
		nextRefresh: time.Now().Add(time.Duration(cfg.RefreshInterval) * time.Second),
	}
}

// Init initializes the model and starts background loops
func (m *Model) Init() tea.Cmd {
	// Initialize charts for all symbols
	for _, sym := range m.symbols {
		m.charts[sym] = &ChartData{Symbol: sym}
	}

	// Return batch of commands
	return tea.Batch(
		m.fetchAllChartsCmd(),
		m.tickCmd(),
	)
}

// Update handles all events and state changes
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Check if it's time to refresh data
		if time.Now().After(m.nextRefresh) {
			m.nextRefresh = time.Now().Add(time.Duration(m.config.RefreshInterval) * time.Second)
			return m, tea.Batch(m.fetchAllChartsCmd(), m.tickCmd())
		}
		return m, m.tickCmd()

	case refreshMsg:
		if msg.err == nil && len(msg.data) > 0 {
			m.charts[msg.symbol].Candles = msg.data
			m.charts[msg.symbol].Error = ""
		} else if msg.err != nil {
			m.charts[msg.symbol].Error = msg.err.Error()
		}
		m.charts[msg.symbol].Updated = time.Now()
		return m, nil

	case tea.QuitMsg:
		return m, tea.Quit
	}

	return m, nil
}

// handleKeyPress processes all keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeMain:
		return m.handleMainKeys(msg)
	case ModeAddSymbol:
		return m.handleAddSymbolKeys(msg)
	case ModeCustomSymbol:
		return m.handleCustomKeys(msg)
	case ModeConfirmQuit:
		return m.handleQuitKeys(msg)
	}
	return m, nil
}

// handleMainKeys handles keys in main view
func (m *Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "ctrl+z":
		// suspend to shell
		return m, suspendCmd()

	case "q", "Q":
		m.mode = ModeConfirmQuit
		return m, nil

	case "a", "A":
		m.mode = ModeAddSymbol
		m.symbolListIndex = 0
		m.scrollOffset = 0
		// Initialize selected states
		for _, sym := range m.symbolList {
			m.selectedSymbols[sym] = m.hasSymbol(sym)
		}
		return m, nil

	case "1":
		m.view = "1D"
		return m, m.fetchAllChartsCmd()

	case "2":
		m.view = "WTD"
		return m, m.fetchAllChartsCmd()

	case "3":
		m.view = "MTD"
		return m, m.fetchAllChartsCmd()

	case "4":
		m.view = "YTD"
		return m, m.fetchAllChartsCmd()
	}

	return m, nil
}

// handleAddSymbolKeys processes keys in symbol selection modal
func (m *Model) handleAddSymbolKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "c" || msg.String() == "C" {
		m.mode = ModeCustomSymbol
		m.customInput = ""
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.mode = ModeMain
		m.selectedSymbols = make(map[string]bool)
		return m, nil

	case "up", "k":
		if m.symbolListIndex > 0 {
			m.symbolListIndex--
			if m.symbolListIndex < m.scrollOffset {
				m.scrollOffset = m.symbolListIndex
			}
		}
		return m, nil

	case "down", "j":
		if m.symbolListIndex < len(m.symbolList)-1 {
			m.symbolListIndex++
			if m.symbolListIndex >= m.scrollOffset+10 {
				m.scrollOffset = m.symbolListIndex - 9
			}
		}
		return m, nil

	case " ":
		sym := m.symbolList[m.symbolListIndex]
		m.selectedSymbols[sym] = !m.selectedSymbols[sym]
		return m, nil

	case "enter":
		var newSymbols []string
		for _, sym := range m.symbolList {
			if m.selectedSymbols[sym] {
				newSymbols = append(newSymbols, sym)
			}
		}

		if len(newSymbols) > 0 {
			m.symbols = newSymbols
			m.config.Symbols = newSymbols
			SaveConfig(m.config)

			newCharts := make(map[string]*ChartData)
			for _, sym := range newSymbols {
				if existing, ok := m.charts[sym]; ok {
					newCharts[sym] = existing
				} else {
					newCharts[sym] = &ChartData{Symbol: sym}
				}
			}
			m.charts = newCharts
		}

		m.mode = ModeMain
		m.selectedSymbols = make(map[string]bool)
		return m, m.fetchAllChartsCmd()

	case "ctrl+c":
		return m, tea.Quit

	case "q", "Q":
		m.mode = ModeConfirmQuit
		return m, nil
	}

	return m, nil
}

// handleCustomKeys processes keys while entering custom symbol text
func (m *Model) handleCustomKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		m.customInput += msg.String()
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.customInput) > 0 {
			m.customInput = m.customInput[:len(m.customInput)-1]
		}
	case tea.KeyEnter:
		if s := strings.TrimSpace(m.customInput); s != "" {
			symbol := strings.ToUpper(s)
			if !strings.Contains(symbol, "-") {
				symbol += "-USD"
			}
			if !m.hasSymbol(symbol) {
				m.symbols = append(m.symbols, symbol)
				m.config.Symbols = m.symbols
				SaveConfig(m.config)
				m.charts[symbol] = &ChartData{Symbol: symbol}
			}
		}
		m.mode = ModeMain
	case tea.KeyEsc:
		m.mode = ModeMain
	}
	return m, nil
}

// handleQuitKeys processes keys in quit confirmation
func (m *Model) handleQuitKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		SaveConfig(m.config)
		return m, tea.Quit

	case "n", "N", "esc":
		m.mode = ModeMain
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// Stop terminates the program. Added for compatibility with older API
// and to provide a place for cleanup if channels are introduced later.
func (m *Model) Stop() tea.Cmd {
	return tea.Quit
}

// View renders the current state
func (m *Model) View() string {
	switch m.mode {
	case ModeMain:
		return m.viewMain()
	case ModeAddSymbol:
		return m.viewAddSymbol()
	case ModeConfirmQuit:
		return m.viewConfirmQuit()
	}
	return ""
}

// viewMain renders the main chart view
func (m *Model) viewMain() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if len(m.symbols) == 0 {
		return "No symbols selected. Press A to add some."
	}

	nextIn := time.Until(m.nextRefresh)
	if nextIn < 0 {
		nextIn = 0
	}
	mins := int(nextIn.Seconds()) / 60
	secs := int(nextIn.Seconds()) % 60
	statusBar := fmt.Sprintf("v%s | Refresh in %02d:%02d | View: %s | A=add 1-4=timeframe Q=quit",
		version, mins, secs, m.view)

	// Use grid layout (balanced rows/cols via shared utility)
	r, c := GridForCount(len(m.symbols))
	cols := c
	rows := r
	chartHeight := (m.height - 3) / rows

	var result strings.Builder

	for row := 0; row < rows; row++ {
		// Render row of charts
		var rowLines [][]string
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx >= len(m.symbols) {
				break
			}

			sym := m.symbols[idx]
			chart, ok := m.charts[sym]
			if !ok {
				chart = &ChartData{Symbol: sym}
				m.charts[sym] = chart
			}

			var content string
			if chart.Error != "" {
				content = fmt.Sprintf("[%s] ERROR: %s", sym, chart.Error)
			} else if len(chart.Candles) == 0 {
				content = fmt.Sprintf("[%s] Loading...", sym)
			} else {
				chartWidth := (m.width / cols) - 1
				content = RenderCandlesASCII(sym, chart.Candles, chartWidth, chartHeight, m.config.Theme)
			}

			lines := strings.Split(content, "\n")
			if len(lines) > chartHeight {
				lines = lines[:chartHeight]
			}
			// Pad to chartHeight
			for len(lines) < chartHeight {
				lines = append(lines, "")
			}
			rowLines = append(rowLines, lines)
		}

		// Merge lines horizontally
		for lineIdx := 0; lineIdx < chartHeight; lineIdx++ {
			for colIdx := 0; colIdx < len(rowLines); colIdx++ {
				if lineIdx < len(rowLines[colIdx]) {
					result.WriteString(rowLines[colIdx][lineIdx])
				}
				// Add spacing between columns
				if colIdx < len(rowLines)-1 {
					result.WriteString(" ")
				}
			}
			result.WriteString("\n")
		}
	}

	result.WriteString(statusBar)
	return result.String()
}

// viewAddSymbol renders the symbol selection modal
func (m *Model) viewAddSymbol() string {
	// if entering custom symbol show only prompt
	if m.mode == ModeCustomSymbol {
		return fmt.Sprintf("\n\n  ╔══════════════════════════════════════╗\n  ║  ADD CUSTOM SYMBOL                   ║\n  ║  Enter ticker (e.g. BTC): %-12s║\n  ╚══════════════════════════════════════╝\n", m.customInput)
	}
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  ╔══════════════════════════════════════╗")
	lines = append(lines, "  ║  SELECT SYMBOLS                        ║")
	lines = append(lines, "  ║  ↑↓ navigate, SPACE toggle, ENTER ok  ║")
	lines = append(lines, "  ╚══════════════════════════════════════╝")
	lines = append(lines, "")

	start := m.scrollOffset
	end := start + 10
	if end > len(m.symbolList) {
		end = len(m.symbolList)
	}

	for i := start; i < end; i++ {
		sym := m.symbolList[i]
		prefix := "  "
		if i == m.symbolListIndex {
			prefix = "▶ "
		}

		checked := " "
		if m.selectedSymbols[sym] {
			checked = "✓"
		}

		line := fmt.Sprintf("%s  [%s] %s", prefix, checked, sym)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, "  ESC=cancel  ENTER=apply  Ctrl+C=quit")
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// viewConfirmQuit renders the quit confirmation
func (m *Model) viewConfirmQuit() string {
	return `

  ╔════════════════════════════════════════╗
  ║                                        ║
  ║      Quit graph-watcher?               ║
  ║                                        ║
  ║      (Y)es or (N)o                     ║
  ║      Ctrl+C to quit immediately       ║
  ║                                        ║
  ╚════════════════════════════════════════╝
`
}

// Commands
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// concurrency semaphore for fetching symbols.
var fetchSem = make(chan struct{}, 2) // limit to 2 concurrent requests

func (m *Model) fetchChartCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		// respect global concurrency limit
		fetchSem <- struct{}{}
		candles, err := m.client.FetchOHLC(symbol, m.view, 60)
		<-fetchSem
		return refreshMsg{symbol: symbol, data: candles, err: err}
	}
}

func (m *Model) fetchAllChartsCmd() tea.Cmd {
	var cmds []tea.Cmd
	for _, sym := range m.symbols {
		cmds = append(cmds, m.fetchChartCmd(sym))
	}
	return tea.Batch(cmds...)
}

func (m *Model) hasSymbol(symbol string) bool {
	for _, s := range m.symbols {
		if strings.EqualFold(s, symbol) {
			return true
		}
	}
	return false
}

// GridForCountCols returns the number of columns for a count.
// It now delegates to the shared GridForCount helper so the UI
// layout matches the chart mathematics and remains balanced.
func GridForCountCols(n int) int {
	if n <= 0 {
		return 0
	}
	_, cols := GridForCount(n)
	return cols
}
