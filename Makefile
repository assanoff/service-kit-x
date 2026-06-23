.DEFAULT_GOAL := help

APP         := service-kit-x
GO          ?= go
LINT        ?= golangci-lint
BIN         := bin/$(APP)

# Version is stamped into the binary (cmd.version) and used by `release`.
# Derived from git: the latest tag, or a short SHA (with -dirty) when untagged.
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -X github.com/assanoff/service-kit-x/cmd.version=$(VERSION)

# Dev tool versions (override to pin).
GOLANGCI    ?= github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
GOFUMPT     ?= mvdan.cc/gofumpt@latest

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Develop
# ---------------------------------------------------------------------------
.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Format code (gofumpt)
	@$(GO) tool gofumpt -w . 2>/dev/null || gofmt -w .

.PHONY: vet
vet: ## go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	$(LINT) run

.PHONY: build
build: ## Build the versioned binary into bin/
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) .
	@echo "built $(BIN) ($(VERSION))"

.PHONY: run
run: ## Run the server (go run . serve)
	$(GO) run -ldflags "$(LDFLAGS)" . serve

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------
.PHONY: test
test: ## Run unit tests (short)
	$(GO) test -race -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests (requires docker)
	$(GO) test -race -count=1 ./internal/tests/...

# ---------------------------------------------------------------------------
# Local infrastructure (Postgres for `make run` / `make migrate`)
# ---------------------------------------------------------------------------
.PHONY: up
up: ## Start local dependencies (docker compose up -d)
	docker compose up -d

.PHONY: down
down: ## Stop local dependencies (docker compose down)
	docker compose down

# ---------------------------------------------------------------------------
# Database migrations (reads config from .env, like the app)
# ---------------------------------------------------------------------------
.PHONY: migrate
migrate: ## Apply all migrations (up)
	$(GO) run . migrate up

.PHONY: migrate-down
migrate-down: ## Roll back one migration
	$(GO) run . migrate down

.PHONY: migrate-status
migrate-status: ## Show migration status
	$(GO) run . migrate status

# ---------------------------------------------------------------------------
# Protobuf / gRPC codegen
# ---------------------------------------------------------------------------
# Baseline the wire contract is diffed against. Defaults to the latest release
# tag; falls back to the `main` branch when there is no tag yet. Override, e.g.
#   make breaking AGAINST='.git#branch=main,subdir=proto'
BUF_BASELINE ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo main)
AGAINST      ?= .git#subdir=proto,ref=$(BUF_BASELINE)

.PHONY: proto
proto: ## Generate gRPC code from .proto (run proto-tools first)
	buf lint proto && buf generate proto

.PHONY: breaking
breaking: ## Detect breaking changes in the gRPC/proto contract vs $(BUF_BASELINE)
	buf breaking proto --against '$(AGAINST)'

.PHONY: proto-tools
proto-tools: ## Install protobuf codegen tools (buf, protoc-gen-go, protoc-gen-go-grpc)
	$(GO) install github.com/bufbuild/buf/cmd/buf@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: tools
tools: proto-tools ## Install all dev tools (lint, fmt, + proto)
	$(GO) install $(GOLANGCI)
	$(GO) install $(GOFUMPT)
	@echo "installed: golangci-lint, gofumpt (+ proto tools)"

# ---------------------------------------------------------------------------
# Release (application, not a Go library)
#
# Unlike the SDK, this is an app: there is no public Go API to diff, so no
# gorelease. A release is a version tag — CI (or `make build`) then produces the
# versioned binary/image. check-version enforces one semver step above the
# latest tag.
# ---------------------------------------------------------------------------
.PHONY: check-clean
check-clean:
	@test -z "$$(git status --porcelain)" || \
		{ echo "working tree is dirty — commit (or stash) changes before releasing"; \
		  git status --short; exit 1; }

.PHONY: check-version
check-version:
	@test -n "$(V)" || { echo "usage: make release V=vX.Y.Z"; exit 1; }
	@echo "$(V)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$' || \
		{ echo "V must be vX.Y.Z (no prerelease/build suffix): got $(V)"; exit 1; }
	@cur=$$(git tag --list 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort -V | tail -n1); \
	if [ -z "$$cur" ]; then \
		echo ">> first release ($(V)); no prior tag to compare against"; \
	else \
		cv=$${cur#v}; cM=$${cv%%.*}; cr=$${cv#*.}; cm=$${cr%%.*}; cp=$${cr##*.}; \
		np="v$$cM.$$cm.$$((cp + 1))"; nm="v$$cM.$$((cm + 1)).0"; nj="v$$((cM + 1)).0.0"; \
		case "$(V)" in \
			"$$np"|"$$nm"|"$$nj") echo ">> $(V) is exactly one step above $$cur" ;; \
			*) echo "ERROR: $(V) must be exactly one step above the latest tag $$cur"; \
			   echo "       allowed: $$np (patch) | $$nm (minor) | $$nj (major)"; exit 1 ;; \
		esac; \
	fi

.PHONY: release
release: check-clean check-version ## Tag & push a release: make release V=v0.1.0
	@echo ">> verifying build & tests"; $(GO) build ./... && $(GO) test -short ./...
	@echo ">> tagging $(V)"; git tag -a $(V) -m "Release $(V)"
	@echo ">> pushing $(V)"; git push origin $(V)
