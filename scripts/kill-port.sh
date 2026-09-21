#!/bin/sh
# Free TCP listen ports by PID.
# Unix: lsof or fuser. Windows (Git Bash): netstat + kill/taskkill.
# Used by `make stop-ui` and e2e-server.sh — do not pkill by process name.

set -eu

kill_pid() {
	pid=$1
	case "$pid" in
	'' | 0 | 4) return 0 ;;
	esac
	if kill "$pid" >/dev/null 2>&1; then
		return 0
	fi
	if command -v taskkill >/dev/null 2>&1; then
		taskkill //F //PID "$pid" >/dev/null 2>&1 || true
	else
		kill -9 "$pid" >/dev/null 2>&1 || true
	fi
}

pids_listening() {
	port=$1
	if command -v lsof >/dev/null 2>&1; then
		lsof -ti:"$port" 2>/dev/null || true
		return 0
	fi
	if command -v fuser >/dev/null 2>&1; then
		# fuser prints "18000/tcp: 1234 5678"
		fuser -n tcp "$port" 2>/dev/null | awk '{ for (i = 1; i <= NF; i++) if ($i ~ /^[0-9]+$/) print $i }'
		return 0
	fi
	# netstat localizes the state word (LISTENING, ABHÖREN, ÉCOUTE).
	# A TCP listener's foreign address is 0.0.0.0:0 or [::]:0 in every locale.
	netstat -ano 2>/dev/null | awk -v port="$port" '
		{
			foreign = $3
			if (foreign != "0.0.0.0:0" && foreign != "[::]:0") next
			addr = $2
			sub(/.*:/, "", addr)
			if (addr == port) print $NF
		}
	'
}

for port in "$@"; do
	for pid in $(pids_listening "$port" | sort -u); do
		kill_pid "$pid"
	done
done
