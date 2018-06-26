PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
UIPATH := $(PWD)/browser/flagr-ui

################################
### Public
################################

all: deps gen build run

rebuild: gen run

test: verifiers
	@go test -race -covermode=atomic github.com/checkr/flagr/pkg/...

ci: verifiers
	@echo "Running all coverage for flagr"
	@retool do courtney -v -o ./coverage.txt -t="-race" -t="-covermode=atomic" github.com/checkr/flagr/pkg/...

build:
	@echo "Building flagr to $(PWD)/flagr ..."
	@CGO_ENABLED=1 go build -o $(PWD)/flagr github.com/checkr/flagr/swagger_gen/cmd/flagr-server

run:
	@$(PWD)/flagr --port 18000

gen: api_docs swagger goqueryset

deps: checks
	@echo "Installing retool" && go get -u github.com/twitchtv/retool
	@retool sync
	@retool build
	@retool do gometalinter --install
	@echo "Sqlite3" && sqlite3 -version

watch:
	@retool do fswatch

serve_docs:
	@yarn global add docsify-cli@4
	@docsify serve $(PWD)/docs

################################
### Private
################################

api_docs:
	@echo "Installing swagger-merger" && yarn global add swagger-merger
	@swagger-merger -i $(PWD)/swagger/index.yaml -o $(PWD)/docs/api_docs/bundle.yaml

checks:
	@echo "Check deps"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)
	@echo "Checking project is in GOPATH"
	@(env bash $(PWD)/buildscripts/checkgopath.sh)

verifiers: verify_gometalinter verify_swagger

verify_gometalinter:
	@echo "Running $@"
	@retool do gometalinter --config=.gometalinter.json ./pkg/...

verify_swagger:
	@echo "Running $@"
	@retool do swagger validate $(PWD)/docs/api_docs/bundle.yaml

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
	@retool do swagger generate server -t ./swagger_gen -f $(PWD)/docs/api_docs/bundle.yaml
	@cp /tmp/configure_flagr.go $(PWD)/swagger_gen/restapi/configure_flagr.go 2>/dev/null || :

goqueryset:
	@retool do go generate ./pkg/...
	@./buildscripts/goqueryset.sh
