package handler

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/openflagr/flagr/pkg/entity"
)

// ValidationResult holds the outcome of validating a flag definition.
type ValidationResult struct {
	Errors   []string
	Warnings []string
}

// OK returns true if there are no errors.
func (r ValidationResult) OK() bool { return len(r.Errors) == 0 }

// HasWarnings returns true if there are warnings.
func (r ValidationResult) HasWarnings() bool { return len(r.Warnings) > 0 }

// ValidateFlags validates a set of entity.Flag structs.
// It performs semantic validation: required fields, key uniqueness,
// constraint expressions, distribution integrity, variant references,
// and percentage ranges.
func ValidateFlags(flags []entity.Flag) ValidationResult {
	var r ValidationResult

	flagKeys := make([]string, 0, len(flags))
	for i := range flags {
		validateFlag(&r, flags[i], i)
		// Only track non-empty keys to avoid spurious duplicate reports
		// when multiple flags have empty keys (already reported as errors
		// by validateFlag).
		if flags[i].Key != "" {
			flagKeys = append(flagKeys, flags[i].Key)
		}
	}

	if dupes := duplicates(flagKeys); len(dupes) > 0 {
		for _, d := range dupes {
			r.Errors = append(r.Errors, fmt.Sprintf("duplicate flag key %q", d))
		}
	}

	return r
}

func validateFlag(r *ValidationResult, f entity.Flag, idx int) {
	if f.Key == "" {
		r.Errors = append(r.Errors, fmt.Sprintf("flag[%d]: missing or empty Key", idx))
		return
	}
	prefix := fmt.Sprintf("flag %q", f.Key)

	if len(f.Variants) == 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("%s: no variants defined", prefix))
	}
	variantKeys := make([]string, 0, len(f.Variants))
	variantKeySet := make(map[string]bool, len(f.Variants))
	for j, v := range f.Variants {
		if v.Key == "" {
			r.Errors = append(r.Errors, fmt.Sprintf("%s, variant[%d]: missing or empty Key", prefix, j))
			continue
		}
		variantKeys = append(variantKeys, v.Key)
		variantKeySet[v.Key] = true

		if len(v.Attachment) > 0 {
			if _, err := json.Marshal(v.Attachment); err != nil {
				r.Errors = append(r.Errors, fmt.Sprintf("%s, variant %q: invalid Attachment JSON: %v", prefix, v.Key, err))
			}
		}
	}
	if dupes := duplicates(variantKeys); len(dupes) > 0 {
		for _, d := range dupes {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: duplicate variant key %q", prefix, d))
		}
	}

	if len(f.Segments) == 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("%s: no segments defined", prefix))
	}
	for j, seg := range f.Segments {
		segDesc := seg.Description
		if segDesc == "" {
			segDesc = fmt.Sprintf("segment[%d]", j)
		}
		segPrefix := fmt.Sprintf("%s, %s", prefix, segDesc)

		if seg.RolloutPercent > 100 {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: RolloutPercent %d out of range (0-100)", segPrefix, seg.RolloutPercent))
		}
		validateDistributions(r, segPrefix, seg, variantKeySet)
		validateConstraints(r, segPrefix, seg)
	}
}

func validateDistributions(r *ValidationResult, prefix string, seg entity.Segment, variantKeySet map[string]bool) {
	if len(seg.Distributions) == 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("%s: no distributions defined", prefix))
		return
	}

	sum := uint(0)
	for _, d := range seg.Distributions {
		sum += d.Percent

		if d.Percent > 100 {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: distribution percent %d out of range (0-100)", prefix, d.Percent))
		}

		if d.VariantKey == "" && d.VariantID == 0 {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: distribution has no VariantKey or VariantID", prefix))
			continue
		}

		if d.VariantKey != "" && !variantKeySet[d.VariantKey] {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: distribution references unknown variant key %q", prefix, d.VariantKey))
		}
	}

	if sum != 100 {
		r.Errors = append(r.Errors, fmt.Sprintf("%s: distribution sum is %d (expected 100)", prefix, sum))
	}
}

func validateConstraints(r *ValidationResult, prefix string, seg entity.Segment) {
	for _, c := range seg.Constraints {
		entityConstraint := entity.Constraint{
			Property: c.Property,
			Operator: c.Operator,
			Value:    c.Value,
		}
		if err := entityConstraint.Validate(); err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: constraint %q %s %q is invalid: %v",
				prefix, c.Property, c.Operator, c.Value, err))
		}
	}
}

// duplicates returns the duplicate values in a string slice, sorted.
func duplicates(ss []string) []string {
	seen := make(map[string]int, len(ss))
	for _, s := range ss {
		seen[s]++
	}
	var dupes []string
	for s, count := range seen {
		if count > 1 {
			dupes = append(dupes, s)
		}
	}
	sort.Strings(dupes)
	return dupes
}
