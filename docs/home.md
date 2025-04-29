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

- Go (1.21+)
- Make (for Makefile)
- NPM (for building UI)

Build from source.

```bash
# get the source
go get -u github.com/openflagr/flagr

# install dependencies, generated code, and start the service in
# development mode
cd $GOPATH/src/github.com/openflagr/flagr
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

## Running Flagr Locally

1. Update Dockerfile (follow comments in Dockerfile):
   - Change port from 18000 to 3000
   - set FLAGR_RECORDER_ENABLED to false for local testing if you dont want kafka setup
   - Add JWT secret and 
   - ENV FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL="http://localhost:3000/login"

2. Update frontend configuration:
   ```javascript
   // browser/src/constants.js
   DEV: {
       VUE_APP_API_URL: 'http://localhost:3000/api/v1',
       VUE_APP_SSO_API_URL: 'https://bff.allen-stage.in/internal-bff/',
   }
   ```

3. Build and run:
   ```bash
   # Build the Docker image
   docker build -t flagr .

   # Run the container
   docker run -it -p 3000:3000 flagr
   ```