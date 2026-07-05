# Built-in Context Injection

Flagr can automatically inject server-side and HTTP request metadata into every
evaluation's `entityContext`. This lets you write **constraint rules** against
time, deployment environment, or proxy-injected headers — without changing your
application code or managing separate flag configurations per environment.

Once enabled, built-in keys appear in the entity context alongside whatever
your application sends. Constraints reference them like any other property.

## Quick start

```sh
# Enable built-in context injection
FLAGR_INJECTED_CONTEXT_ENABLED=true

# Expose specific headers as context keys
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS="X-Environment,X-Tenant-ID"

# Or expose all headers matching a prefix (e.g., all Cloudflare headers)
FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES="CF-"
```

Restart Flagr. Now every evaluation automatically includes:

| Key | Value | Source |
|-----|-------|--------|
| `@ts` | `1751666400` | Unix epoch seconds (server time) |
| `@ts_hour` | `14` | Hour of day, 0–23 UTC |
| `@ts_weekday` | `1` | Day of week, 0=Sunday–6=Saturday |
| `@ts_month` | `7` | Month, 1–12 |
| `@http_x_environment` | `"production"` | From `X-Environment` header |
| `@http_x_tenant_id` | `"acme-corp"` | From `X-Tenant-ID` header |
| `@http_cf_ipcountry` | `"US"` | From `CF-IPCountry` header |

Your application still sends the same `entityContext` — the built-in keys are
merged in server-side before evaluation.

---

## Configuration reference

| Variable | Default | Description |
|----------|---------|-------------|
| `FLAGR_INJECTED_CONTEXT_ENABLED` | `false` | Master switch. When `false`, no injection occurs. |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADERS` | `""` | Comma-separated header names to expose as `@http_*` keys. Case-insensitive. |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES` | `""` | Comma-separated prefixes. Any header starting with these is injected. |

### Naming convention

Built-in keys follow a simple rule:

- **Core keys:** `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month` — always present
  when enabled, no configuration needed.
- **Header keys:** `@http_<header_name>` — header name lowercased, `-` replaced
  with `_`. Example: `X-Environment` → `@http_x_environment`.

The `@` prefix marks server-injected keys. Client-provided `entityContext`
keys have no prefix. This prevents collisions.

### Case insensitivity

Header names in config are case-insensitive. These are equivalent:

```sh
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS="X-Environment"
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS="x-environment"
```

---

## Real-world examples

### 1. Single Flagr instance, multiple environments

**Problem:** You have one Flagr instance serving staging, canary, and
production. You need different flag behavior per environment without
separate deployments.

**Solution:** Your load balancer or reverse proxy sets an `X-Environment`
header. Flagr reads it and makes it available as `@http_x_environment`.

**Proxy config (nginx example):**

```nginx
# Staging backend
location /staging/ {
    proxy_set_header X-Environment staging;
    proxy_pass http://flagr-backend;
}

# Production backend
location / {
    proxy_set_header X-Environment production;
    proxy_pass http://flagr-backend;
}
```

**Flag configuration:**

```
Flag: "new-checkout-flow"
Variants:
  - on
  - off

Segment 1 (staging only):
  Constraint: {@http_x_environment} EQ "staging"
  Rollout: 100%
  Distribution: on 100%, off 0%

Segment 2 (production gradual rollout):
  Constraint: {@http_x_environment} EQ "production"
  Rollout: 10%
  Distribution: on 100%, off 0%

Segment 3 (default — off everywhere else):
  Rollout: 100%
  Distribution: on 0%, off 100%
```

**Result:**
- Staging: flag is ON for 100% of traffic
- Production: flag is ON for 10% of traffic (gradual rollout)
- Everything else: flag is OFF

No separate flag configs. No separate databases. One flag, environment-aware
constraints.

---

### 2. Scheduled feature launch

**Problem:** You want to enable a feature at midnight on launch day, without
staying up to flip the switch manually.

**Solution:** Use `@ts` (Unix epoch seconds) in a constraint. The flag
activates automatically when server time passes the threshold.

**Flag configuration:**

```
Flag: "black-friday-banner"
Variants:
  - show
  - hide

Segment 1 (active during Black Friday week):
  Constraint: {@ts} GTE 1764038400    ← Nov 24, 2025 00:00:00 UTC
  Constraint: {@ts} LT 1764643200     ← Dec 1, 2025 00:00:00 UTC
  Rollout: 100%
  Distribution: show 100%, hide 0%

Segment 2 (default — hidden):
  Rollout: 100%
  Distribution: show 0%, hide 100%
```

**Result:**
- Before Nov 24: banner hidden
- Nov 24 – Nov 30: banner shown
- After Dec 1: banner hidden again

The UI shows a human-readable hint when you enter `@ts` constraints:
`{1764038400 = Nov 24, 2025 00:00:00 UTC}`.

---

### 3. Business hours targeting

**Problem:** Enable a feature only during business hours (9 AM – 5 PM UTC),
or run different experiments during work hours vs off-hours.

**Solution:** Use `@ts_hour` for hour-of-day targeting.

**Flag configuration:**

