//go:build integration

package flagr_integration

import (
	"fmt"
)

// Flag spec definitions for seeding.
// Each flag is created via HTTP API: flag → segment → constraints → variants → distributions → tags.

type constraintDef struct {
	Property string
	Operator string
	Value    string
}

type flagDef struct {
	Key         string
	Description string
	EntityType  string
	Enabled     bool
	Constraints []constraintDef
	Tags        []string
}

// initFlagDefs returns all flag definitions to seed.
// Called explicitly from seedFlags; no init() global side effects.
func initFlagDefs() []flagDef {
	// Flat list of (operator, property, value) entries — one per flag.
	type entry struct {
		op    string
		prop  string
		value string
	}
	entries := []entry{
		// EQ
		{"EQ", "region", `"us-west"`},
		{"EQ", "tier", `"premium"`},
		{"EQ", "status", `"active"`},
		{"EQ", "color", `"blue"`},
		// NEQ
		{"NEQ", "region", `"us-east"`},
		{"NEQ", "env", `"prod"`},
		{"NEQ", "status", `"banned"`},
		{"NEQ", "plan", `"free"`},
		// LT
		{"LT", "age", `18`},
		{"LT", "score", `100`},
		{"LT", "level", `5`},
		{"LT", "attempts", `3`},
		// LTE
		{"LTE", "age", `65`},
		{"LTE", "rating", `4.5`},
		{"LTE", "max_retries", `10`},
		{"LTE", "version", `2`},
		// GT
		{"GT", "age", `21`},
		{"GT", "revenue", `1000`},
		{"GT", "count", `100`},
		{"GT", "priority", `3`},
		// GTE
		{"GTE", "age", `18`},
		{"GTE", "score", `80`},
		{"GTE", "years_exp", `2`},
		{"GTE", "tier_num", `5`},
		// EREG
		{"EREG", "email", `".+@company\\.com"`},
		{"EREG", "phone", `"^\\+1[0-9]{10}$"`},
		{"EREG", "zip", `"^[0-9]{5}$"`},
		{"EREG", "user_agent", `".*Mobile.*"`},
		// NEREG
		{"NEREG", "email", `".*@spam\\.com"`},
		{"NEREG", "domain", `"^internal\\."`},
		{"NEREG", "path", `"^/admin"`},
		{"NEREG", "input", `"bad-word"`},
		// IN
		{"IN", "region", `["us-west","us-east"]`},
		{"IN", "role", `["admin","editor"]`},
		{"IN", "state", `["CA","NY","TX"]`},
		{"IN", "category", `["a","b","c"]`},
		// NOTIN
		{"NOTIN", "blacklist", `["10.0.0.0/8","192.168.0.0/16"]`},
		{"NOTIN", "banned_words", `["evil","spam"]`},
		{"NOTIN", "excluded", `["internal"]`},
		{"NOTIN", "blocked", `["v1","v2"]`},
		// CONTAINS
		{"CONTAINS", "tags", `"premium"`},
		{"CONTAINS", "permissions", `"delete"`},
		{"CONTAINS", "features", `"beta"`},
		{"CONTAINS", "groups", `"engineering"`},
		// NOTCONTAINS
		{"NOTCONTAINS", "exclusions", `"banned"`},
		{"NOTCONTAINS", "blocklist", `"deprecated"`},
		{"NOTCONTAINS", "disabled", `"off"`},
		{"NOTCONTAINS", "muted", `"silent"`},
	}

	var defs []flagDef
	for i, e := range entries {
		defs = append(defs, flagDef{
			Key:         "int_flag_" + e.op + "_" + fmt.Sprintf("%02d", i+1),
			Description: "Integration test flag: " + e.op + " on " + e.prop,
			EntityType:  "user",
			Enabled:     true,
			Constraints: []constraintDef{
				{
					Property: e.prop,
					Operator: e.op,
					Value:    e.value,
				},
			},
			Tags: []string{"int_test", "constraint_" + e.op},
		})
	}

	// Ensure all defs have an EntityType — simplifies seedFlags condition.
	for i := range defs {
		if defs[i].EntityType == "" {
			defs[i].EntityType = "user"
		}
	}

	// Extra flags
	extras := []flagDef{
		{
			Key:         "int_flag_complex_and",
			Description: "Complex AND constraints: EQ + GT + IN",
			EntityType:  "user",
			Enabled:     true,
			Constraints: []constraintDef{
				{Property: "region", Operator: "EQ", Value: `"us-west"`},
				{Property: "age", Operator: "GT", Value: `18`},
				{Property: "tier", Operator: "IN", Value: `["premium","enterprise"]`},
			},
			Tags: []string{"int_test", "complex_and"},
		},
		{
			Key:         "int_flag_entity_type_override",
			Description: "Flag with custom entityType for propagation testing",
			EntityType:  "custom_entity",
			Enabled:     true,
			Constraints: []constraintDef{
				{Property: "region", Operator: "EQ", Value: `"us-west"`},
			},
			Tags: []string{"int_test", "entity_type"},
		},
		{
			Key:         "int_flag_disabled",
			Description: "Disabled flag — eval should return empty",
			EntityType:  "user",
			Enabled:     false,
			Constraints: []constraintDef{
				{Property: "region", Operator: "EQ", Value: `"us-west"`},
			},
			Tags: []string{"int_test", "disabled"},
		},
	}

	return append(defs, extras...)
}

