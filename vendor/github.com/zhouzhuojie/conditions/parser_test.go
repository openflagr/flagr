package conditions

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var invalidTestData = []string{
	"",
	"A",
	"{var0} == DEMO",
	"{var0} == CA",
	"{var0} == PA",
	"{var0} == 'DEMO'",
	"!{var0}",
	"{var0} <> `DEMO`",
	"{foo} in [foobar]",
	"{foo} in [foobar, baz]",
	"{foo} in [\"foobar\", baz]",
	"{foo} in {foobar",
	"{foo} in [foobar",
	"{foo} in ['foobar']",
	"{foo} in ['foobar'",
	"{foo} in [1, 2, \"3\"]",
	"{foo} in [\"3\", 2, 1]",
	"{foo} in [\"3\", 2, 1",
	"{foo} not in [foobar]",
}

func TestInvalid(t *testing.T) {

	var (
		expr Expr
		err  error
	)

	for _, cond := range invalidTestData {
		t.Log("--------")
		t.Logf("Parsing: %s", cond)

		p := NewParser(strings.NewReader(cond))
		expr, err = p.Parse()
		if err == nil {
			t.Error("Should receive error")
			break
		}
		if expr != nil {
			t.Error("Expression should nil")
			break
		}
	}
}

var validTestData = []struct {
	cond   string
	args   map[string]interface{}
	result bool
	isErr  bool
}{
	{"true", nil, true, false},
	{"false", nil, false, false},
	{"false OR true OR false OR false OR true", nil, true, false},
	{"((false OR true) AND false) OR (false OR true)", nil, true, false},
	{"{var0}", map[string]interface{}{"var0": true}, true, false},
	{"{var0}", map[string]interface{}{"var0": false}, false, false},
	{"{var0} > true", nil, false, true},
	{"{var0} > true", map[string]interface{}{"var0": 43}, false, true},
	{"{var0} > true", map[string]interface{}{"var0": false}, false, true},
	{"{var0} and {var1}", map[string]interface{}{"var0": true, "var1": true}, true, false},
	{"{var0} AND {var1}", map[string]interface{}{"var0": true, "var1": false}, false, false},
	{"{var0} AND {var1}", map[string]interface{}{"var0": false, "var1": true}, false, false},
	{"{var0} AND {var1}", map[string]interface{}{"var0": false, "var1": false}, false, false},
	{"{var0} AND false", map[string]interface{}{"var0": true}, false, false},
	{"56.43", nil, false, true},
	{"{var5}", nil, false, true},
	{"{var0} > -100 AND {var0} < -50", map[string]interface{}{"var0": -75.4}, true, false},
	{"{var0}", map[string]interface{}{"var0": true}, true, false},
	{"{var0}", map[string]interface{}{"var0": false}, false, false},
	{"\"OFF\"", nil, false, true},
	{"`ON`", nil, false, true},
	{"{var0} == \"OFF\"", map[string]interface{}{"var0": "OFF"}, true, false},

	// AND
	{"{var0} > 10 AND {var1} == \"OFF\"", map[string]interface{}{"var0": 14, "var1": "OFF"}, true, false},
	{"({var0} > 10) AND ({var1} == \"OFF\")", map[string]interface{}{"var0": 14, "var1": "OFF"}, true, false},
	{"({var0} > 10) AND ({var1} == \"OFF\") OR true", map[string]interface{}{"var0": 1, "var1": "ON"}, true, false},
	{"{foo}{dfs} == true and {bar} == true", map[string]interface{}{"foo.dfs": true, "bar": true}, true, false},
	{"{foo}{dfs}{a} == true and {bar} == true", map[string]interface{}{"foo.dfs.a": true, "bar": true}, true, false},
	{"{@foo}{a} == true and {bar} == true", map[string]interface{}{"@foo.a": true, "bar": true}, true, false},
	{"{foo}{unknow} == true and {bar} == true", map[string]interface{}{"foo.dfs": true, "bar": true}, false, true},
	{"{foo} == 123", map[string]interface{}{"foo": json.Number("123"), "bar": true}, true, false},

	// OR
	{"{foo} == true OR {foo} > 1", map[string]interface{}{"foo": true}, false, true},
	{"{foo} == true OR {foo} == false", map[string]interface{}{"foo": true}, true, false},
	{"{foo} > 100 OR {foo} < 99 ", map[string]interface{}{"foo": 100}, false, false},
	{"{foo}{dfs} == true or {bar} == true", map[string]interface{}{"foo.dfs": true, "bar": true}, true, false},

	//XOR
	{"false XOR false", nil, false, false},
	{"false xor true", nil, true, false},
	{"true XOR false", nil, true, false},
	{"true xor true", nil, false, false},

	//NAND
	{"false NAND false", nil, true, false},
	{"false nand true", nil, true, false},
	{"true nand false", nil, true, false},
	{"true NAND true", nil, false, false},

	// IN
	{"{foo} in {foobar}", map[string]interface{}{"foo": "findme", "foobar": []string{"notme", "may", "findme", "lol"}}, true, false},
	{"{foo} in [123]", map[string]interface{}{"foo": json.Number("123"), "baz": true}, true, false},
	{"{foo} in [123]", map[string]interface{}{"foo": json.Number("124"), "baz": true}, false, false},

	// NOT IN
	{"{foo} not in {foobar}", map[string]interface{}{"foo": "dontfindme", "foobar": []string{"notme", "may", "findme", "lol"}}, true, false},

	// IN with array of string
	{`{foo} in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "le monde"}, true, false},
	{`{foo} in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "world"}, false, false},

	// NOT IN with array of string
	{`{foo} not in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "le monde"}, false, false},
	{`{foo} not in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "world"}, true, false},

	// IN with array of numbers
	{`{foo} in [2,3,4]`, map[string]interface{}{"foo": 4}, true, false},
	{`{foo} in [2,3,4] AND {foo} == 4`, map[string]interface{}{"foo": 4}, true, false},
	{`{foo} in [2,3,4] AND {foo} == 3`, map[string]interface{}{"foo": 4}, false, false},
	{`{foo} in [2,3,4]`, map[string]interface{}{"foo": 5}, false, false},

	//{NOT}IN with array of numbers
	{`{foo} not in [2,3,4]`, map[string]interface{}{"foo": 4}, false, false},
	{`{foo} not in [2,3,4]`, map[string]interface{}{"foo": 5}, true, false},

	//{CON}AINS
	{`{foo} contains "2"`, map[string]interface{}{"foo": []string{"1", "2"}}, true, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []string{"1", "2"}}, false, true},
	{`{foo} contains "2" and {foo} contains "1"`, map[string]interface{}{"foo": []string{"1", "2"}}, true, false},
	{`{foo} contains "2" and {foo} contains "0"`, map[string]interface{}{"foo": []string{"1", "2"}}, false, false},
	{`{foo} contains "2" or {foo} contains "0"`, map[string]interface{}{"foo": []string{"1", "2"}}, true, false},
	{`{foo} contains 2 and {foo} contains 1`, map[string]interface{}{"foo": []int{1, 2}}, true, false},
	{`{foo} contains 2 and {foo} contains 1`, map[string]interface{}{"foo": []int{1, 2}}, true, false},
	{`{foo} contains "2" and {foo} contains 1`, map[string]interface{}{"foo": []int{1, 2}}, false, true},
	{`{foo} contains {bar}`, map[string]interface{}{"foo": []string{"1", "2"}, "bar": "1"}, true, false},
	{`{foo} contains {bar}`, map[string]interface{}{"foo": []int{1, 2}, "bar": int32(1)}, true, false},
	{`{foo} contains {bar}`, map[string]interface{}{"foo": []int{1, 2, 3}, "bar": float32(1.0 + 2.0)}, true, false},
	{`{foo} contains {bar}`, map[string]interface{}{"foo": []float64{0.29}, "bar": float32(29.0 / 100)}, true, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []json.Number{"2"}}, true, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []json.Number{"2", "3"}}, true, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []json.Number{"3"}}, false, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []interface{}{json.Number("2")}}, true, false},
	{`{foo} contains 2`, map[string]interface{}{"foo": []interface{}{json.Number("3")}}, false, false},

	//{NOT}CONTAINS
	{`{foo} not contains "2"`, map[string]interface{}{"foo": []string{"1", "2"}}, false, false},
	{`{foo} not contains "0"`, map[string]interface{}{"foo": []string{"1", "2"}}, true, false},
	{`{foo} not contains 0`, map[string]interface{}{"foo": []string{"1", "2"}}, false, true},
	{`{foo} not contains 0`, map[string]interface{}{"bar": []string{"1", "2"}}, false, true},

	//{=~
	{"{status} =~ /^5\\d\\d/", map[string]interface{}{"status": "500"}, true, false},
	{"{status} =~ /^4\\d\\d/", map[string]interface{}{"status": "500"}, false, false},

	//{!~
	{"{status} !~ /^5\\d\\d/", map[string]interface{}{"status": "500"}, false, false},
	{"{status} !~ /^4\\d\\d/", map[string]interface{}{"status": "500"}, true, false},
}

