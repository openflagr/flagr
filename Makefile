# Flagr — command entrypoint (repo root)
#
# Naming: multi-word targets use hyphens (build-ui, test-e2e, stop-ui).
# Run `make` or `make help` for the public catalog.

PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
FLAGR_UI_DIR := browser/flagr-ui
INTEGRATION_DIR := integration_tests

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
	@echo "  make build             Go server → ./flagr"
	@echo "  make build-ui          UI: npm install, lint, typecheck, Vite → dist/"
	@echo "  make rebuild           gen + build (server only)"
	@echo ""
	@echo "Run (dev)"
	@echo "  make start             run + run-ui in parallel (run make build first)"
	@echo "  make run               Pre-built ./flagr on :18000"
	@echo "  make run-ui            UI dev server on :8080"
	@echo "  make stop-ui           Free :18000 and :8080 (lsof, not pkill)"
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
	@echo "Building Flagr Server to $(PWD)/flagr ..."
	@CGO_ENABLED=0 go build -o $(PWD)/flagr github.com/openflagr/flagr/cmd/flagr-server

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
	@test -x ./flagr || (echo "Run make build first" && exit 1)
	@./flagr --port 18000

run-ui: flagr-ui-npm
	@cd $(FLAGR_UI_DIR) && npm run dev

start:
	$(MAKE) -j run run-ui

stop-ui:
	@-kill $$(lsof -ti:18000 2>/dev/null) 2>/dev/null; kill $$(lsof -ti:8080 2>/dev/null) 2>/dev/null; sleep 1; echo "Stopped UI services"

rebuild-run: build stop-ui start

DOCS_DIR := docs

# Copy pkg/config/env.go into docs/snippets for VitePress code import on flagr_env.md.
.PHONY: docs-sync-snippets
docs-sync-snippets:
	@mkdir -p $(DOCS_DIR)/snippets
	@cp $(PWD)/pkg/config/env.go $(DOCS_DIR)/snippets/env.go

serve-docs: docs-sync-snippets
	@cd $(DOCS_DIR) && npm ci && npm run docs:dev -- --port 8081 --host 127.0.0.1

build-docs: docs-sync-snippets
	@cd $(DOCS_DIR) && npm ci && npm run docs:build
	@mkdir -p $(DOCS_DIR)/.vitepress/dist/api_docs
	@cp $(DOCS_DIR)/api_docs/bundle.yaml $(DOCS_DIR)/api_docs/index.html $(DOCS_DIR)/.vitepress/dist/api_docs/
	@# Append api_docs to sitemap (copied post-build, not in VitePress page graph)
	@python3 -c "from pathlib import Path; p=Path('$(DOCS_DIR)/.vitepress/dist/sitemap.xml'); t=p.read_text(); u='https://openflagr.github.io/flagr/api_docs/';\
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
# Maintenance
# ------------------------------------------------------------------------------

.PHONY: swagger clean all

swagger: verify_swagger
	@echo "Regenerate swagger files"
	@rm -f /tmp/configure_flagr.go
	@cp $(PWD)/swagger_gen/restapi/configure_flagr.go /tmp/configure_flagr.go 2>/dev/null || :
	@rm -rf $(PWD)/swagger_gen
	@mkdir $(PWD)/swagger_gen
	@swagger generate server -t ./swagger_gen -f $(PWD)/docs/api_docs/bundle.yaml
	@cp /tmp/configure_flagr.go $(PWD)/swagger_gen/restapi/configure_flagr.go 2>/dev/null || :
	@mkdir -p $(PWD)/cmd/flagr-server
	@cp $(PWD)/swagger_gen/cmd/flagr-server/main.go $(PWD)/cmd/flagr-server/main.go
	@rm -rf $(PWD)/swagger_gen/cmd

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf build
	@rm -rf release

# Full local bootstrap (uncommon)
all: deps gen build build-ui run

# ------------------------------------------------------------------------------
# Private
# ------------------------------------------------------------------------------

api_docs:
	@echo "Installing swagger-merger" && npm install swagger-merger -g
	@swagger-merger -i $(PWD)/swagger/index.yaml -o $(PWD)/docs/api_docs/bundle.yaml

verifiers: verify_fmt verify_lint verify_swagger

verify_fmt:
	@echo "Running $@"
	@unformatted=$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*')); \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted" | xargs gofmt -w; \
		echo "gofmt reformatted the above files. Please review and re-commit."; \
		exit 1; \
	fi

verify_lint:
	@echo "Running $@"
	@golangci-lint run --timeout 5m -D errcheck ./pkg/...

verify_swagger:
	@echo "Running $@"
	@swagger validate $(PWD)/docs/api_docs/bundle.yaml