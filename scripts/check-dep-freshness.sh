#!/usr/bin/env bash
#
# check-dep-freshness.sh — Reject Go dependencies published too recently.
#
# Supply-chain attacks often land in fresh releases. This script checks every
# direct Go module dependency and fails if any resolved version was published
# within the last MIN_AGE_DAYS days.
#
# NOTE: npm cooldown is handled natively via .npmrc min-release-age=5.
# This script only covers Go modules, which have no native equivalent.
#
# Usage:
#   MIN_AGE_DAYS=5 ./scripts/check-dep-freshness.sh
#
# Exit codes:
#   0 — all deps are old enough
#   1 — at least one dep is too fresh
#   2 — tooling error

set -euo pipefail

MIN_AGE_DAYS="${MIN_AGE_DAYS:-5}"
TODAY_EPOCH=$(date +%s)
CUTOFF_EPOCH=$((TODAY_EPOCH - MIN_AGE_DAYS * 86400))
FAILURES=0

# ── Helpers ──────────────────────────────────────────────────────────────────

date_to_epoch() {
  date -d "$1" +%s 2>/dev/null || echo 0
}

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

# ── Go dependencies ──────────────────────────────────────────────────────────

check_go_deps() {
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

echo "Go dependency freshness check (min age: ${MIN_AGE_DAYS} days)"
echo "(npm cooldown handled by .npmrc min-release-age)"
echo ""

check_go_deps

echo ""
if [[ "$FAILURES" -gt 0 ]]; then
  echo "FAILED: ${FAILURES} Go dependencies are younger than ${MIN_AGE_DAYS} days."
  echo "Wait for them to age, or set MIN_AGE_DAYS=<N> to override."
  exit 1
else
  echo "PASSED: all Go dependencies are at least ${MIN_AGE_DAYS} days old."
fi
