######################################
# Prepare go_builder
######################################
FROM golang:1.19.0-alpine3.16 as go_builder
WORKDIR /go/src/github.com/paubox/paubox-flagr

RUN apk add --no-cache git make build-base
ADD . .
RUN make build

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# ADD --chown=appuser:appgroup ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

EXPOSE 3000
CMD ./flagr
