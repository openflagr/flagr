# Get Started

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides feature flags, experimentation (A/B testing), and dynamic configuration. It has clear swagger REST APIs for flags management and flag evaluation. For more details, see [Flagr Overview](flagr_overview)

## Run

Run directly with docker.

```bash
# Start the docker container
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

# Open the Flagr UI
open localhost:18000
```

## Deploy

We recommend directly use the openflagr/flagr image, and configure everything in the env variables. See more in [Server Configuration](flagr_env).

```bash
# Set env variables. For example,
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR=root:@tcp(127.0.0.1:18100)/flagr?parseTime=true

# Run the docker image. Ideally, the deployment will be handled by Kubernetes or Mesos.
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr
```

## Development

Install Dependencies.

- Go (1.24+)
- Make (for Makefile)
- Node (20+) (for building UI)

Build from source.

```bash
# get the source
git clone https://github.com/openflagr/flagr.git
cd flagr

# install dependencies, generate code, and start the service in
# development mode
make build start
```

If you just want to run the pre-built backend (without the UI development service):

```
make run
```

And alternatively to just run the UI service:

```
make run_ui
```

## Testing

Flagr has three kinds of tests, each serving a different purpose.

### Unit tests

Run the Go unit tests (no external services required):

```bash
make test
```

Or directly:

```bash
go test ./pkg/...
```

### E2E tests (Flagr UI)

Playwright-based end-to-end tests for the Vue 3 UI. Builds the Go
server, starts the backend and UI servers, runs Playwright, then cleans up:

```bash
make test-e2e
```

### Integration tests (API, multi-DB)

HTTP-level integration tests covering all CRUD and eval endpoints.
Seeds ~50 realistic flags across all 12 constraint operators.

**Local mode** — SQLite `:memory:`, auto-starts server on random port:

```bash
make test-integration
```

**Docker Compose mode** — runs the same test suite against 6 flagr
instances (SQLite, MySQL, MySQL 8, PostgreSQL 9, PostgreSQL 13,
checkr/flagr):

```bash
cd integration_tests && make test
```

**HTTP eval benchmarks** — measures end-to-end eval latency through HTTP:

```bash
make bench-integration
```

To run against a single Docker Compose instance:

```bash
cd integration_tests && make test-instance INSTANCE=flagr_with_mysql
```
