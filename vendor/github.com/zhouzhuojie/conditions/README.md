# conditions

This package offers a parser of a simple conditions specification language (reduced set of arithmetic/logical operations). The package is mainly created for Flow-Based Programming components that require configuration to perform some operations on the data received from multiple input ports. But it can be used whereever you need externally define some logical conditions on the internal variables.

Additional credits for this package go to [Handwritten Parsers & Lexers in Go](http://blog.gopheracademy.com/advent-2014/parsers-lexers/) by Ben Johnson on [Gopher Academy blog](http://blog.gopheracademy.com) and [InfluxML package from InfluxDB repository](https://github.com/influxdb/influxdb/tree/master/influxql).

## Usage example 
```
package main

import (
    "fmt"
    "strings"

    "github.com/zhouzhuojie/conditions"
)

func main() {
    // Our condition to check
    s := `({foo} > 0.45) AND ({bar} == "ON" OR {baz} IN ["ACTIVE", "CLEAR"])`

    // Parse the condition language and get expression
    p := conditions.NewParser(strings.NewReader(s))
    expr, err := p.Parse()
    if err != nil {
        // ...
    }

    // Evaluate expression passing data for $vars
    data := map[string]interface{}{"foo": 0.12, "bar": "OFF", "baz": "ACTIVE"}
    r, err := conditions.Evaluate(expr, data)
    if err != nil {
        // ...
    }

    // r is false
    fmt.Println("Evaluation result:", r)
}
```

## Credit
Forked from [https://github.com/oleksandr/conditions](https://github.com/oleksandr/conditions)

The main differences are

- Changed the syntax of variables from `[foo]` to `{foo}`.
- Added `CONTAINS`.
- Added float comparison with epsilon error torlerence.
- Optimized long array `IN`/`CONTAINS` operator.
- Removed redundant RWMutex for better performance.
