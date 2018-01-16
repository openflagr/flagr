PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
UIPATH := $(PWD)/browser/flagr-ui

################################
### Public
################################

all: deps gen build run

rebuild: gen run

test: verifiers
	@echo "Running all coverage for flagr"
	@courtney -v -o ./coverage.txt -t="-race" -t="-covermode=atomic" github.com/checkr/flagr/pkg/...


build:
	@echo "Building flagr to $(PWD)/flagr ..."
	@CGO_ENABLED=1 go build -o $(PWD)/flagr github.com/checkr/flagr/swagger_gen/cmd/flagr-server

run:
	@$(PWD)/flagr --port 18000

gen: api_docs swagger goqueryset

deps: checks
	@echo "Installing dep" && go get -u github.com/golang/dep/cmd/dep
	@echo "Installing gometalinter" && go get -u github.com/alecthomas/gometalinter
	@gometalinter --install
	@echo "Installing go-swagger" && go get -u github.com/go-swagger/go-swagger/cmd/swagger
	@echo "Installing goqueryset" && go get -u github.com/jirfag/go-queryset/cmd/goqueryset
	@echo "Installing courtney" && go get -u github.com/dave/courtney
	@echo "Installing gomock" && go get -u github.com/golang/mock/gomock && go get github.com/golang/mock/mockgen
	@echo "Installing fswatch" && go get -u github.com/codeskyblue/fswatch
	@echo "Sqlite3" && sqlite3 -version

watch:
	@fswatch

serve_docs:
	@yarn global add docsify-cli
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
	@gometalinter --config=.gometalinter.json ./pkg/...

verify_swagger:
	@echo "Running $@"
	@swagger validate $(PWD)/docs/api_docs/bundle.yaml

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

goqueryset:
	@go generate ./pkg/...
	@./buildscripts/goqueryset.sh
