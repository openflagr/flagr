######################################
# Prepare npm_builder
######################################
FROM node:25 as npm_builder
WORKDIR /go/src/github.com/openflagr/flagr
RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
ADD . .
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
RUN make build_ui

######################################
# Prepare go_builder
######################################
FROM golang:1.25.6-alpine as go_builder
WORKDIR /go/src/github.com/openflagr/flagr

RUN apk add --no-cache git make build-base
ADD . .
ENV CGO_ENABLED=0
ENV GOEXPERIMENT=greenteagc
RUN make build

FROM alpine

COPY --from=go_builder /go/src/github.com/openflagr/flagr/flagr .

ENV HOST=0.0.0.0
ENV PORT=18000
ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=false

COPY --from=npm_builder /go/src/github.com/openflagr/flagr/browser/flagr-ui/dist ./browser/flagr-ui/dist

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ADD --chown=appuser:appgroup ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

EXPOSE 18000

CMD "./flagr"
