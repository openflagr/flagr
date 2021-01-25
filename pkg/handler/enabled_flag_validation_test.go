package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestRegexMatchFuncNotStringArg1(t *testing.T) {
	_, err := RegexMatchFunc(1, "foo")
	assert.NotNil(t, err)
}

func TestRegexMatchFuncNotStringArg2(t *testing.T) {
	_, err := RegexMatchFunc("foo", 1)
	assert.NotNil(t, err)
}

func TestRegexMatchFuncMatching(t *testing.T) {
	result, err := RegexMatchFunc("Bob", "\\w+")
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestRegexMatchFuncNotMatching(t *testing.T) {
	result, err := RegexMatchFunc("Bob", "\\d+")
	assert.Nil(t, err)
	assert.Equal(t, false, result)
}

func TestRegexMatchTooManyArgs(t *testing.T) {
	_, err := RegexMatchFunc("Bob", "Frank", "Jim")
	assert.NotNil(t, err)
}

func TestRegexMatchTooFewArgs(t *testing.T) {
	_, err := RegexMatchFunc("Bob")
	assert.NotNil(t, err)
}

func TestAnyTooFewArgs(t *testing.T) {
	_, err := Any(1)
	assert.NotNil(t, err)
}

func TestAnyArg0NotAString(t *testing.T) {
	_, err := Any(1, 1)
	assert.NotNil(t, err)
}

func TestAnyArg1NotASlice(t *testing.T) {
	_, err := Any("foo", 1)
	assert.NotNil(t, err)
}

func TestAnyMatch(t *testing.T) {
	regexp := "regexMatch(Value, \"^JIRA:[a-zA-Z]{3}[a-zA-Z]*$\")"
	slice := make([]interface{}, 0)
	slice = append(slice, map[string]interface{}{"Value": "FOOBAR"})
	slice = append(slice, map[string]interface{}{"Value": "JIRA:EPLT"})
	result, err := Any(regexp, slice)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestAnyNoMatch(t *testing.T) {
	regexp := "regexMatch(Value, \"^JIRA:[a-zA-Z]{3}[a-zA-Z]*$\")"
	slice := make([]interface{}, 0)
	slice = append(slice, map[string]interface{}{"Value": "FOOBAR"})
	slice = append(slice, map[string]interface{}{"Value": "QUIX"})
	result, err := Any(regexp, slice)
	assert.Nil(t, err)
	assert.Equal(t, false, result)
}

func setFlagValidationConfig(rules []string, operation string) (reset func()) {
	old := config.Config
	config.Config.EnabledFlagValidationRules = rules
	config.Config.EnabledFlagValidationOperation = operation
	createEvaluationRule.Reset()

	return func() {
		config.Config = old
	}
}
func TestValidationEnabledFlag(t *testing.T) {
	rules := []string{
		`any("regexMatch(Value, \"^JIRA:[a-zA-Z]{3}[a-zA-Z]*$\")", Tags) `,
	}
	resetter := setFlagValidationConfig(rules, config.FlagValidationOperationAND)
	defer resetter()

	t.Run("valid tag will pass", func(t *testing.T) {
		flag := &entity.Flag{
			Tags: []entity.Tag{
				{Value: "QUX"},
				{Value: "JIRA:EPLT"},
			},
		}
		result, err := validateEnabledFlag(flag)
		assert.Nil(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("invalid tag will fail", func(t *testing.T) {
		flag := &entity.Flag{
			Tags: []entity.Tag{
				{Value: "QUX"},
			},
		}
		result, err := validateEnabledFlag(flag)
		assert.Nil(t, err)
		assert.Equal(t, false, result)
	})
}
