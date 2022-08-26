######################################
# Prepare npm_builder
######################################
FROM node:16 as npm_builder
WORKDIR /go/src/github.com/openflagr/flagr
ADD . .
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
RUN make build_ui

######################################
# Prepare go_builder
######################################
FROM golang:1.17 as go_builder
WORKDIR /go/src/github.com/openflagr/flagr
ADD . .
RUN make build

######################################
# Copy from builder to alpine image
######################################
FROM frolvlad/alpine-glibc:alpine-3.15
RUN apk add --no-cache curl
WORKDIR /go/src/github.com/openflagr/flagr
VOLUME ["/data"]
ENV HOST=0.0.0.0
ENV PORT=18000
ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=false

COPY --from=go_builder /go/src/github.com/openflagr/flagr/flagr ./flagr
COPY --from=npm_builder /go/src/github.com/openflagr/flagr/browser/flagr-ui/dist ./browser/flagr-ui/dist
ADD ./buildscripts ./buildscripts
ADD ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 18000
CMD ./flagr