func TestValid(t *testing.T) {

	var (
		expr Expr
		err  error
		r    bool
	)

	for _, td := range validTestData {
		t.Log("--------")
		t.Logf("Parsing: %s", td.cond)

		p := NewParser(strings.NewReader(td.cond))
		expr, err = p.Parse()
		t.Logf("Expression: %s", expr)
		if err != nil {
			t.Errorf("Unexpected error parsing expression: %s", td.cond)
			t.Error(err.Error())
			break
		}

		t.Logf("Evaluating with: %#v", td.args)
		r, err = Evaluate(expr, td.args)
		if err != nil {
			if td.isErr {
				continue
			}
			t.Errorf("Unexpected error evaluating: %s", expr)
			t.Error(err.Error())
			break
		} else {
			if td.isErr {
				t.Errorf("Expected no error evaluating: %s", expr)
				t.Error(err.Error())
				break
			}
		}
		if r != td.result {
			t.Errorf("Expected %v, received: %v", td.result, r)
			break
		}
	}
}

func TestExpressionsVariableNames(t *testing.T) {
	cond := "{@foo}{a} == true and {bar} == true or {var9} > 10"
	p := NewParser(strings.NewReader(cond))
	expr, err := p.Parse()
	assert.Nil(t, err)

	args := Variables(expr)
	assert.Contains(t, args, "@foo.a", "...")
	assert.Contains(t, args, "bar", "...")
	assert.Contains(t, args, "var9", "...")
	assert.NotContains(t, args, "foo", "...")
	assert.NotContains(t, args, "@foo", "...")
}

