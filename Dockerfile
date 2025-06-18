######################################
# Prepare npm_builder
######################################
FROM 537984406465.dkr.ecr.ap-south-1.amazonaws.com/docker-hub/library/node:18 as npm_builder
WORKDIR /go/src/github.com/Allen-Career-Institute/flagr
COPY . .
ARG FLAGR_UI_POSSIBLE_ENTITY_TYPES=null
ENV VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES ${FLAGR_UI_POSSIBLE_ENTITY_TYPES}
RUN make build_ui

######################################
# Prepare go_builder
######################################
FROM 537984406465.dkr.ecr.ap-south-1.amazonaws.com/docker-hub/library/golang:1.21-alpine as go_builder
WORKDIR /go/src/github.com/Allen-Career-Institute/flagr

RUN apk add --no-cache build-base git make
COPY . .
RUN make build

FROM 537984406465.dkr.ecr.ap-south-1.amazonaws.com/docker-hub/library/alpine:3.21

COPY --from=go_builder /go/src/github.com/Allen-Career-Institute/flagr/flagr .

ENV HOST=0.0.0.0
# ENV PORT=3000 (for local testing)
ENV PORT=18000

ENV FLAGR_DB_DBDRIVER=sqlite3
ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=true

# JWT Environment Variables
ENV FLAGR_JWT_AUTH_ENABLED=true
ENV FLAGR_JWT_AUTH_DEBUG=true
ENV FLAGR_JWT_AUTH_WHITELIST_PATHS="/api/v1/health,/api/v1/evaluation,/login,/callback,/static,/favicon.ico,/flags,/debug/pprof"
ENV FLAGR_JWT_AUTH_EXACT_WHITELIST_PATHS=",/,/login,/callback"
ENV FLAGR_JWT_AUTH_COOKIE_TOKEN_NAME="access_token"
# ENV FLAGR_JWT_AUTH_SECRET="secret"
ENV FLAGR_JWT_AUTH_NO_TOKEN_STATUS_CODE=401
# ENV FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL="http://localhost:3000/login" (for local testing)
# ENV FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL="http://localhost:18000/login" (to be overriden in k8 yaml file as per env)
ENV FLAGR_JWT_AUTH_USER_CLAIM=uid
ENV FLAGR_JWT_AUTH_SIGNING_METHOD=HS256

# CORS Environment Variables
ENV FLAGR_CORS_ALLOWED_METHODS="GET,POST,PUT,DELETE,PATCH,OPTIONS"
ENV FLAGR_CORS_ALLOWED_HEADERS="*"

COPY --from=npm_builder /go/src/github.com/Allen-Career-Institute/flagr/browser/flagr-ui/dist ./browser/flagr-ui/dist

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --chown=appuser:appgroup ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

# EXPOSE 3000 (for local testing)
EXPOSE 18000

CMD "./flagr"
