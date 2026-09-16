# Flagr — command entrypoint (repo root)
#
# POSIX recipes so the same targets work on Linux, macOS, and Windows.
# On Windows, install Git for Windows; the Makefile prepends Git usr/bin so
# POSIX tools (sleep, test, mkdir, rm) resolve. Run `make` or `make help`.
#
# Naming: multi-word targets use hyphens (build-ui, test-e2e, stop-ui).

# Native Windows only. WSL inherits OS=Windows_NT but has /proc.
ifeq ($(OS),Windows_NT)
ifeq ($(wildcard /proc/version),)
  EXE := .exe
  PYTHON := python
  GIT_BIN := $(wildcard C:/PROGRA~1/Git/bin)
  GIT_USR_BIN := $(wildcard C:/PROGRA~1/Git/usr/bin)
  ifneq ($(GIT_USR_BIN),)
    export PATH := $(GIT_USR_BIN);$(GIT_BIN);$(PATH)
  endif
endif
endif

EXE ?=
PYTHON ?= python3

# Prepend GOPATH/bin in POSIX recipes. Git Bash splits PATH on `:`, so a
# Windows GOPATH (`C:\Users\...`) must be converted with cygpath first.
define with_gobin
gopath=$$(go env GOPATH); \
if command -v cygpath >/dev/null 2>&1; then gopath=$$(cygpath -u "$$gopath"); fi; \
export PATH="$$gopath/bin:$$PATH"; \
$(1)
endef

FLAGR_BIN := flagr$(EXE)
FLAGR_UI_DIR := browser/flagr-ui
INTEGRATION_DIR := integration_tests
DOCS_DIR := docs
SWAGGER_CONFIG_BAK := $(CURDIR)/.configure_flagr.go.bak
DEV_PORTS := 18000 8080

.DEFAULT_GOAL := help

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

.PHONY: help
help:
	@echo "Flagr Makefile — run from repository root"
	@echo ""
	@echo "Setup"
	@echo "  make deps              Go tools (swagger, golangci-lint)"
	@echo "  make gen               OpenAPI bundle + swagger_gen + cmd stub"
	@echo ""
	@echo "Build"
	@echo "  make build             Go server → ./$(FLAGR_BIN)"
	@echo "  make build-ui          UI: npm install, lint, typecheck, Vite → dist/"
	@echo "  make rebuild           gen + build (server only)"
	@echo ""
	@echo "Run (dev)"
	@echo "  make start             run + run-ui in parallel (run make build first)"
	@echo "  make run               Pre-built ./$(FLAGR_BIN) on :18000"
	@echo "  make run-ui            UI dev server on :8080"
	@echo "  make stop-ui           Free :18000 and :8080 (listener PIDs only)"
	@echo "  make rebuild-run       build → stop-ui → start"
	@echo "  make serve-docs        VitePress docs dev server (./docs, :8081)"
	@echo "  make build-docs        VitePress production build → docs/.vitepress/dist"
	@echo ""
	@echo "Test"
	@echo "  make test              Lint + swagger validate + Go unit tests"
	@echo "  make test-e2e          build + UI check + Playwright"
	@echo "  make test-integration  build + API integration (SQLite, local server)"
	@echo "  make test-integration-compose"
	@echo "                         Same suite vs 6 Docker Compose instances"
	@echo "  make bench-integration HTTP eval benchmarks (local server)"
	@echo "  make benchmark         Go package benchmarks (pkg/)"
	@echo ""
	@echo "CI (GitHub Actions call these)"
	@echo "  make ci                Unit test gate (lint + swagger + go test)"
	@echo "  make ci-swagger        Regenerate swagger; fail if git dirty"
	@echo "  make build-docs        VitePress docs build (docs_build job + Pages)"
	@echo "  make ci-integration    Compose integration tests + benchmarks"
	@echo ""
	@echo "Helm (helm/)"
	@echo "  make helm-lint         helm lint --strict + helm template"
	@echo "  make helm-unittest     helm unittest helm/ (plugin v1.1.2)"
	@echo ""
	@echo "Other"
	@echo "  make swagger           Regenerate swagger_gen/ (do not hand-edit)"
	@echo "  make clean             Remove test binaries and build artifacts"
	@echo "  make vendor            go mod tidy + vendor"

