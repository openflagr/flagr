tidy:
	@GO111MODULE=on go mod tidy

vendor: tidy
	@GO111MODULE=on go mod vendor

build:
	@GO111MODULE=on go build

test:
	@GO111MODULE=on go test -race -covermode=atomic .
