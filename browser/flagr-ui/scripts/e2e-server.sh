#!/bin/sh
# Start backend + frontend for e2e tests. Idempotent — skips already-running services.
# Repo root: make build (backend), make run-ui (frontend). Playwright webServer entrypoint.

set -e

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
BACKEND_PORT=18000
FRONTEND_PORT=8080
KILL_PORT="$ROOT_DIR/scripts/kill-port.sh"
LOG_DIR="${TMPDIR:-/tmp}"
BACKEND_LOG="$LOG_DIR/flagr-e2e-backend.log"
FRONTEND_LOG="$LOG_DIR/flagr-e2e-frontend.log"

# WSL and Git Bash can both see a Windows checkout. Prefer flagr.exe only in
# Git Bash, or when the Unix binary is absent.
flagr_bin() {
	case "$(uname -s 2>/dev/null || true)" in
	MINGW* | MSYS* | CYGWIN*)
		if [ -f "$ROOT_DIR/flagr.exe" ]; then
			echo "$ROOT_DIR/flagr.exe"
			return
		fi
		;;
	esac
	if [ -f "$ROOT_DIR/flagr" ]; then
		echo "$ROOT_DIR/flagr"
	else
		echo "$ROOT_DIR/flagr.exe"
	fi
}

backend_pid=""
frontend_pid=""

cleanup() {
	trap - INT TERM EXIT
	[ -n "$backend_pid" ] && kill "$backend_pid" 2>/dev/null || true
	[ -n "$frontend_pid" ] && kill "$frontend_pid" 2>/dev/null || true
	sh "$KILL_PORT" "$BACKEND_PORT" "$FRONTEND_PORT" 2>/dev/null || true
	exit 0
}
trap cleanup INT TERM EXIT

started_any=false

# --- Backend ---
if curl -sf "http://127.0.0.1:$BACKEND_PORT/api/v1/health" > /dev/null 2>&1; then
	echo "e2e-server: backend already running on $BACKEND_PORT" >&2
else
	echo "e2e-server: starting backend on $BACKEND_PORT..." >&2
	BIN="$(flagr_bin)"
	if [ ! -f "$BIN" ]; then
		echo "e2e-server: make build..." >&2
		(cd "$ROOT_DIR" && make build) >&2
		BIN="$(flagr_bin)"
	fi
	# e2e runs several Playwright workers against one SQLite file. WAL + a busy
	# timeout + immediate write transactions keep concurrent writers waiting
	# instead of failing with SQLITE_BUSY.
	if [ -z "${FLAGR_DB_DBCONNECTIONSTR:-}" ]; then
		FLAGR_DB_DBCONNECTIONSTR="file:flagr.sqlite?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_txlock=immediate"
		export FLAGR_DB_DBCONNECTIONSTR
	fi
	# Redirect to a file: a surviving grandchild must not hold Playwright's
	# piped stdout open, or its Windows webServer teardown hangs forever.
	"$BIN" --port "$BACKEND_PORT" >"$BACKEND_LOG" 2>&1 &
	backend_pid=$!
	i=1
	while [ "$i" -le 30 ]; do
		if curl -sf "http://127.0.0.1:$BACKEND_PORT/api/v1/health" > /dev/null 2>&1; then
			echo "e2e-server: backend ready (pid $backend_pid, log $BACKEND_LOG)" >&2
			break
		fi
		sleep 1
		i=$((i + 1))
	done
	started_any=true
fi

# --- Frontend ---
if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT" 2>/dev/null; then
	echo "e2e-server: frontend already running on $FRONTEND_PORT" >&2
else
	echo "e2e-server: make run-ui on $FRONTEND_PORT..." >&2
	(cd "$ROOT_DIR" && make run-ui) >"$FRONTEND_LOG" 2>&1 &
	frontend_pid=$!
	i=1
	while [ "$i" -le 30 ]; do
		if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT" 2>/dev/null; then
			echo "e2e-server: frontend ready (pid $frontend_pid, log $FRONTEND_LOG)" >&2
			break
		fi
		sleep 1
		i=$((i + 1))
	done
	started_any=true
fi

if [ "$started_any" = true ]; then
	# Keep the process alive for Playwright; a signal-interruptible loop is more
	# portable than `wait`, which can swallow the trap on Git Bash.
	while :; do
		sleep 1
	done
fi
