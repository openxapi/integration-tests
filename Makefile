# Integration Tests Makefile

.PHONY: test test-spot test-umfutures test-all deps clean

# Default test target
test: test-all

# Test all modules
test-all: test-spot test-umfutures

# Test Binance Spot WebSocket
test-spot:
	@echo "Running Binance Spot WebSocket integration tests..."
	cd binance/asyncapi/spot && go test -v ./...

# Test Binance USD-M Futures WebSocket
test-umfutures:
	@echo "Running Binance USD-M Futures WebSocket integration tests..."
	cd binance/asyncapi/umfutures && go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean test cache
clean:
	go clean -testcache

# Run specific test by name
test-name:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make test-name NAME=TestName"; \
		exit 1; \
	fi
	cd binance/asyncapi/spot && go test -v -run $(NAME) ./...
	cd binance/asyncapi/umfutures && go test -v -run $(NAME) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd binance/asyncapi/spot && go test -v -coverprofile=coverage.out ./...
	cd binance/asyncapi/umfutures && go test -v -coverprofile=coverage.out ./...

# Run tests in verbose mode with race detection
test-race:
	@echo "Running tests with race detection..."
	cd binance/asyncapi/spot && go test -v -race ./...
	cd binance/asyncapi/umfutures && go test -v -race ./...