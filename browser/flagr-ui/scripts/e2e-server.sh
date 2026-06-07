#!/bin/sh
# Start backend + frontend for e2e tests. Idempotent — skips already-running services.
# Usage: ./scripts/e2e-server.sh
# Designed to be used as Playwright webServer command.

set -e

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
UI_DIR="$ROOT_DIR/browser/flagr-ui"
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
    echo "e2e-server: building backend binary..." >&2
    (cd "$ROOT_DIR" && go build -o flagr ./swagger_gen/cmd/flagr-server/) >&2
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
  echo "e2e-server: starting frontend on $FRONTEND_PORT..." >&2
  cd "$UI_DIR" && npx vite --port "$FRONTEND_PORT" &
  for i in $(seq 1 30); do
    if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT" 2>/dev/null; then
      echo "e2e-server: frontend ready" >&2
      break
    fi
    sleep 1
  done
  started_any=true
fi

# If we started servers, keep running so Playwright can manage the process.
# If servers already existed, exit immediately (Playwright will reuse).
if [ "$started_any" = true ]; then
  wait
fi
