# Server Configuration

Flagr is configured entirely through environment variables. See [env.go](https://github.com/openflagr/flagr/blob/master/pkg/config/env.go) for the full list.

[env.go](https://raw.githubusercontent.com/openflagr/flagr/master/pkg/config/env.go ':include :type=code')

```sh
# Example: set the database driver
export FLAGR_DB_DBDRIVER=mysql
# This sets Config.DBDriver = "mysql" at runtime
```

## Database drivers

| Driver | Use case |
|--------|----------|
| `mysql` | Production MySQL/PostgreSQL |
| `postgres` | Production PostgreSQL |
| `sqlite3` | Development (default, no external deps) |
| `json_file` | Load flags from a local JSON file ([format spec](flagr_json_flag_spec.md)) |
| `json_http` | Load flags from a URL (CI artifact, S3, GCS) |

For JSON-based workflows (GitOps, eval-only mode), see the [JSON Flag Source](flagr_json_flag_spec.md) guide.

### Basic Auth (web interface)

```sh
FLAGR_BASIC_AUTH_ENABLED=true
FLAGR_BASIC_AUTH_USERNAME=admin
FLAGR_BASIC_AUTH_PASSWORD=password
```

UI access prompts for username/password. API paths can be whitelisted to skip auth:

```sh
FLAGR_BASIC_AUTH_WHITELIST_PATHS="/api/v1/flags,/api/v1/evaluation"
FLAGR_BASIC_AUTH_EXACT_WHITELIST_PATHS=""
```

Note: Basic auth protects the web UI. It does not prevent direct API calls to `/api/v1/flags`.

### JWT Auth

See [env.go](https://github.com/openflagr/flagr/blob/master/pkg/config/env.go) for JWT configuration options.

## Data record destinations

### Kinesis (AWS)

Authenticate with standard AWS methods:

```sh
AWS_ACCESS_KEY_ID=example123
AWS_SECRET_ACCESS_KEY=example123
AWS_DEFAULT_REGION=eu-central-1
```

Other options include credentials files, container credentials, and instance profiles. See the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#config-settings-and-precedence).

Make sure the IAM key has permissions to push records to the Kinesis stream.

### Pub/Sub (Google Cloud)

For development:

```sh
gcloud auth application-default login
```

For production, create a service account and point to the key file:

```sh
FLAGR_RECORDER_PUBSUB_PROJECT_ID=google-project-id
FLAGR_RECORDER_PUBSUB_KEYFILE=/path/to/service/account.json
```

Alternatively, set `GOOGLE_APPLICATION_CREDENTIALS` (this affects all Google services in the environment).

## Datar

Datar is an optional in-memory aggregate analytics engine. Requires `FLAGR_RECORDER_ENABLED=true`. List `datar` in `FLAGR_RECORDER_TYPE` to enable it along with `FLAGR_RECORDER_DATAR_FLUSH_INTERVAL`. See [Datar](flagr_datar.md) for details.
