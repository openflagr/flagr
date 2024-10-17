######################################
# Prepare npm_builder
######################################
FROM node:18 as npm_builder
WORKDIR /go/src/github.com/Allen-Career-Institute/flagr
ARG ENVIRONMENT=production
ENV ENVIRONMENT=${ENVIRONMENT}
ADD . .
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
RUN make build_ui

######################################
# Prepare go_builder
######################################
FROM golang:1.21-alpine as go_builder
WORKDIR /go/src/github.com/Allen-Career-Institute/flagr

RUN apk add --no-cache git make build-base
ADD . .
RUN make build

FROM alpine

COPY --from=go_builder /go/src/github.com/Allen-Career-Institute/flagr/flagr .

ENV HOST=0.0.0.0
ENV PORT=3000
ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=false

# JWT Environment Variables
ENV FLAGR_JWT_AUTH_ENABLED=true
ENV FLAGR_JWT_AUTH_DEBUG=true
ENV FLAGR_JWT_AUTH_WHITELIST_PATHS="/api/v1/health,/api/v1/evaluation,/login,/callback,/static,/favicon.ico,/flags"
ENV FLAGR_JWT_AUTH_EXACT_WHITELIST_PATHS=",/,/login,/callback"
ENV FLAGR_JWT_AUTH_COOKIE_TOKEN_NAME="access_token"
#ENV FLAGR_JWT_AUTH_SECRET="<your-jwt-secret>"
ENV FLAGR_JWT_AUTH_NO_TOKEN_STATUS_CODE=307
ENV FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL="http://localhost:3000/login"
ENV FLAGR_JWT_AUTH_USER_PROPERTY=flagr_user
ENV FLAGR_JWT_AUTH_USER_CLAIM=uid
ENV FLAGR_JWT_AUTH_SIGNING_METHOD=HS256

COPY --from=npm_builder /go/src/github.com/Allen-Career-Institute/flagr/browser/flagr-ui/dist ./browser/flagr-ui/dist

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ADD --chown=appuser:appgroup ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

EXPOSE 3000

CMD "./flagr"
