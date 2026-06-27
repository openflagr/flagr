#!/bin/sh
# Start backend + frontend for e2e tests. Idempotent — skips already-running services.
# Repo root: make build (backend), make run-ui (frontend). Playwright webServer entrypoint.

set -e

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
BACKEND_PORT=18000
FRONTEND_PORT=8080

cleanup() {
  kill "$(lsof -ti:"$BACKEND_PORT" 2>/dev/null)" 2>/dev/null || true
  kill "$(lsof -ti:"$FRONTEND_PORT" 2>/dev/null)" 2>/dev/null || true
  exit
}
trap cleanup INT TERM

started_any=false

# --- Backend ---
if curl -sf "http://127.0.0.1:$BACKEND_PORT/api/v1/health" > /dev/null 2>&1; then
  echo "e2e-server: backend already running on $BACKEND_PORT" >&2
else
  echo "e2e-server: starting backend on $BACKEND_PORT..." >&2
  if [ ! -x "$ROOT_DIR/flagr" ]; then
    echo "e2e-server: make build..." >&2
    (cd "$ROOT_DIR" && make build) >&2
  fi
  "$ROOT_DIR/flagr" --port "$BACKEND_PORT" &
  for i in $(seq 1 30); do
    if curl -sf "http://127.0.0.1:$BACKEND_PORT/api/v1/health" > /dev/null 2>&1; then
      echo "e2e-server: backend ready" >&2
      break
    fi
    sleep 1
  done
  started_any=true
fi

# --- Frontend ---
if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT" 2>/dev/null; then
  echo "e2e-server: frontend already running on $FRONTEND_PORT" >&2
else
  echo "e2e-server: make run-ui on $FRONTEND_PORT..." >&2
  (cd "$ROOT_DIR" && make run-ui) &
  for i in $(seq 1 30); do
    if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT" 2>/dev/null; then
      echo "e2e-server: frontend ready" >&2
      break
    fi
    sleep 1
  done
  started_any=true
fi

if [ "$started_any" = true ]; then
  wait
fi