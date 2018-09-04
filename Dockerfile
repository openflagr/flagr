FROM checkr/flagr-ci as builder
WORKDIR /go/src/github.com/checkr/flagr
ADD . .
RUN cd ./browser/flagr-ui/ && yarn install && yarn run build
RUN make build



FROM alpine:3.6
RUN apk add --no-cache libc6-compat ca-certificates curl
WORKDIR /go/src/github.com/checkr/flagr
VOLUME ["/data"]

ENV HOST=0.0.0.0
ENV PORT=18000
ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=false

COPY --from=builder /go/src/github.com/checkr/flagr/flagr ./flagr
COPY --from=builder /go/src/github.com/checkr/flagr/browser/flagr-ui/dist ./browser/flagr-ui/dist
COPY --from=builder /go/src/github.com/checkr/flagr/buildscripts ./buildscripts
ADD ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

EXPOSE 18000
CMD ./flagr