# ------------------------------------------------------------------------------
# Setup
# ------------------------------------------------------------------------------

.PHONY: deps gen vendor

deps:
	@CGO_ENABLED=0 go install github.com/go-swagger/go-swagger/cmd/swagger@v0.34.1
	@CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

gen: api_docs swagger

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------

.PHONY: build build-ui rebuild

build:
	@echo "Building Flagr Server to $(CURDIR)/$(FLAGR_BIN) ..."
	@CGO_ENABLED=0 go build -o $(CURDIR)/$(FLAGR_BIN) github.com/openflagr/flagr/cmd/flagr-server

flagr-ui-npm:
	@cd $(FLAGR_UI_DIR) && npm install

flagr-ui-check: flagr-ui-npm
	@cd $(FLAGR_UI_DIR) && npm run lint && npm run typecheck && npm run test

build-ui: flagr-ui-check
	@echo "Building Flagr UI ..."
	@cd $(FLAGR_UI_DIR) && npm run build

rebuild: gen build

# ------------------------------------------------------------------------------
# Run (dev)
# ------------------------------------------------------------------------------

.PHONY: start run run-ui stop-ui rebuild-run serve-docs build-docs

run:
	@test -f ./$(FLAGR_BIN) || (echo "Run make build first" && exit 1)
	@./$(FLAGR_BIN) --port 18000

run-ui: flagr-ui-npm
	@cd $(FLAGR_UI_DIR) && npm run dev

start:
	$(MAKE) -j run run-ui

stop-ui:
	@-sh $(CURDIR)/scripts/kill-port.sh $(DEV_PORTS)
	@sleep 1
	@echo "Stopped UI services"

rebuild-run: build stop-ui start

# Copy pkg/config/env.go into docs/snippets for VitePress code import on flagr_env.md.
.PHONY: docs-sync-snippets
docs-sync-snippets:
	@mkdir -p $(DOCS_DIR)/snippets
	@cp $(CURDIR)/pkg/config/env.go $(DOCS_DIR)/snippets/env.go

serve-docs: docs-sync-snippets
	@cd $(DOCS_DIR) && npm ci && npm run docs:dev -- --port 8081 --host 127.0.0.1

build-docs: docs-sync-snippets
	@cd $(DOCS_DIR) && npm ci && npm run docs:build
	@mkdir -p $(DOCS_DIR)/.vitepress/dist/api_docs
	@cp $(DOCS_DIR)/api_docs/bundle.yaml $(DOCS_DIR)/api_docs/index.html $(DOCS_DIR)/.vitepress/dist/api_docs/
	@# Append api_docs to sitemap (copied post-build, not in VitePress page graph)
	@$(PYTHON) -c "from pathlib import Path; p=Path('$(DOCS_DIR)/.vitepress/dist/sitemap.xml'); t=p.read_text(); u='https://openflagr.github.io/flagr/api_docs/';\
	(t:=t.replace('</urlset>', f'<url><loc>{u}</loc></url></urlset>')) if u not in t else None; p.write_text(t)"
	@echo "Docs built to $(DOCS_DIR)/.vitepress/dist (api_docs + sitemap)"

# ------------------------------------------------------------------------------
# Test
# ------------------------------------------------------------------------------

.PHONY: test test-e2e test-integration test-integration-compose bench-integration benchmark ci ci-swagger ci-integration

test: verifiers
	@go test -covermode=atomic -coverprofile=coverage.txt github.com/openflagr/flagr/pkg/...

test-e2e: build flagr-ui-check
	@echo "Installing Playwright browsers (chromium)..."
	@cd $(FLAGR_UI_DIR) && npx playwright install chromium
	@echo "Running Flagr UI e2e tests..."
	@cd $(FLAGR_UI_DIR) && npx playwright test

test-integration: build
	@echo "Running Go integration tests (local auto-start mode)..."
	@go test -tags=integration -count=1 -v ./integration_tests/

test-integration-compose:
	@$(MAKE) -C $(INTEGRATION_DIR) test

bench-integration: build
	@echo "Running Go integration benchmarks (local auto-start mode)..."
	@go test -tags=integration -bench=. -benchmem -count=1 -run=^$$ ./integration_tests/ > integration-bench.txt
	@echo "Benchmarks saved to integration-bench.txt"

