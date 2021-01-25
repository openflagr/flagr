PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
UIPATH := $(PWD)/browser/flagr-ui

################################
### Public
################################

all: deps gen build build_ui run

rebuild: gen build

test: verifiers
	@GO111MODULE=on go test -mod=vendor -race -covermode=atomic -coverprofile=coverage.txt github.com/checkr/flagr/pkg/...

.PHONY: benchmark
benchmark:
	@GO111MODULE=on go test -mod=vendor -race -benchmem -run=^$$ -bench . ./pkg/...

ci: test

.PHONY: vendor
vendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

build:
	@echo "Building Flagr Server to $(PWD)/flagr ..."
	@CGO_ENABLED=1 GO111MODULE=on go build -mod=vendor -o $(PWD)/flagr github.com/checkr/flagr/swagger_gen/cmd/flagr-server

build_ui:
	@echo "Building Flagr UI ..."
	@cd ./browser/flagr-ui/; npm install && npm run build

run_ui:
	@cd ./browser/flagr-ui/; npm run serve

run:
	@$(PWD)/flagr --port 18000

start:
	$(MAKE) -j run run_ui

gen: api_docs swagger

deps:
	@GO111MODULE=off go get -u github.com/myitcv/gobin
	@gobin github.com/go-swagger/go-swagger/cmd/swagger@v0.26.0
	@gobin github.com/golangci/golangci-lint/cmd/golangci-lint@v1.35.2

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
	@golangci-lint run -D errcheck ./pkg/...

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
