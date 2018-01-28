# RateLimit

[![Build Status](https://travis-ci.org/bsm/ratelimit.png?branch=master)](https://travis-ci.org/bsm/ratelimit)
[![GoDoc](https://godoc.org/github.com/bsm/ratelimit?status.png)](http://godoc.org/github.com/bsm/ratelimit)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/ratelimit)](https://goreportcard.com/report/github.com/bsm/ratelimit)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Simple, thread-safe Go rate-limiter.
Inspired by Antti Huima's algorithm on http://stackoverflow.com/a/668327

### Example

```go
package main

import (
  "github.com/bsm/ratelimit"
  "log"
)

func main() {
  // Create a new rate-limiter, allowing up-to 10 calls
  // per second
  rl := ratelimit.New(10, time.Second)

  for i:=0; i<20; i++ {
    if rl.Limit() {
      fmt.Println("DOH! Over limit!")
    } else {
      fmt.Println("OK")
    }
  }
}
```

### Documentation

Full documentation is available on [GoDoc](http://godoc.org/github.com/bsm/ratelimit)
