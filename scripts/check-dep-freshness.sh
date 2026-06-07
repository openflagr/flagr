#!/usr/bin/env bash
#
# check-dep-freshness.sh — Reject dependencies published too recently.
#
# Supply-chain attacks often land in fresh releases. This script checks every
# production dependency (npm + Go) and fails if any resolved version was
# published within the last MIN_AGE_DAYS days.
#
# Usage:
#   MIN_AGE_DAYS=14 ./scripts/check-dep-freshness.sh
#
# Exit codes:
#   0 — all deps are old enough
#   1 — at least one dep is too fresh
#   2 — tooling error (npm / go not found)

set -euo pipefail

MIN_AGE_DAYS="${MIN_AGE_DAYS:-14}"
TODAY_EPOCH=$(date +%s)
CUTOFF_EPOCH=$((TODAY_EPOCH - MIN_AGE_DAYS * 86400))
FAILURES=0

# ── Helpers ──────────────────────────────────────────────────────────────────

date_to_epoch() {
  date -d "$1" +%s 2>/dev/null || echo 0
}

# check_freshness <name> <version> <publish_date_iso>
check_freshness() {
  local name="$1" version="$2" pub_date="$3"
  local pub_epoch
  pub_epoch=$(date_to_epoch "$pub_date")
  if [[ "$pub_epoch" -eq 0 ]]; then
    echo "  ⚠  $name@$version — could not parse date: $pub_date"
    return
  fi
  local age_days=$(( (TODAY_EPOCH - pub_epoch) / 86400 ))
  if [[ "$pub_epoch" -gt "$CUTOFF_EPOCH" ]]; then
    echo "  ❌  $name@$version — published ${age_days}d ago ($pub_date) — younger than ${MIN_AGE_DAYS}d threshold"
    FAILURES=$((FAILURES + 1))
  else
    echo "  ✅  $name@$version — published ${age_days}d ago ($pub_date)"
  fi
}

# get_npm_publish_date <pkg> <version>
# Returns ISO date string or empty on failure.
get_npm_publish_date() {
  local pkg="$1" ver="$2"
  # Fetch full time object, extract the version key
  npm view "$pkg" time --json 2>/dev/null | node -e "
    const d = JSON.parse(require('fs').readFileSync('/dev/stdin','utf8'));
    process.stdout.write(d['$ver'] || '');
  " 2>/dev/null || true
}

# ── npm dependencies ─────────────────────────────────────────────────────────

check_npm_deps() {
  echo ""
  echo "=== npm production dependencies (browser/flagr-ui) ==="
  if [[ ! -f browser/flagr-ui/package.json ]]; then
    echo "  skipped — package.json not found"
    return
  fi

  cd browser/flagr-ui

  # Get installed prod deps as "name@version" lines
  local deps
  deps=$(npm ls --prod --json 2>/dev/null | node -e "
    const data = JSON.parse(require('fs').readFileSync('/dev/stdin','utf8'));
    for (const [name, info] of Object.entries(data.dependencies || {})) {
      if (info.version && !info.version.startsWith('file:') && !info.version.startsWith('link:')) {
        console.log(name + ' ' + info.version);
      }
    }
  " 2>/dev/null || true)

  if [[ -z "$deps" ]]; then
    echo "  skipped — no production deps found"
    cd - >/dev/null
    return
  fi

  while IFS=' ' read -r name version; do
    [[ -z "$name" ]] && continue
    local pub_date
    pub_date=$(get_npm_publish_date "$name" "$version")
    if [[ -n "$pub_date" ]]; then
      check_freshness "$name" "$version" "$pub_date"
    else
      echo "  ⚠  $name@$version — could not fetch publish date"
    fi
  done <<< "$deps"

  cd - >/dev/null
}

# ── Go dependencies ──────────────────────────────────────────────────────────

check_go_deps() {
  echo ""
  echo "=== Go module dependencies ==="
  if [[ ! -f go.mod ]]; then
    echo "  skipped — go.mod not found"
    return
  fi

  # -mod=mod bypasses vendor directory
  local modules
  modules=$(go list -mod=mod -m -json all 2>/dev/null | node -e "
    const chunks = require('fs').readFileSync('/dev/stdin','utf8').split('\n}\n');
    for (let i = 0; i < chunks.length; i++) {
      let s = chunks[i].trim();
      if (!s.endsWith('}')) s += '}';
      try {
        const m = JSON.parse(s);
        if (m.Version && !m.Main && !m.Indirect) {
          console.log(m.Path + ' ' + m.Version);
        }
      } catch {}
    }
  " 2>/dev/null || true)

  if [[ -z "$modules" ]]; then
    echo "  skipped — no Go modules found"
    return
  fi

  while IFS=' ' read -r name version; do
    [[ -z "$name" ]] && continue
    # Skip pseudo-versions and +incompatible (pre-modules era, no proxy metadata)
    [[ "$version" == v0.0.0-* ]] && continue
    [[ "$version" == *"+incompatible" ]] && continue

    local pub_date
    pub_date=$(curl -sf "https://proxy.golang.org/${name}/@v/${version}.info" 2>/dev/null | node -e "
      const d = JSON.parse(require('fs').readFileSync('/dev/stdin','utf8'));
      process.stdout.write(d.Time || '');
    " 2>/dev/null || true)

    if [[ -n "$pub_date" ]]; then
      check_freshness "$name" "$version" "$pub_date"
    else
      echo "  ⚠  $name@$version — could not fetch publish date from proxy"
    fi
  done <<< "$modules"
}

# ── Main ─────────────────────────────────────────────────────────────────────

echo "Dependency freshness check (min age: ${MIN_AGE_DAYS} days)"

check_npm_deps
check_go_deps

echo ""
if [[ "$FAILURES" -gt 0 ]]; then
  echo "FAILED: ${FAILURES} dependencies are younger than ${MIN_AGE_DAYS} days."
  echo "Wait for them to age, or set MIN_AGE_DAYS=<N> to override."
  exit 1
else
  echo "PASSED: all dependencies are at least ${MIN_AGE_DAYS} days old."
fi
