//go:build integration

package flagr_integration

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

// matchingEntity returns entity context that makes all constraints of this flag match.
func (f flagDef) matchingEntity() map[string]any {
	m := map[string]any{}
	_ = f // placeholder — used per-flag in eval tests
	return m
}

var allFlagDefs []flagDef

func init() {
	type opGroup struct {
		operator string
		cases    []struct {
			prop  string
			value string
		}
	}

	groups := []opGroup{
		{
			operator: "EQ",
			cases: []struct {
				prop  string
				value string
			}{
				{"region", `"us-west"`},
				{"tier", `"premium"`},
				{"status", `"active"`},
				{"color", `"blue"`},
			},
		},
		{
			operator: "NEQ",
			cases: []struct {
				prop  string
				value string
			}{
				{"region", `"us-east"`},
				{"env", `"prod"`},
				{"status", `"banned"`},
				{"plan", `"free"`},
			},
		},
		{
			operator: "LT",
			cases: []struct {
				prop  string
				value string
			}{
				{"age", `18`},
				{"score", `100`},
				{"level", `5`},
				{"attempts", `3`},
			},
		},
		{
			operator: "LTE",
			cases: []struct {
				prop  string
				value string
			}{
				{"age", `65`},
				{"rating", `4.5`},
				{"max_retries", `10`},
				{"version", `2`},
			},
		},
		{
			operator: "GT",
			cases: []struct {
				prop  string
				value string
			}{
				{"age", `21`},
				{"revenue", `1000`},
				{"count", `100`},
				{"priority", `3`},
			},
		},
		{
			operator: "GTE",
			cases: []struct {
				prop  string
				value string
			}{
				{"age", `18`},
				{"score", `80`},
				{"years_exp", `2`},
				{"tier_num", `5`},
			},
		},
		{
			operator: "EREG",
			cases: []struct {
				prop  string
				value string
			}{
				{"email", `".+@company\\.com"`},
				{"phone", `"^\\+1[0-9]{10}$"`},
				{"zip", `"^[0-9]{5}$"`},
				{"user_agent", `".*Mobile.*"`},
			},
		},
		{
			operator: "NEREG",
			cases: []struct {
				prop  string
				value string
			}{
				{"email", `".*@spam\\.com"`},
				{"domain", `"^internal\\."`},
				{"path", `"^/admin"`},
				{"input", `"bad-word"`},
			},
		},
		{
			operator: "IN",
			cases: []struct {
				prop  string
				value string
			}{
				{"region", `["us-west","us-east"]`},
				{"role", `["admin","editor"]`},
				{"state", `["CA","NY","TX"]`},
				{"category", `["a","b","c"]`},
			},
		},
		{
			operator: "NOTIN",
			cases: []struct {
				prop  string
				value string
			}{
				{"blacklist", `["10.0.0.0/8","192.168.0.0/16"]`},
				{"banned_words", `["evil","spam"]`},
				{"excluded", `["internal"]`},
				{"blocked", `["v1","v2"]`},
			},
		},
		{
			operator: "CONTAINS",
			cases: []struct {
				prop  string
				value string
			}{
				{"tags", `"premium"`},
				{"permissions", `"delete"`},
				{"features", `"beta"`},
				{"groups", `"engineering"`},
			},
		},
		{
			operator: "NOTCONTAINS",
			cases: []struct {
				prop  string
				value string
			}{
				{"exclusions", `"banned"`},
				{"blocklist", `"deprecated"`},
				{"disabled", `"off"`},
				{"muted", `"silent"`},
			},
		},
	}

	idx := 0
	for _, g := range groups {
		for _, c := range g.cases {
			idx++
			allFlagDefs = append(allFlagDefs, flagDef{
				Key:         "int_flag_" + g.operator + "_" + itoa(idx),
				Description: "Integration test flag: " + g.operator + " on " + c.prop,
				EntityType:  "user",
				Enabled:     true,
				Constraints: []constraintDef{
					{
						Property: c.prop,
						Operator: g.operator,
						Value:    c.value,
					},
				},
				Tags: []string{"int_test", "constraint_" + g.operator},
			})
		}
	}

	// Extra flags
	extras := []flagDef{
		{
			Key:         "int_flag_multi_segment",
			Description: "Multi-segment flag with 3 segments at different rollout %",
			EntityType:  "user",
			Enabled:     true,
			Constraints: []constraintDef{
				{Property: "region", Operator: "EQ", Value: `"us-west"`},
				{Property: "age", Operator: "GT", Value: `18`},
				{Property: "tier", Operator: "IN", Value: `["premium","enterprise"]`},
			},
			Tags: []string{"int_test", "multi_segment"},
		},
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

	allFlagDefs = append(allFlagDefs, extras...)
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0' + n/10)) + string(rune('0' + n%10))
}
