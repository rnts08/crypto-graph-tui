# Detailed Implementation TODO

This list breaks down every required feature into discrete, actionable items. Each checkbox represents a task or sub‑task that will guide development and track progress.

## Configuration

- [x] Design `Config` struct (symbols, view, refresh interval, API key, etc.)
- [x] Implement `LoadConfig()` reading JSON from `${XDG_CONFIG_HOME}/graph-watcher/config.json` (fallback to `~/.graph-watcher.json`).
- [x] Implement `SaveConfig()` writing current state.
- [x] Add command‑line argument parsing to override symbols.
- [x] Merge CLI args and config on startup, with CLI taking precedence.
- [x] Invoke `SaveConfig()` whenever symbols/view/interval change or on graceful exit.

## Data provider layer

- [x] Define `PriceProvider` interface with `FetchOHLC(symbol, view string, count int) ([]Candle, error)`.
- [x] Implement CoinGecko provider:
  - [ ] Map symbols to CG IDs.
  - [ ] Call `/coins/{id}/ohlc` with `vs_currency=usd`.
  - [ ] Decode and convert to `Candle` slice.
- [x] Implement Coinbase provider:
  - [ ] Normalize symbol to `BTC-USD` format.
  - [ ] Iterate time windows to respect 300‑row API limit.
  - [ ] Decode and convert raw slices.
- [ ] (Optional) Implement CoinMarketCap provider using API key.
- [ ] (Optional) Implement CoinMarketCap provider using API key.
- [x] Create `Client` that tries providers in order and returns first successful result.
- [x] Handle rate limiting (sleep between requests) and retry logic.
- [x] Add debug logging controlled by env var `CMC_DEBUG`.

## Core utilities

- [x] Define `Candle` type.
- [x] Implement `RenderCandlesASCII` in `chart.go`:
  - [x] Border drawing, scaling, wicks, bodies.
  - [x] Title with symbol, price range, and timestamps.
- [x] Implement layout math (`gridForCount`, `mathCeilSqrt`).
- [x] Add helper functions for view-to-duration and granularity.

## UI

### General structure

 - [x] Create `AppUI` struct containing:
  - `*tview.Application`, `*tview.Pages`, layout flexes.
  - Slice of `ChartWidget` (symbol, TextView, candles, lock, errMsg).
  - Countdown display and help text.
  - Fields for view type, refresh interval, candle count.
  - Channels for stop and nextTick atomic.
 - [x] Initialize UI in `NewApp()`:
  - [x] Build root flex layout and bottom bar.
  - [x] Add default symbols from config/CLI.
  - [x] Setup key capture handlers.

### Input handling

 - [x] Global key capture: Ctrl‑C → immediate quit; Ctrl‑Z → suspend.
 - [x] `a`/`A`: show add-symbol modal.
 - [x] `q`/`Q`: show quit confirmation.
 - [x] Numeric keys `1`–`4` to switch views and trigger refresh.

### Add-symbol modal

 - [x] Maintain a static list of common tokens (BTC, ETH, SOL, LTC, ADA, etc.).
 - [x] Create `tview.Form` with a checkbox per token.
 - [x] Track selections in a map, toggled by spacebar.
 - [x] Buttons: "Apply" (add selected, rebuild layout, refresh); "Cancel".
 - [x] Input capture within modal to handle Esc, Ctrl‑C, Ctrl‑Z, and 'q' for quit.

### Quit confirmation modal

 - [x] Modal text "Quit graph-watcher? (y/n)" with Yes/No buttons.
 - [x] Input capture for Y/N/Ctrl‑C.

### Layout management

 - [x] `addChart(symbol)` adds new widget with text view.
 - [x] `hasSymbol(symbol)` checks existing list ignoring case.
 - [x] `rebuildLayout()` arranges charts:
  - Special 3-chart layout for first three symbols.
  - Otherwise compute rows/cols grid and place each chart.

### Background loops

 - [x] `refreshLoop()` runs continuously:
  - Update all charts sequentially with `updateChartData`.
  - Wait `refreshInterval` between passes.
  - Respect `stopCh`.
 - [x] `updateAllCharts()` fetches data, delays between symbols, updates text.
 - [x] `countdownLoop()` updates countdown text every second.
 - [x] `updateChartData()` uses client to fetch and store candles/errMsg.

### Suspension

 - [x] Implement `suspendToShell()` to restore terminal and send SIGTSTP.

### Shutdown

 - [x] `Stop()` method closes `stopCh` once and stops the app.

## Error handling & resilience

- [ ] Ensure any API error sets `errMsg` and doesn’t crash the UI.
- [ ] If a fatal error occurs in `Run()` propagate to `main` and exit.
- [ ] On infinite loop or panic, application must stop gracefully.

## Configuration persistence

- [x] Store symbol list and view in config.
- [x] Save layout when user adds/removes symbols or changes view.
- [x] Load config at start and apply before building UI.

## Tests

- [x] Unit tests for parsing symbols, layout math, config load/save.
- [x] Test rendering with synthetic candle data.
- [x] Provider tests using mock HTTP servers.

## Documentation

- [x] Update README with full usage, controls, config file path.
- [x] Document environment variable `CMC_DEBUG`.
- [x] Provide build/test instructions in Makefile.

## Build & release

- [x] Create Makefile targets: `clean`, `build`, `test`.
- [x] Ensure `go.mod` required packages (`tview`, `tcell`).
- [ ] Bump version or add changelog.

## Optional enhancements

- [x] Allow entering a custom symbol not in the common list.
- [x] Persist window size or refresh interval.
- [x] Add command-line flags for API key, refresh interval, and theme.
- [ ] Show volume or other stats in chart title.
- [ ] Add color themes or alternate renderers.

---

Each item above should be checked off as it is implemented and tested. This checklist serves as the canonical roadmap for the project. 