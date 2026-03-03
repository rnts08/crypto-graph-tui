# Changelog

All notable changes to this project will be documented in this file.

## [1.1.1] - 2026-03-03

### Added
- Users can now press space to update all graphs immediately
- There's a visual indicator in the status field that shows the process.
- Added the ability to zoom in on specific graphs with the Z key.
- Added indicators for the timescale on the graphs

## [1.1.0] - 2026-02-27

### Added
- Volume display in chart titles.
- CoinMarketCap price provider (optional, requires API key).
- Version constant and display in UI.
- Improved error handling and resilience.

### Changed
- Refined UI status bar to show more information.

## [1.0.0] - 2026-01-30

### Added
- Initial release with CoinGecko and Coinbase providers.
- ASCII candlestick charting.
- Interactive TUI with symbol selection and timeframe switching.
- Persistent configuration.
- Support for custom symbols.
- Suspend to shell support (Ctrl+Z).
- CLI flags for symbols, interval, theme, and API key.
