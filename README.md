<p align="center">
    <img src="./docs/images/logo.png" width="150">
</p>

<p align="center">
    <a href="https://goreportcard.com/report/github.com/checkr/flagr" target="_blank">
        <img src="https://goreportcard.com/badge/github.com/checkr/flagr">
    </a>
    <a href="https://circleci.com/gh/checkr/flagr" target="_blank">
        <img src="https://circleci.com/gh/checkr/flagr.svg?style=shield">
    </a>
    <a href="https://godoc.org/github.com/checkr/flagr" target="_blank">
        <img src="https://img.shields.io/badge/godoc-reference-green.svg">
    </a>
    <a href="https://github.com/checkr/flagr/releases" target="_blank">
        <img src="https://img.shields.io/github/release/checkr/flagr.svg?style=flat&color=green">
    </a>
    <a href="https://codecov.io/gh/checkr/flagr" target="_blank">
        <img src="https://codecov.io/gh/checkr/flagr/branch/master/graph/badge.svg">
    </a>
    <a href="https://hub.docker.com/r/checkr/flagr" target="_blank">
        <img src="https://github.com/checkr/flagr/workflows/Publish%20DockerHub/badge.svg?branch=master&event=release">
    </a>
</p>

## Introduction

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides feature flags, experimentation (A/B testing), and dynamic configuration. It has clear swagger REST APIs for flags management and flag evaluation.

## Documentation
- [Introducing Flagr Blog](https://engineering.checkr.com/introducing-flagr-a-robust-high-performance-service-for-feature-flagging-and-a-b-testing-f037c219b7d5)
- [Documentation](https://checkr.github.io/flagr/)

## Quick demo

Try it with Docker.

```sh
# Start the docker container
docker pull checkr/flagr
docker run -it -p 18000:18000 checkr/flagr

# Open the Flagr UI
open localhost:18000
```

Or try it on [https://try-flagr.herokuapp.com](https://try-flagr.herokuapp.com), it may take a while for a cold start.

```
curl --request POST \
     --url https://try-flagr.herokuapp.com/api/v1/evaluation \
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
| Go | [goflagr](https://github.com/checkr/goflagr) |
| Javascript | [jsflagr](https://github.com/checkr/jsflagr) |
| Python | [pyflagr](https://github.com/checkr/pyflagr) |
| Ruby | [rbflagr](https://github.com/checkr/rbflagr) |


# put flag templates
1. set env FLAGR_NEW_FLAG_TEMPLATES
2. create a flag by POST at /api/v1/flags where the description is a JSON body representing parameters for the template in FLAGR_NEW_FLAG_TEMPLATES, and `template` is the name of the template.
3. Observe that tags are created on the flag if the template included tags.
example:
```
FLAGR_NEW_FLAG_TEMPLATES='{"banana": "{\"dataRecordsEnabled\": false,    \"description\": \"This is my bestest flag\",    \"enabled\": false,    \"id\": 4,    \"key\": \"kxibnoxp4qqcwj3e7\",    \"segments\": [],    \"tags\": [ {            \"id\": 1,            \"value\": \"{{.TagVal}}\"        }],    \"updatedAt\": \"2020-12-23T15:19:17.909-07:00\",    \"variants\": []}"}' FLAGR_NEW_FLAG_TEMPLATE= PORT=18000 ./flagr
curl --location --request POST "http://localhost:18000/api/v1/flags" \
    --header "Connection: keep-alive" \
    --header "Accept: application/json, text/plain, */*" \
    --header "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36" \
    --header "Content-Type: application/json;charset=UTF-8" \
    --header "Origin: https://flagr.checkrhq.net" \
    --header "Sec-Fetch-Site: same-origin" \
    --header "Sec-Fetch-Mode: cors" \
    --header "Sec-Fetch-Dest: empty" \
    --header "Referer: https://flagr.checkrhq.net/" \
    --header "Accept-Language: en-US,en;q=0.9" \
    --data "{
    \"description\": \"{\\\"TagVal\\\": \\\"jkz\\\"}\",
    \"template\": \"banana\"
}"
# observe tag is created
curl http://localhost:18000/api/v1/flags/4
# observe tag is still there
```

TODO: need to templatize / autofill IDs somehow.  Also tag uniqueness?