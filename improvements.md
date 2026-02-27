# Improvements & Future Work

These are actionable items discovered while auditing the current codebase; each entry includes rationale and suggested changes.

1. **Implemented** suspendToShell() and Stop() semantics
   - Rationale: Users expect Ctrl+Z to suspend to shell and resume; Bubble Tea handles terminal state but explicit suspend support improves UX.
   - Suggestion: Added `suspendCmd()` and a `Stop()` method in `ui.go`. (Done)

2. **Implemented** formal CLI flag parsing
   - Rationale: Passing `--help` currently fell back to defaults; now we support `--help`, `--symbols`, `--interval`, and `--api-key`.
   - Suggestion: Added flags in `main.go` merging with config. (Done)

3. **Documented** `CMC_DEBUG` in README
   - Rationale: Debug logging is present but undocumented.
   - Suggestion: Added paragraph under Data Sources. (Done)

4. **Improved** grid/layout algorithm
   - Rationale: `GridForCountCols` was simplistic.
   - Suggestion: Now delegates to `GridForCount` for balanced rows/cols; UI uses same math as chart utils. (Done)

5. **Parallelized** fetching with rate-limit awareness

6. **Implemented** add themes and color configuration
   - Rationale: Hardcoded lipgloss colors are fine, but user-selectable themes improve accessibility and aesthetics.
   - Suggestion: Exposed a `theme` field in `Config`, two presets (light/dark), and used lipgloss styles derived from the theme. (Done)

7. **Persisted** additional UI state
   - Rationale: Restoring refresh interval, last window size, and symbol ordering improves continuity across sessions.
   - Suggestion: Extended `Config` with `WindowWidth`/`WindowHeight`, saved on exit and restored on start. (Done)

8. **Implemented** add custom symbol text entry
   - Rationale: Users may want symbols not in the static list.
   - Suggestion: Added modal with `c` key to enter arbitrary ticker; symbols normalized and persisted. (Done)
   - Suggestion: Expose a `theme` field in `Config`, provide a couple of presets, and use lipgloss styles derived from the theme.

7. **Persisted** additional UI state
   - Rationale: Restoring refresh interval, last window size, and symbol ordering improves continuity across sessions.
   - Suggestion: Extended `Config` with `WindowWidth`/`WindowHeight`, saved on exit and restored on start. (Done)

8. **Implemented** add custom symbol text entry
   - Rationale: Users may want symbols not in the static list.
   - Suggestion: Added modal with `c` key to enter arbitrary ticker; symbols normalized and persisted. (Done)

9. **Implemented** expand provider test coverage with httptest
   - Rationale: Unit tests used mocks; now real HTTP servers exercise JSON decoding, error paths, and rate-limiting behavior.
   - Added tests simulating 429, malformed JSON, partial entries, and happy paths for both CoinGecko and Coinbase. (Done)

10. Changelog and release process
    - Rationale: No changelog or versioning present; useful for releases and user-facing notes.
    - Suggestion: Add a `CHANGELOG.md` and tag releases; include a `version` constant in `main.go` that can be baked into builds.

11. Optional: Add CoinMarketCap provider
    - Rationale: Additional paid sources increase resilience for users with API keys.
    - Suggestion: Implement as a provider behind a config-provided API key and mark it higher priority when configured.

12. UX polishing: smoother titles and volume display
    - Rationale: Users may want volume/time-range in title and clearer formatting when widths are narrow.
    - Suggestion: Include latest volume in title and truncate intelligently; consider vertical compression modes.

---

If you'd like, I can start implementing the high-priority items (CLI flags, CMC_DEBUG docs, and suspend support) next — tell me which to prioritize.