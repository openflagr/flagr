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

- Go (1.26+)
- Make (for Makefile)
- Node (20+) (for building UI)

Build from source.

```bash
# get the source
git clone https://github.com/openflagr/flagr.git
cd flagr

# install dependencies, generate code, and start the service in
# development mode (backend + frontend dev server)
make build start

# Just run the pre-built backend (without UI dev server):
make run

# Just run the UI dev server (Vite on :8080, proxies API to :18000):
cd browser/flagr-ui && npm run dev

# After backend changes, rebuild and restart in one step:
make rebuild-run

# Run e2e tests (builds Go binary, starts servers, runs Playwright):
make test-e2e

# Stop dev servers (port-based, safe for multi-project setups):
make stop-ui
