# This Dockerfile builds a docker image which prints utilization information.
# It is designed to be build from the top level (go-agent directory).
#
# To build:
#   docker build -t utilization -f internal/tools/utilization/Dockerfile .
#
# Then to run:
#   docker run utilization
#

FROM golang:1.7

ENV GOPATH /tmp/gopath

RUN mkdir -p $GOPATH/src/github.com/newrelic/go-agent/

COPY . $GOPATH/src/github.com/newrelic/go-agent/

CMD go run ${GOPATH}/src/github.com/newrelic/go-agent/internal/tools/utilization/main.go
