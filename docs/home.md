# Get Started

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides feature flags, experimentation (A/B testing), and dynamic configuration. It has clear swagger REST APIs for flags management and flag evaluation. For more details, see [Flagr Overview](flagr_overview)

## Run

Run directly with docker. https://hub.docker.com/r/checkr/flagr/


```bash
# Start the docker container
docker run -it -p 18000:18000 checkr/flagr
```

## Deploy

We recommend directly use the checkr/flagr image, and configure everything in the env variables. See more in [Server Configuration](flagr_env).

```bash
# Set env variables. For example,
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR=root:@tcp(127.0.0.1:18100)/flagr?parseTime=true

# Run the docker image. Ideally, the deploymenet will be handled by Kubernetes or Mesos.
docker run -it -p 18000:18000 checkr/flagr
```

## Development

Install Dependencies.

- Go
- Make (for Makefile)
- Yarn (for building UI)
- SQLite3 (for testing)

Build from source.

```bash
# get the source
go get -u github.com/checkr/flagr

# install dependencies, generated code, and run the app
cd $GOPATH/src/github.com/checkr/flagr
make all
```
