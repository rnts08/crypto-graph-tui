.PHONY: build test clean help

help:
	@echo "graph-watcher - Terminal Crypto Candle Chart Viewer"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run all tests with verbose output"
	@echo "  clean       - Remove binary and built artifacts"
	@echo "  help        - Show this help message"

build: clean
	go build -o graph-watcher .
	@echo "✓ Built: ./graph-watcher"

test:
	go test -v -cover ./...

clean:
	rm -f graph-watcher
	@echo "✓ Cleaned"
