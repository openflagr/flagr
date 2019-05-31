######################################
# Prepare go_builder
######################################
FROM golang:1.12 as go_builder
WORKDIR /go/src/github.com/checkr/flagr
ADD . .
RUN make build

######################################
# Copy from builder to alpine image
######################################
FROM alpine:3.6
RUN apk add --no-cache libc6-compat ca-certificates curl
WORKDIR /go/src/github.com/checkr/flagr
COPY --from=go_builder /go/src/github.com/checkr/flagr/flagr ./flagr
VOLUME ["/data"]
ENV HOST=0.0.0.0
ENV PORT=18000
ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_RECORDER_ENABLED=false
EXPOSE 18000
CMD ./flagr
