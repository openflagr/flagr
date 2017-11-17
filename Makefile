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
	@./buildscripts/go-coverage.sh

build:
	@echo "Building flagr to $(PWD)/flagr ..."
	@CGO_ENABLED=1 go build -o $(PWD)/flagr github.com/checkr/flagr/swagger_gen/cmd/flagr-server

run:
	@$(PWD)/flagr --port 18000

gen: api_docs swagger goqueryset

deps: checks
	@echo "Installing dep" && go get -u github.com/golang/dep/cmd/dep
	@echo "Installing golint" && go get -u github.com/golang/lint/golint
	@echo "Installing gocyclo" && go get -u github.com/fzipp/gocyclo
	@echo "Installing deadcode" && go get -u github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get -u github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go get -u github.com/gordonklaus/ineffassign
	@echo "Installing go-swagger" && go get -u github.com/go-swagger/go-swagger/cmd/swagger
	@echo "Installing goqueryset" && go get -u github.com/jirfag/go-queryset/cmd/goqueryset
	@echo "Installing gt" && go get -u rsc.io/gt
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

verifiers: vet fmt lint cyclo spelling verify_swagger

vet:
	@echo "Running $@"
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult pkg

fmt:
	@echo "Running $@"
	@gofmt -d pkg

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golint -set_exit_status github.com/checkr/flagr/pkg...

ineffassign:
	@echo "Running $@"
	@${GOPATH}/bin/ineffassign .

cyclo:
	@echo "Running $@"
	@${GOPATH}/bin/gocyclo -over 100 pkg

deadcode:
	@echo "Running $@"
	@${GOPATH}/bin/deadcode

spelling:
	@echo "Running $@"
	@${GOPATH}/bin/misspell -error `find pkg/`
	@${GOPATH}/bin/misspell -error `find docs/` -source markdown-like

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
