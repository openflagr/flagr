# Introduction

## What is Flagr

Flagr delivers the right experience to the right entity and monitors the impact. It’s a microservice that handles feature flagging, experimentation (A/B testing), and dynamic configuration. Flagr is designed from the ground up to serve high volume traffic of feature flag and A/B testing evaluation requests. Flagr is also perfectly capable of powering sophisticated feature rollout when used in combination with all the constraints it supports.  The core of Flagr is focused on the user segmentation, constraints setting, and high performance of evaluation.

## Get Started

The easiest way to try out Flagr is using the flagr-mini docker image.

```
# Start the docker container
docker run -it -p 18000:18000 checkr/flagr

# Or with attached volume
docker run -it -p 18000:18000 -v /tmp/flagr_data:/data checkr/flagr
```

And then open http://localhost:18000

## Ready for More?
