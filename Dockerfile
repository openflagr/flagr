######################################
# Prepare npm_builder
######################################
# $BUILDPLATFORM keeps npm install + Vite on the native builder instead of the
# target arch: running node under QEMU (linux/arm64 on an amd64 runner) dies with
# "qemu: uncaught target signal 4 (Illegal instruction)". The UI is a static
# artifact, so one native build serves every platform.
FROM --platform=$BUILDPLATFORM node:26-alpine AS npm_builder
WORKDIR /go/src/github.com/openflagr/flagr
ADD . .
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV VITE_FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
# Dist only. Lint/typecheck/tests run in CI (make build-ui), not in the image build.
RUN apk add --no-cache make && make flagr-ui-dist

######################################
# Prepare go_builder
######################################
FROM golang:1.27-alpine AS go_builder
WORKDIR /go/src/github.com/openflagr/flagr

RUN apk add --no-cache git make build-base
ADD . .
ENV CGO_ENABLED=0
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
