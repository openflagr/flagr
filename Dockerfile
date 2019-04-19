######################################
# Prepare yarn and go build in builder
######################################
FROM checkr/flagr-ci:go1.12 as builder
WORKDIR /go/src/github.com/checkr/flagr
ADD . .

# Build UI
# FLAGR_UI_POSSIBLE_ENTITY_TYPES is useful for limiting the choices of entity types
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
RUN cd ./browser/flagr-ui/ && yarn install && yarn run build

# Build Go server
RUN make build


######################################
# Copy from builder to alpine image
######################################
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