benchmark:
	@go test -benchmem -run=^$$ -bench . ./pkg/...

ci: test

ci-swagger: swagger
	@echo "Checking swagger_gen is committed"
	@git diff --exit-code

ci-integration:
	@$(MAKE) -C $(INTEGRATION_DIR) test-and-bench

# ------------------------------------------------------------------------------
# Helm (helm/)
# ------------------------------------------------------------------------------

HELM_CHART := helm
HELM_UNITTEST_VERSION := v1.1.2
# Helm 4 plugin install from git requires --verify=false; Helm 3 ignores unknown flags poorly, so only pass it on v4.
HELM_MAJOR := $(shell helm version --short 2>/dev/null | sed -n 's/^v\([0-9]*\).*/\1/p')
HELM_UNITTEST_INSTALL_FLAGS := $(if $(filter 4,$(HELM_MAJOR)),--verify=false,)

.PHONY: helm-lint helm-unittest helm-plugin-unittest

helm-lint:
	helm lint --strict $(HELM_CHART)
	helm template flagr $(HELM_CHART) >/dev/null

helm-plugin-unittest:
	@helm plugin list 2>/dev/null | grep -q '^unittest' || \
		helm plugin install https://github.com/helm-unittest/helm-unittest.git --version $(HELM_UNITTEST_VERSION) $(HELM_UNITTEST_INSTALL_FLAGS)

helm-unittest: helm-plugin-unittest
	helm unittest $(HELM_CHART)

# ------------------------------------------------------------------------------
# Maintenance
# ------------------------------------------------------------------------------

.PHONY: swagger clean all

swagger: verify_swagger
	@echo "Regenerate swagger files"
	@rm -f $(SWAGGER_CONFIG_BAK)
	@cp $(CURDIR)/swagger_gen/restapi/configure_flagr.go $(SWAGGER_CONFIG_BAK) 2>/dev/null || :
	@rm -rf $(CURDIR)/swagger_gen
	@mkdir $(CURDIR)/swagger_gen
	@$(call with_gobin,swagger generate server -t ./swagger_gen -f $(CURDIR)/docs/api_docs/bundle.yaml)
	@cp $(SWAGGER_CONFIG_BAK) $(CURDIR)/swagger_gen/restapi/configure_flagr.go 2>/dev/null || :
	@rm -f $(SWAGGER_CONFIG_BAK)
	@mkdir -p $(CURDIR)/cmd/flagr-server
	@cp $(CURDIR)/swagger_gen/cmd/flagr-server/main.go $(CURDIR)/cmd/flagr-server/main.go
	@rm -rf $(CURDIR)/swagger_gen/cmd

clean:
	@echo "Cleaning up all the generated files"
	@find . \( -name '*.test' -o -name '*.test.exe' \) -exec rm -f {} +
	@rm -f ./flagr ./flagr.exe ./flagr-validate ./flagr-validate.exe $(SWAGGER_CONFIG_BAK)
	@rm -rf build release

# Full local bootstrap (uncommon)
all: deps gen build build-ui run

# ------------------------------------------------------------------------------
# Private
# ------------------------------------------------------------------------------

api_docs:
	@echo "Installing swagger-merger" && npm install swagger-merger -g
	@swagger-merger -i $(CURDIR)/swagger/index.yaml -o $(CURDIR)/docs/api_docs/bundle.yaml

verifiers: verify_fmt verify_lint verify_swagger

verify_fmt:
	@echo "Running $@"
	@unformatted=$$(gofmt -l $$(go list -f '{{.Dir}}' ./...) $$(go list -tags=integration -f '{{.Dir}}' ./integration_tests)); \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted" | xargs gofmt -w; \
		echo "gofmt reformatted the above files. Please review and re-commit."; \
		exit 1; \
	fi

verify_lint:
	@echo "Running $@"
	@$(call with_gobin,golangci-lint run --timeout 5m -D errcheck ./pkg/...)

verify_swagger:
	@echo "Running $@"
	@$(call with_gobin,swagger validate $(CURDIR)/docs/api_docs/bundle.yaml)
