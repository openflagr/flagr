package internal

import (
	"encoding/json"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestMetricRules(t *testing.T) {
	var tcs []struct {
		Testname string      `json:"testname"`
		Rules    metricRules `json:"rules"`
		Tests    []struct {
			Input    string `json:"input"`
			Expected string `json:"expected"`
		} `json:"tests"`
	}

	err := crossagent.ReadJSON("rules.json", &tcs)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range tcs {
		// This test relies upon Perl-specific regex syntax (negative
		// lookahead assertions) which are not implemented in Go's
		// regexp package. We believe these types of rules are
		// exceedingly rare in practice, so we're skipping
		// implementation of this exotic syntax for now.
		if tc.Testname == "saxon's test" {
			continue
		}

		for _, x := range tc.Tests {
			out := tc.Rules.Apply(x.Input)
			if out != x.Expected {
				t.Fatal(tc.Testname, x.Input, out, x.Expected)
			}
		}
	}
}

func TestMetricRuleWithNegativeLookaheadAssertion(t *testing.T) {
	js := `[{
		"match_expression":"^(?!account|application).*",
		"replacement":"*",
		"ignore":false,
		"eval_order":0,
		"each_segment":true
	}]`
	var rules metricRules
	err := json.Unmarshal([]byte(js), &rules)
	if nil != err {
		t.Fatal(err)
	}
	if 0 != rules.Len() {
		t.Fatal(rules)
	}
}

func TestNilApplyRules(t *testing.T) {
	var rules metricRules

	input := "hello"
	out := rules.Apply(input)
	if input != out {
		t.Fatal(input, out)
	}
}

func TestAmbiguousReplacement(t *testing.T) {
	js := `[{
		"match_expression":"(.*)/[^/]*.(bmp|css|gif|ico|jpg|jpeg|js|png)",
		"replacement":"\\\\1/*.\\2",
		"ignore":false,
		"eval_order":0
	}]`
	var rules metricRules
	err := json.Unmarshal([]byte(js), &rules)
	if nil != err {
		t.Fatal(err)
	}
	if 0 != rules.Len() {
		t.Fatal(rules)
	}
}

func TestBadMetricRulesJSON(t *testing.T) {
	js := `{}`
	var rules metricRules
	err := json.Unmarshal([]byte(js), &rules)
	if nil == err {
		t.Fatal("missing bad json error")
	}
}
