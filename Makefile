# Integration Tests Makefile

BINANCE_WS_GO_MODULES := $(shell find src/binance/go/ws -mindepth 1 -maxdepth 2 -name go.mod -exec dirname {} \; | sort)
BINANCE_WS_GO_TEST_FLAGS := -v -run TestFullIntegrationSuite -timeout 20m
BINANCE_WS_GO_SOURCE ?= local
BINANCE_WS_GO_TAG ?=
BINANCE_GO_REPO_URL ?= https://github.com/openxapi/binance-go.git
BINANCE_WS_SPOT_ENV := BINANCE_API_KEY=$$BINANCE_SPOT_API_KEY BINANCE_SECRET_KEY=$$BINANCE_SPOT_SECRET_KEY BINANCE_RSA_API_KEY=$$BINANCE_SPOT_RSA_API_KEY BINANCE_RSA_PRIVATE_KEY_PATH=$$BINANCE_SPOT_RSA_PRIVATE_KEY_PATH BINANCE_ED25519_API_KEY=$$BINANCE_SPOT_ED25519_API_KEY BINANCE_ED25519_PRIVATE_KEY_PATH=$$BINANCE_SPOT_ED25519_PRIVATE_KEY_PATH

.PHONY: test test-all test-binance test-binance-ws test-binance-ws-go deps clean

# Default test target
test:
	@$(MAKE) test-all || exit -1

# Test all modules
test-all:
	@$(MAKE) test-binance || exit -1

test-binance:
	@$(MAKE) test-binance-ws || exit -1

test-binance-ws:
	@$(MAKE) test-binance-ws-go || exit -1

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

# Run all Binance WebSocket integration tests
test-binance-ws-go:
	@if [ "$(BINANCE_WS_GO_SOURCE)" != "local" ] && [ "$(BINANCE_WS_GO_SOURCE)" != "main" ] && [ "$(BINANCE_WS_GO_SOURCE)" != "tag" ]; then \
		echo "BINANCE_WS_GO_SOURCE must be one of: local, main, tag"; \
		exit 1; \
	fi
	@RESOLVED_TAG=""; \
	RESOLVED_COMMIT=""; \
	if [ "$(BINANCE_WS_GO_SOURCE)" = "tag" ]; then \
		RESOLUTION_OUTPUT=$$(python3 scripts/resolve_binance_tag.py --tag "$(BINANCE_WS_GO_TAG)" --repo "$(BINANCE_GO_REPO_URL)"); \
		STATUS=$$?; \
		if [ $$STATUS -ne 0 ]; then \
			echo "$$RESOLUTION_OUTPUT"; \
			exit $$STATUS; \
		fi; \
		eval "$$RESOLUTION_OUTPUT"; \
		if [ -z "$$RESOLVED_COMMIT" ]; then \
			echo "Failed to resolve commit for BINANCE_WS_GO_TAG=$(BINANCE_WS_GO_TAG)"; \
			exit 1; \
		fi; \
	fi; \
	FAILED=0; \
	for module in $(BINANCE_WS_GO_MODULES); do \
		echo "Running Binance WS integration tests in $$module..."; \
		if [ "$$module" = "src/binance/go/ws/spot" ]; then \
			ENV_PREFIX="$(BINANCE_WS_SPOT_ENV) "; \
		else \
			ENV_PREFIX=""; \
		fi; \
		if [ "$(BINANCE_WS_GO_SOURCE)" = "local" ]; then \
			if ! (cd $$module && env $${ENV_PREFIX}go test $(BINANCE_WS_GO_TEST_FLAGS)); then \
				echo "Tests failed in $$module"; \
				FAILED=1; \
			fi; \
			else \
				if ! (cd $$module && \
					TMP_MOD=".$$(basename $$module).remote.$$$$.mod"; \
					TMP_SUM="$${TMP_MOD%.mod}.sum"; \
					cp go.mod $$TMP_MOD; \
					if [ -f go.sum ]; then \
						cp go.sum $$TMP_SUM; \
					else \
						:; \
					fi; \
					go mod edit -modfile=$$TMP_MOD -dropreplace=github.com/openxapi/binance-go/ws >/dev/null 2>&1 || true; \
					go mod edit -modfile=$$TMP_MOD -dropreplace=github.com/openxapi/binance-go/rest >/dev/null 2>&1 || true; \
					go mod edit -modfile=$$TMP_MOD -droprequire=github.com/openxapi/binance-go/ws >/dev/null 2>&1 || true; \
					go mod edit -modfile=$$TMP_MOD -droprequire=github.com/openxapi/binance-go/rest >/dev/null 2>&1 || true; \
					VERSION_REF=""; \
					DESC="@latest"; \
					if [ "$(BINANCE_WS_GO_SOURCE)" = "main" ]; then \
						VERSION_REF="main"; \
						DESC="@main"; \
					elif [ "$(BINANCE_WS_GO_SOURCE)" = "tag" ]; then \
						VERSION_REF="$$RESOLVED_COMMIT"; \
						DESC="@$$RESOLVED_TAG (commit $$RESOLVED_COMMIT)"; \
					fi; \
					if [ -n "$$VERSION_REF" ]; then \
						REF_SUFFIX="@"$$VERSION_REF; \
					else \
						REF_SUFFIX=""; \
					fi; \
					GOFLAGS_VALUE="-modfile=$$TMP_MOD"; \
					echo "Syncing github.com/openxapi/binance-go/{rest,ws}$$DESC for $$module"; \
					STATUS=0; \
					if GOFLAGS="$$GOFLAGS_VALUE" go get github.com/openxapi/binance-go/ws$$REF_SUFFIX github.com/openxapi/binance-go/rest$$REF_SUFFIX >/dev/null; then \
						if env GOFLAGS="$$GOFLAGS_VALUE" $${ENV_PREFIX}go test $(BINANCE_WS_GO_TEST_FLAGS); then \
							STATUS=0; \
						else \
							STATUS=1; \
						fi; \
					else \
						STATUS=1; \
					fi; \
					rm -f $$TMP_MOD $$TMP_SUM; \
					exit $$STATUS; \
				); then \
				echo "Tests failed in $$module"; \
				FAILED=1; \
			fi; \
		fi; \
	done; \
	if [ $$FAILED -ne 0 ]; then \
		echo "One or more Binance WS modules failed"; \
		exit -1; \
	fi; \
	echo "All Binance WS integration tests passed"
