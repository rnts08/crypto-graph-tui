# Detailed Implementation TODO

This list breaks down every required feature into discrete, actionable items. Each checkbox represents a task or sub‑task that will guide development and track progress.

## Configuration

- [x] Design `Config` struct (symbols, view, refresh interval, API key, themes, window size).
- [x] Implement `LoadConfig()` reading JSON (XDG_CONFIG_HOME or home dir).
- [x] Implement `SaveConfig()` writing current state (symbols, window size, etc.).
- [x] Add command‑line argument parsing (symbols, interval, API key, theme).
- [x] Merge CLI args and config on startup.
- [x] Invoke `SaveConfig()` on updates and graceful exit.

## Data provider layer

- [x] Define `PriceProvider` interface with `FetchOHLC`.
- [x] Implement CoinGecko provider (mapping, fetching, decoding).
- [x] Implement Coinbase provider (granularity, multi-page fetching).
- [x] (Optional) Implement CoinMarketCap provider (API key, priority).
- [x] Create `Client` with fallback logic and rate-limit awareness (semaphores).
- [x] Add debug logging via `CMC_DEBUG`.

## Core utilities

- [x] Define `Candle` type.
- [x] Implement `RenderCandlesASCII` in `chart.go`:
  - [x] Border drawing, scaling, wicks, bodies.
  - [x] Title with symbol, price, volume, and timestamps.
  - [x] Adaptive layout (vertical compression).
- [x] Implement layout math (square-ish grid via `GridForCount`).
- [x] Add helper functions for view-to-duration and granularity.

## UI (Bubble Tea Implementation)

### General structure

- [x] Create `Model` struct containing all state.
- [x] Implement `Init` to start fetch loops.
- [x] Implement `Update` for reactive event handling.
- [x] Implement `View` for composable terminal rendering.

### Input handling

- [x] Global key capture: Ctrl‑C (quit), Ctrl‑Z (suspend).
- [x] `a`/`A`: show add-symbol modal.
- [x] `q`/`Q`: show quit confirmation.
- [x] `1`–`4`: switch views (1D, WTD, MTD, YTD).

### Modals

- [x] Add-symbol modal:
  - [x] Common tokens list with navigation (↑/↓, j/k).
  - [x] Multi-select with Spacebar.
  - [x] Custom symbol entry via `c`.
- [x] Quit confirmation modal.

### Layout & Rendering

- [x] Automatic grid tiling based on symbol count.
- [x] Intelligent chart scaling and title truncation.
- [x] Support for color themes (default, dark, retro).

## Error handling & resilience

- [x] API errors set `Error` state in charts without crashing.
- [x] Graceful shutdown and config save on signals.
- [x] Recover from panics in main loop.

## Tests

- [x] Unit tests for config, layout math, and symbol parsing.
- [x] Synthetic data rendering tests.
- [x] Provider tests with `httptest`.

## Documentation

- [x] Update README with usage, controls, and architecture.
- [x] Document `CMC_DEBUG` and API key usage.
- [x] Provide build/test instructions.

## Build & release

- [x] Makefile with `build`, `test`, `clean`.
- [x] Versioning and CHANGELOG.md.