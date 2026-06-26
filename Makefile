PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
FLAGR_UI_DIR := browser/flagr-ui

################################
### Public
################################

all: deps gen build build_ui run

rebuild: gen build

test: verifiers
	@go test -covermode=atomic -coverprofile=coverage.txt github.com/openflagr/flagr/pkg/...

.PHONY: benchmark
benchmark:
	@go test -benchmem -run=^$$ -bench . ./pkg/...

ci: test

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor

build:
	@echo "Building Flagr Server to $(PWD)/flagr ..."
	@CGO_ENABLED=0 go build -o $(PWD)/flagr github.com/openflagr/flagr/cmd/flagr-server

.PHONY: deps_ui lint-ui typecheck-ui verify_ui build_ui run_ui test-e2e
deps_ui:
	@cd $(FLAGR_UI_DIR) && npm install

lint-ui: deps_ui
	@cd $(FLAGR_UI_DIR) && npm run lint

typecheck-ui: deps_ui
	@cd $(FLAGR_UI_DIR) && npm run typecheck

verify_ui: deps_ui
	@cd $(FLAGR_UI_DIR) && npm run lint && npm run typecheck

build_ui: verify_ui
	@echo "Building Flagr UI ..."
	@cd $(FLAGR_UI_DIR) && npm run build

run_ui: deps_ui
	@cd $(FLAGR_UI_DIR) && npm run dev

test-e2e: build verify_ui
	@echo "Running Flagr UI e2e tests..."
	@cd $(FLAGR_UI_DIR) && npx playwright test

.PHONY: test-integration
test-integration: build
	@echo "Running Go integration tests (local auto-start mode)..."
	@go test -tags=integration -count=1 -v ./integration_tests/

.PHONY: bench-integration
bench-integration: build
	@echo "Running Go integration benchmarks (local auto-start mode)..."
	@go test -tags=integration -bench=. -benchmem -count=1 -run=^$$ ./integration_tests/ > integration-bench.txt
	@echo "Benchmarks saved to integration-bench.txt"


stop-ui:
	@-kill $$(lsof -ti:18000 2>/dev/null) 2>/dev/null; kill $$(lsof -ti:8080 2>/dev/null) 2>/dev/null; sleep 1; echo "Stopped UI services"

rebuild-run: build stop-ui start

start:
	$(MAKE) -j run run_ui

gen: api_docs swagger

deps:
	@CGO_ENABLED=0 go install github.com/go-swagger/go-swagger/cmd/swagger@v0.34.1
	@CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

serve_docs:
	@npm install -g docsify-cli@4
	@docsify serve $(PWD)/docs

################################
### Private
################################

api_docs:
	@echo "Installing swagger-merger" && npm install swagger-merger -g
	@swagger-merger -i $(PWD)/swagger/index.yaml -o $(PWD)/docs/api_docs/bundle.yaml

verifiers: verify_lint verify_swagger

verify_lint:
	@echo "Running $@"
	@golangci-lint run --timeout 5m -D errcheck ./pkg/...

verify_swagger:
	@echo "Running $@"
	@swagger validate $(PWD)/docs/api_docs/bundle.yaml

verify_swagger_nochange: swagger
	@echo "Running verify_swagger_nochange to make sure the swagger generated code is checked in"
	@git diff --exit-code

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf build
	@rm -rf release

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