func TestFloat64Equal(t *testing.T) {
	epsilon := 1e-6
	assert.True(t, float64Equal(0.01, 0.01, epsilon))
	assert.True(t, float64Equal(0.01, 0.01000001, epsilon))
	assert.False(t, float64Equal(0.01, 0.0100001, epsilon))
	assert.False(t, float64Equal(0.0, 0.0000001, epsilon))
	assert.False(t, float64Equal(0, 0.0000000000000000001, epsilon))
}

func TestSetDefaultEpsilon(t *testing.T) {
	defer SetDefaultEpsilon(1e-6)

	t.Run("0.1 == 0.1", func(t *testing.T) {
		SetDefaultEpsilon(1e-6)
		p := NewParser(strings.NewReader("{foo} == 0.1"))
		expr, _ := p.Parse()
		r, err := Evaluate(expr, map[string]interface{}{"foo": 0.1})
		assert.True(t, r)
		assert.NoError(t, err)
	})

	t.Run("0.1 == 0.100000000001", func(t *testing.T) {
		SetDefaultEpsilon(1e-6)
		p := NewParser(strings.NewReader("{foo} == 0.1"))
		expr, _ := p.Parse()
		r, err := Evaluate(expr, map[string]interface{}{"foo": 0.100000000001})
		assert.True(t, r)
		assert.NoError(t, err)
	})

	t.Run("0.1 != 0.100001", func(t *testing.T) {
		SetDefaultEpsilon(1e-6)
		p := NewParser(strings.NewReader("{foo} == 0.1"))
		expr, _ := p.Parse()
		r, err := Evaluate(expr, map[string]interface{}{"foo": 0.100001})
		assert.False(t, r)
		assert.NoError(t, err)
	})

	t.Run("0.1 == 0.100001 if set epsilon to 1e-5", func(t *testing.T) {
		SetDefaultEpsilon(1e-5)
		p := NewParser(strings.NewReader("{foo} == 0.1"))
		expr, _ := p.Parse()
		r, err := Evaluate(expr, map[string]interface{}{"foo": 0.100001})
		assert.True(t, r)
		assert.NoError(t, err)
	})
}

