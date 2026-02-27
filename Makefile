BINARY_NAME=graph-watcher
VERSION=1.1.0
BUILD_DIR=bin

.PHONY: build build-all build-linux build-windows build-darwin clean test help

help:
	@echo "graph-watcher - Terminal Crypto Candle Chart Viewer"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build for current OS"
	@echo "  build-all     - Build for Linux, Windows, and macOS (Intel & Silicon)"
	@echo "  test          - Run all tests"
	@echo "  clean         - Remove build artifacts"
	@echo "  help          - Show this help message"

build: clean
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "✓ Built Linux (amd64)"

build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Built Windows (amd64)"

build-darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "✓ Built macOS (amd64 & arm64)"

build-all: clean build-linux build-windows build-darwin
	@echo "✓ All platforms built in ./$(BUILD_DIR)"

test:
	go test -v -cover ./...

clean:
	rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned"
