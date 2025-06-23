<p align="center">
    <a href="https://github.com/Allen-Career-Institute/flagr/actions/workflows/ci.yml?query=branch%3Amain+" target="_blank">
        <img src="https://github.com/Allen-Career-Institute/flagr/actions/workflows/ci.yml/badge.svg?branch=main">
    </a>
    <a href="https://goreportcard.com/report/github.com/Allen-Career-Institute/flagr" target="_blank">
        <img src="https://goreportcard.com/badge/github.com/Allen-Career-Institute/flagr">
    </a>
    <a href="https://godoc.org/github.com/Allen-Career-Institute/flagr" target="_blank">
        <img src="https://img.shields.io/badge/godoc-reference-green.svg">
    </a>
    <a href="https://github.com/Allen-Career-Institute/flagr/releases" target="_blank">
        <img src="https://img.shields.io/github/release/Allen-Career-Institute/flagr.svg?style=flat&color=green">
    </a>
    <a href="https://codecov.io/gh/Allen-Career-Institute/flagr">
        <img src="https://codecov.io/gh/Allen-Career-Institute/flagr/branch/main/graph/badge.svg?token=iwjv26grrN">
    </a>
</p>

## Introduction
`Openflagr/flagr` is a community-driven OSS effort of advancing the development of Flagr.

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides feature flags, experimentation (A/B testing), and dynamic configuration. It has clear swagger REST APIs for flags management and flag evaluation.

## Documentation
- https://openflagr.github.io/flagr....

## Quick demo

Try it with Docker.

```sh
# Start the docker container
docker pull ghcr.io/Allen-Career-Institute/flagr
docker run -it -p 18000:18000 ghcr.io/Allen-Career-Institute/flagr

# Open the Flagr UI
open localhost:18000
```

Or try it on [https://try-flagr.onrender.com](https://try-flagr.onrender.com),
it may take a while for a cold start, and every commit to the `main` branch will trigger
a redeployment of the demo website.

```
curl --request POST \
     --url https://try-flagr.onrender.com/api/v1/evaluation \
     --header 'content-type: application/json' \
     --data '{
       "entityID": "127",
       "entityType": "user",
       "entityContext": {
         "state": "NY"
       },
       "flagID": 1,
       "enableDebug": true
     }'
```


## Flagr Evaluation Performance

Tested with `vegeta`. For more details, see [benchmarks](./benchmark).

```
Requests      [total, rate]            56521, 2000.04
Duration      [total, attack, wait]    28.2603654s, 28.259999871s, 365.529µs
Latencies     [mean, 50, 95, 99, max]  371.632µs, 327.991µs, 614.918µs, 1.385568ms, 12.50012ms
Bytes In      [total, mean]            23250552, 411.36
Bytes Out     [total, mean]            8308587, 147.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:56521
Error Set:
```

## Flagr UI

<p align="center">
    <img src="./docs/images/demo_readme.png" width="900">
</p>

## Client Libraries

| Language | Clients |
| -------- | ------- |
| Go | [goflagr](https://github.com/openflagr/goflagr) |
| Javascript | [jsflagr](https://github.com/openflagr/jsflagr) |
| Python | [pyflagr](https://github.com/openflagr/pyflagr) |
| Ruby | [rbflagr](https://github.com/openflagr/rbflagr) |

## License and Credit
- [`Allen-Career-Institute/flagr`](https://github.com/Allen-Career-Institute/flagr) Apache 2.0
- [`checkr/flagr`](https://github.com/checkr/flagr) Apache 2.0

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