func TestReadmeExample(t *testing.T) {
	// Our condition to check
	s := `({foo} > 0.45) AND ({bar} == "ON" OR {baz} IN ["ACTIVE", "CLEAR"])`

	// Parse the condition language and get expression
	p := NewParser(strings.NewReader(s))
	expr, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	data := map[string]interface{}{"foo": 0.62, "bar": "ON", "baz": "ACTIVE"}
	r, err := Evaluate(expr, data)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Evaluation result:", r)
}

func TestJSON(t *testing.T) {
	var tests = []struct {
		cond    string
		jsonStr string
		result  bool
		isErr   bool
	}{
		{`{foo} == 123`, `{"foo": 123}`, true, false},
		{`{foo} in [123]`, `{"foo": 123}`, true, false},
		{`{foo} in [124]`, `{"foo": 123}`, false, false},
		{`{foo} in [123]`, `{"foo": 123, "bar": "baz"}`, true, false},
		{`{foo} in [124]`, `{"foo": 123, "bar": "baz"}`, false, false},

		{`{foo} == "123"`, `{"foo": 123}`, false, true},
		{`{foo} == "123"`, `{"foo": "123"}`, true, false},
		{`{foo} not in ["123"]`, `{"foo": "123"}`, false, false},

		{`{foo} contains "123"`, `{"foo": ["123"]}`, true, false},
		{`{foo} contains 123`, `{"foo": [123]}`, true, false},
		{`{foo} contains 123`, `{"foo": ["123"]}`, false, true},
		{`{foo} not contains 123`, `{"foo": [124]}`, true, false},
		{`{foo} not contains "123"`, `{"foo": ["124"]}`, true, false},
	}

	for _, test := range tests {
		p := NewParser(strings.NewReader(test.cond))
		expr, _ := p.Parse()
		data := make(map[string]interface{})
		json.Unmarshal([]byte(test.jsonStr), &data)
		r, err := Evaluate(expr, data)
		assert.Equal(t, r, test.result, test.cond, test.jsonStr)
		if test.isErr {
			assert.Error(t, err, test.cond, test.jsonStr)
		} else {
			assert.NoError(t, err, test.cond)
		}
	}
}

func BenchmarkParser(b *testing.B) {
	cond := "({foo}{dfs}{a} == true AND {bar} == true) AND false"
	args := map[string]interface{}{"foo.dfs.a": true, "bar": true, "something": 1.0}
	p := NewParser(strings.NewReader(cond))
	expr, _ := p.Parse()

	for n := 0; n < b.N; n++ {
		r, _ := Evaluate(expr, args)
		fmt.Println(r)
	}
}