```
Flag: "live-chat-support"
Variants:
  - on
  - off

Segment 1 (business hours):
  Constraint: {@ts_weekday} GTE 1      ← Monday
  Constraint: {@ts_weekday} LE 5       ← Friday
  Constraint: {@ts_hour} GTE 9         ← 9 AM UTC
  Constraint: {@ts_hour} LT 17         ← 5 PM UTC
  Rollout: 100%
  Distribution: on 100%, off 0%

Segment 2 (default — off):
  Rollout: 100%
  Distribution: on 0%, off 100%
```

**Result:**
- Weekdays 9–17 UTC: live chat enabled
- Weekends and off-hours: live chat disabled

Combine with `@ts_month` for seasonal features:

```
Constraint: {@ts_month} EQ 12          ← December only
Constraint: {@ts_month} EQ 1           ← or January
```

---

### 4. Multi-tenant feature gating

**Problem:** Different customers (tenants) should see different features.
Your API gateway sets a `X-Tenant-ID` header.

**Solution:** Use `@http_x_tenant_id` in constraints.

**Proxy config:**

```nginx
# Your API gateway sets this based on the authenticated tenant
proxy_set_header X-Tenant-ID "acme-corp";
```

**Flag configuration:**

```
Flag: "advanced-analytics-dashboard"
Variants:
  - premium
  - basic

Segment 1 (premium tenants):
  Constraint: {@http_x_tenant_id} IN "acme-corp,globex-inc,initech"
  Rollout: 100%
  Distribution: premium 100%, basic 0%

Segment 2 (default — basic for everyone else):
  Rollout: 100%
  Distribution: premium 0%, basic 100%
```

**Result:**
- Acme, Globex, Initech: see advanced analytics dashboard
- Everyone else: see basic dashboard

---

### 5. Country-based targeting with Cloudflare

**Problem:** You're behind Cloudflare and want to enable features for
specific countries (e.g., GDPR compliance, regional launches).

**Solution:** Cloudflare automatically sets `CF-IPCountry`. Expose it with
the prefix config.

**Flag config:**

```sh
FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES="CF-"
```

**Flag configuration:**

```
Flag: "eu-cookie-consent-banner"
Variants:
  - show
  - hide

Segment 1 (EU countries):
  Constraint: {@http_cf_ipcountry} IN "AT,BE,BG,HR,CY,CZ,DK,EE,FI,FR,DE,GR,HU,IE,IT,LV,LT,LU,MT,NL,PL,PT,RO,SK,SI,ES,SE"
  Rollout: 100%
  Distribution: show 100%, hide 0%

Segment 2 (default — no banner):
  Rollout: 100%
  Distribution: show 0%, hide 100%
```

**Result:**
- EU visitors: see cookie consent banner
- Everyone else: no banner

---

### 6. Canary deployment with custom headers

**Problem:** Your service mesh sets `X-Canary: true` on traffic routed to
canary instances. You want canary instances to test new features.

**Solution:** Use `@http_x_canary` in constraints.

**Flag configuration:**

```
Flag: "redesigned-homepage"
Variants:
  - new
  - old

Segment 1 (canary instances):
  Constraint: {@http_x_canary} EQ "true"
  Rollout: 100%
  Distribution: new 100%, old 0%

Segment 2 (default — old for everyone):
  Rollout: 100%
  Distribution: new 0%, old 100%
```

---

## How it works

### Injection flow

```
Application sends:  {"country": "US", "tier": "premium"}
                         ↓
Flagr injects:      {"country": "US", "tier": "premium",
                      "@ts": 1751666400,
                      "@ts_hour": 14,
                      "@ts_weekday": 1,
                      "@ts_month": 7,
                      "@http_x_environment": "production"}
                         ↓
Constraint engine:  evaluates {@ts_hour} GTE 9 AND {@ts_hour} LT 17
                         ↓
Result:             segment matches (it's 2 PM UTC)
```

### Performance

Injection adds ~1.2µs to a 328µs evaluation — negligible overhead.
The `@ts` keys are computed from `time.Now().UTC()` (1 allocation for the map).
Header matching uses `sync.Once`-cached sets — the config is parsed once, not
per-request. Header injection iterates `r.Header` directly (no `Clone()`).
When the feature is disabled, `InjectBuiltInContext` returns the original
context unchanged (zero cost).

### Namespace isolation

| Prefix | Meaning | Example |
|--------|---------|---------|
| `@ts` | Server time (always injected) | `@ts`, `@ts_hour` |
| `@http_` | HTTP header (configurable) | `@http_x_environment` |
| *(none)* | Client-provided | `country`, `tier` |

Server-injected keys (`@` prefix) overwrite client-provided values with the
same name. This prevents spoofing — a client cannot fake `@ts` to bypass
scheduling constraints.

---

## FAQ

**Q: Can I use built-in keys in the debug console?**

A: Yes. Open the Debug Console, enter an eval request with empty
`entityContext: {}`, and enable debug mode. The response will show the
injected keys in `evalDebugLog`.

**Q: What happens if a header is empty?**

A: Empty headers are skipped — they don't appear in the context.

**Q: What if I configure a header that doesn't exist in the request?**

A: Silent no-op. The key simply isn't injected.

**Q: Can I override a built-in key from my application?**

A: No. Server-injected keys always overwrite client values. If your
application sends `{"@ts": 0}`, it gets overwritten with the real server
time. This is intentional — it prevents spoofing.

**Q: Does this work with batch evaluation?**

A: Yes. Each entity in a batch request gets built-in context injected
individually — same semantics as the single-eval path (in-place mutation
when the context is already a map). When disabled, no injection occurs.