// seedFlags creates all flag definitions via the HTTP API.
func seedFlags(errorf func(string, ...any)) {
	defs := initFlagDefs()
	for _, f := range defs {
		var flag flagResponse
		doReqAndDecode("POST", "/api/v1/flags", map[string]any{
			"description": f.Description,
			"key":         f.Key,
			"enabled":     f.Enabled,
		}, &flag, errorf)
		seedFlagIDs = append(seedFlagIDs, flag.ID)
		seedFlagKeys = append(seedFlagKeys, flag.Key)

		if f.EntityType != "user" {
			doReqAndDecode("PUT", fmt.Sprintf("/api/v1/flags/%d", flag.ID), map[string]any{
				"entityType": f.EntityType,
			}, nil, errorf)
		}

		if f.Enabled {
			doReqAndDecode("PUT", fmt.Sprintf("/api/v1/flags/%d/enabled", flag.ID), map[string]any{
				"enabled": true,
			}, nil, errorf)
		}

		var seg segmentResponse
		doReqAndDecode("POST", fmt.Sprintf("/api/v1/flags/%d/segments", flag.ID), map[string]any{
			"description":    "default segment",
			"rolloutPercent": 100,
		}, &seg, errorf)

		for _, c := range f.Constraints {
			doReqAndDecode("POST", fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flag.ID, seg.ID), map[string]any{
				"property": c.Property,
				"operator": c.Operator,
				"value":    c.Value,
			}, nil, errorf)
		}

		var v1, v2 variantResponse
		doReqAndDecode("POST", fmt.Sprintf("/api/v1/flags/%d/variants", flag.ID), map[string]any{
			"key": "variant_control",
		}, &v1, errorf)
		doReqAndDecode("POST", fmt.Sprintf("/api/v1/flags/%d/variants", flag.ID), map[string]any{
			"key": "variant_treatment",
		}, &v2, errorf)

		doReqAndDecode("PUT", fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flag.ID, seg.ID), map[string]any{
			"distributions": []map[string]any{
				{"percent": 100, "variantID": v1.ID, "variantKey": v1.Key},
				{"percent": 0, "variantID": v2.ID, "variantKey": v2.Key},
			},
		}, nil, errorf)

		for _, tag := range f.Tags {
			doReqAndDecode("POST", fmt.Sprintf("/api/v1/flags/%d/tags", flag.ID), map[string]any{
				"value": tag,
			}, nil, errorf)
		}
	}

	fmt.Printf("Seeded %d flags\n", len(seedFlagIDs))
}
