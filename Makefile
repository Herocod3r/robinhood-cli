.PHONY: build test test-race lint fmt vet tidy clean run help

GO ?= go
BINARY := rh
PKG := github.com/herocod3r/robinhood-cli
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X $(PKG)/internal/buildinfo.Version=$(VERSION) -X $(PKG)/internal/buildinfo.Commit=$(COMMIT)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the rh binary
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/rh

test: ## Run unit tests
	$(GO) test ./...

test-race: ## Run tests with -race
	$(GO) test -race ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format
	$(GO) fmt ./...

vet: ## go vet
	$(GO) vet ./...

tidy: ## go mod tidy
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf dist/

run: build ## Build and run
	./$(BINARY) $(ARGS)

# ---- Release ----
# GoReleaser is invoked via `go run` so contributors do not have to
# install it locally; the version is pinned to match .github/workflows/release.yaml.
GORELEASER ?= go run github.com/goreleaser/goreleaser/v2@v2.4.4

.PHONY: release-dry-run release-snapshot

release-dry-run: ## Full release snapshot (all platforms, no publish)
	$(GORELEASER) release --snapshot --clean --skip=publish

release-snapshot: ## Fast single-target snapshot build
	$(GORELEASER) build --snapshot --clean --single-target
