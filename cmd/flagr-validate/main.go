// flagr-validate validates a Flagr JSON flag definition file.
//
// Usage:
//
//	flagr-validate <flags.json>
//	flagr-validate --help
//
// Exit codes:
//
//	0 — valid (may have warnings)
//	1 — errors found
//	2 — usage error
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/openflagr/flagr/pkg/handler"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags.json>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nValidates a Flagr JSON flag definition file.\n")
		fmt.Fprintf(os.Stderr, "Checks: valid JSON, required fields, key uniqueness,\n")
		fmt.Fprintf(os.Stderr, "distribution sums, variant references.\n")
		os.Exit(2)
	}

	path := os.Args[1]
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	var ecj handler.EvalCacheJSON
	if err := json.Unmarshal(b, &ecj); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid JSON: %v\n", err)
		os.Exit(1)
	}

	result := handler.ValidateFlags(ecj.Flags)

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
	}
	for _, e := range result.Errors {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", e)
	}

	if !result.OK() {
		fmt.Fprintf(os.Stderr, "%s: %d error(s)\n", path, len(result.Errors))
		os.Exit(1)
	}

	if result.HasWarnings() {
		fmt.Fprintf(os.Stderr, "%s: valid with %d warning(s)\n", path, len(result.Warnings))
	} else {
		fmt.Fprintf(os.Stderr, "%s: valid\n", path)
	}
}
