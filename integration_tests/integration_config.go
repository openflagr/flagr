//go:build integration

package flagr_integration

import "time"

// Local auto-start server env (startLocalServer) — keep in sync with these durations.
const (
	integrationDatarFlushInterval     = 500 * time.Millisecond
	integrationEvalCacheRefresh       = time.Second
	integrationDatarFlushWaitMargin   = 100 * time.Millisecond
	pollInterval                      = 500 * time.Millisecond
	datarPollEvalsPerAttempt          = 5
	datarPollTimeout                  = 30 * time.Second
	exposureRecorderGateTimeout       = 15 * time.Second
	serverHealthWaitTimeout           = 30 * time.Second
	evalCacheReadyTimeout             = 20 * time.Second
	integrationHTTPClientTimeout      = 10 * time.Second
	integrationServerProcessWaitDelay = 5 * time.Second
)

// integrationDatarFlushWait is how long Datar tests sleep after posting evals so the
// in-process flush (FLAGR_RECORDER_DATAR_FLUSH_INTERVAL) can persist hourly_events.
func integrationDatarFlushWait() time.Duration {
	return integrationDatarFlushInterval + integrationDatarFlushWaitMargin
}

// Built-in context integration test constants.
const (
	// builtinCtxEvalCacheWait waits for the eval cache to pick up newly created
	// flags/constraints. The local auto-start server uses 1s refresh
	// (integrationEvalCacheRefresh), but docker-compose backends use the default
	// 3s (FLAGR_EVALCACHE_REFRESHINTERVAL envDefault). Use 2× the default to
	// account for DB latency on postgres/mysql/etc.
	builtinCtxEvalCacheWait = 6 * time.Second

	builtinCtxRolloutPercent = 100 // full rollout for constraint-only segments
	builtinCtxVariantOn      = "on"
	builtinCtxVariantEnabled = "enabled"
	builtinCtxEnvHeaderValue = "test"
)
