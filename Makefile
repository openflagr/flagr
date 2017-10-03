PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

all: build

checks:
	@echo "Check deps"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)
	@echo "Checking project is in GOPATH"
	@(env bash $(PWD)/buildscripts/checkgopath.sh)

getdeps: checks
	@echo "Installing golint" && go get github.com/golang/lint/golint
	@echo "Installing gocyclo" && go get github.com/fzipp/gocyclo
	@echo "Installing deadcode" && go get github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go get github.com/gordonklaus/ineffassign
	@echo "Installing go-swagger" && go get github.com/go-swagger/go-swagger/cmd/swagger

verifiers: getdeps vet fmt lint cyclo spelling verify_swagger

vet:
	@echo "Running $@"
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult cmd
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult pkg

fmt:
	@echo "Running $@"
	@gofmt -d cmd
	@gofmt -d pkg

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golint -set_exit_status github.com/checkr/flagr/cmd...
	@${GOPATH}/bin/golint -set_exit_status github.com/checkr/flagr/pkg...

ineffassign:
	@echo "Running $@"
	@${GOPATH}/bin/ineffassign .

cyclo:
	@echo "Running $@"
	@${GOPATH}/bin/gocyclo -over 100 cmd
	@${GOPATH}/bin/gocyclo -over 100 pkg

deadcode:
	@echo "Running $@"
	@${GOPATH}/bin/deadcode

spelling:
	@echo "Running $@"
	@${GOPATH}/bin/misspell -error `find cmd/`
	@${GOPATH}/bin/misspell -error `find pkg/`
	@${GOPATH}/bin/misspell -error `find docs/`

verify_swagger:
	@echo "Running $@"
	@swagger validate ./swagger.yml

# Builds flagr, runs the verifiers then runs the tests.
check: test
test: verifiers build
	@echo "Running unit tests"
	@go test $(GOFLAGS) .
	@go test $(GOFLAGS) github.com/checkr/flagr/cmd...
	@go test $(GOFLAGS) github.com/checkr/flagr/pkg...

coverage: build
	@echo "Running all coverage for flagr"
	@./buildscripts/go-coverage.sh

# Builds flagr locally.
build: swagger
	@echo "Building flagr to $(PWD)/flagr ..."
	@CGO_ENABLED=0 go build -o $(PWD)/flagr

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf build
	@rm -rf release

swagger:
	@echo "Regenerate swagger files"
	@cp $(PWD)/swagger_gen/restapi/configure_flagr.go /tmp/configure_flagr.go
	@rm -rf $(PWD)/swagger_gen
	@mkdir $(PWD)/swagger_gen
	@swagger generate server -t ./swagger_gen -f ./swagger.yml
	@cp /tmp/configure_flagr.go $(PWD)/swagger_gen/restapi/configure_flagr.go
