# Server Configuration

Flagr is a single binary with no external runtime dependencies beyond its
database (and even that is optional). It follows the twelve-factor principle
of **configuration via the environment**: every knob — database driver, auth
strategy, recorder backends, cache intervals — is an environment variable,
parsed once at startup into a single config struct. There are no YAML files,
no config databases, no hot-reload daemons. This makes Flagr trivial to run
in a container, easy to configure via Kubernetes secrets, and identical
across environments when the env vars match. See
[env.go](https://github.com/openflagr/flagr/blob/master/pkg/config/env.go)
for the full list.

[env.go](https://raw.githubusercontent.com/openflagr/flagr/master/pkg/config/env.go ':include :type=code')

```sh
# Example: set the database driver
export FLAGR_DB_DBDRIVER=mysql
# Sets Config.DBDriver = "mysql" at runtime
```

## Database drivers

The driver chooses where flag state lives. For production you'll use MySQL or
PostgreSQL — a durable, shared store that survives restarts and scales across
instances. For local development SQLite needs no external process. The two
`json_*` drivers are a different mode entirely: they skip the database and
load flags from a file or URL, turning Flagr into an eval-only engine for
GitOps workflows where Git is the source of truth.

| Driver | Use case |
|--------|----------|
| `sqlite3` | Development and testing (default, no external deps) |
| `mysql` | Production MySQL |
| `postgres` | Production PostgreSQL |
| `json_file` | Load flags from a local JSON file ([format spec](flagr_json_flag_spec.md)) |
| `json_http` | Load flags from a URL (CI artifact, S3, GCS) |

For JSON-based workflows (GitOps, eval-only mode), see the
[JSON Flag Source](flagr_json_flag_spec.md) guide.

## Authentication

By default Flagr's API is open — convenient for a local dev instance behind a
firewall, dangerous in production. Flagr ships two opt-in auth layers: **Basic
Auth** for the web UI, and **JWT** for API consumers. Both use *whitelists*
rather than blanket protection: every path requires credentials *except*
those you explicitly open. This lets you keep evaluation public (so your app
can call it without tokens) while guarding flag mutations behind auth.

### Basic Auth (web interface)

```sh
FLAGR_BASIC_AUTH_ENABLED=true
FLAGR_BASIC_AUTH_USERNAME=admin
FLAGR_BASIC_AUTH_PASSWORD=password
```

UI access prompts for username/password. API paths can be whitelisted to skip
auth:

```sh
FLAGR_BASIC_AUTH_WHITELIST_PATHS="/api/v1/flags,/api/v1/evaluation"
FLAGR_BASIC_AUTH_EXACT_WHITELIST_PATHS=""
```

- `WHITELIST_PATHS` uses **prefix** matching (`/api/v1/flags` matches
  `/api/v1/flags/123`).
- `EXACT_WHITELIST_PATHS` uses exact path equality.

> **Note:** The **default** `FLAGR_BASIC_AUTH_WHITELIST_PATHS` is
> `/api/v1/health,/api/v1/flags,/api/v1/evaluation,/api/v1/exposures`, so by
> default the flag, evaluation, and exposure endpoints bypass basic auth.
> Basic auth applies to **every** path; only whitelisted paths skip it.
> Remove entries from the whitelist to require credentials on those API paths.

### JWT Auth

Flagr supports JWT-based authentication for API access. Configure the signing
key and algorithm via environment variables. Key options (see
[env.go](https://github.com/openflagr/flagr/blob/master/pkg/config/env.go) for
the full list):

| Variable | Default | Description |
|----------|---------|-------------|
| `FLAGR_JWT_AUTH_ENABLED` | `false` | Enable JWT auth |
| `FLAGR_JWT_AUTH_SECRET` | — | Signing secret (HS256/HS512) |
| `FLAGR_JWT_AUTH_SIGNING_METHOD` | `HS256` | `HS256`, `HS512`, or `RS256` |
| `FLAGR_JWT_AUTH_WHITELIST_PATHS` | `/api/v1/health,/api/v1/evaluation,/api/v1/exposures,/static` | Prefix-whitelisted paths (open when JWT auth is on) |
| `FLAGR_JWT_AUTH_EXACT_WHITELIST_PATHS` | `/,` | Exact-whitelisted paths |
| `FLAGR_JWT_AUTH_NO_TOKEN_STATUS_CODE` | `307` | Status when no token is present |
| `FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL` | — | Optional redirect URL |

## Data record destinations

Evaluation produces a decision; analytics needs a stream of *what happened*.
Flagr can emit **evaluation** and **exposure** rows to one or more backends —
the same event, fanned out to every recorder you enable. This is opt-in for a
reason: recording adds load, and not every flag needs analytics. You turn it
on globally with `FLAGR_RECORDER_ENABLED=true`, then per-flag with
`dataRecordsEnabled: true`, so a noisy internal flag never floods your
Kafka topic while your production experiments stream cleanly.

| Goal | Doc |
|------|-----|
| Exposure API (`POST /exposures`) | [Exposure Logging](flagr_exposure.md) |
| Stream eval + exposure to your pipeline (Kafka, Kinesis, or Pub/Sub) + A/B patterns | [Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md) |
| Built-in eval counts only (no stream, no exposures) | [Datar Analytics](flagr_datar.md) |

Exposure rows use `recordSource: "exposure"` on the same wire shape as
evaluations. `FLAGR_EXPOSURE_BATCH_SIZE` (default **100**) caps rows per
exposure request.

Set `FLAGR_RECORDER_TYPE` to a comma-separated list (e.g. `kafka`,
`kafka,datar`). Default: `kafka`.

### Kafka (default)

```sh
FLAGR_RECORDER_ENABLED=true
FLAGR_RECORDER_TYPE=kafka
FLAGR_RECORDER_KAFKA_BROKERS=kafka1:9092,kafka2:9092
FLAGR_RECORDER_KAFKA_TOPIC=flagr-records
```

Additional Kafka options include SSL/TLS (`FLAGR_RECORDER_KAFKA_CERTFILE`,
`FLAGR_RECORDER_KAFKA_KEYFILE`, `FLAGR_RECORDER_KAFKA_CAFILE`), SASL
authentication (`FLAGR_RECORDER_KAFKA_SASL_USERNAME`,
`FLAGR_RECORDER_KAFKA_SASL_PASSWORD`), idempotent producers, and encryption.
See [env.go](https://github.com/openflagr/flagr/blob/master/pkg/config/env.go)
for the full list.

For consuming eval and exposure events (any streaming recorder) and A/B
analysis, see [Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md).

### Kinesis (AWS)

Authenticate with standard AWS methods:

```sh
AWS_ACCESS_KEY_ID=example123
AWS_SECRET_ACCESS_KEY=example123
AWS_DEFAULT_REGION=eu-central-1
```

Other options include credentials files, container credentials, and instance
profiles. See the
[AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#config-settings-and-precedence).

Make sure the IAM key has permissions to push records to the Kinesis stream.

Kinesis and Pub/Sub use the same **record frame** as Kafka for both evaluation
and exposure rows. See
[Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md).

### Pub/Sub (Google Cloud)

For development:

```sh
gcloud auth application-default login
```

For production, create a service account and point to the key file:

```sh
FLAGR_RECORDER_PUBSUB_PROJECT_ID=google-project-id
FLAGR_RECORDER_PUBSUB_KEYFILE=/path/to/service/account.json
FLAGR_RECORDER_PUBSUB_TOPIC_NAME=flagr-records
```

Alternatively, set `GOOGLE_APPLICATION_CREDENTIALS` (this affects all Google
services in the environment).

### Datar (built-in analytics)

Datar is an optional in-memory aggregate analytics engine. List `datar` in
`FLAGR_RECORDER_TYPE` alongside other recorders, or use it alone for a
zero-dependency analytics setup:

```sh
FLAGR_RECORDER_ENABLED=true
FLAGR_RECORDER_TYPE=datar
FLAGR_RECORDER_DATAR_FLUSH_INTERVAL=60s
```

See [Datar Analytics](flagr_datar.md) for endpoint documentation, data model,
and resource usage.