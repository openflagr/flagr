[![Go Report Card](https://goreportcard.com/badge/github.com/checkr/flagr)](https://goreportcard.com/report/github.com/checkr/flagr)

# Flagr Quickstart Guide

Flagr delivers the right experience to the right entity and monitors the impact. It’s a micro service that provides the functionality of feature flags, experimentation (A/B testing), and dynamic configuration.

## Install from Source

Source installation is only intended for developers and advanced users.

```sh
# get the source
go get -u github.com/checkr/flagr
cd $GOPATH/src/github.com/checkr/flagr

# docker-compose with the infra
docker-compose up

# install dependencies, generated code, and run the app
make all
```

## Test using Flagr UI
Flagr Server comes with an embedded web based UI. Point your web browser to http://127.0.0.1:18000 ensure your server has started successfully.

## Explore Further
- [Flagr QuickStart Guide]()
- [Flagr Documentation Website]()

## Contribute to Flagr Project
Please follow Flagr [Contributor's Guide](https://github.com/checkr/flagr/blob/master/CONTRIBUTING.md)
