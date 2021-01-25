package handler

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/Knetic/govaluate"
	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/fatih/structs"
	"github.com/go-openapi/runtime/middleware"
	"github.com/matryer/resync"
)

var (
	evaluationRules           []*govaluate.EvaluableExpression
	createEvaluationRule      resync.Once
	evaluationRulesSetupError error
)

/* with thanks to https://github.com/casbin/casbin/blob/d0d65d828c784211a47c23f9cc63b7429da0287e/util/builtin_operators.go */
// validate the variadic parameter size and type as string
func validateVariadicArgs(expectedLen int, args ...interface{}) error {
	if len(args) != expectedLen {
		return fmt.Errorf("Expected %d arguments, but got %d", expectedLen, len(args))
	}

	for _, p := range args {
		_, ok := p.(string)
		if !ok {
			return errors.New("Argument must be a string")
		}
	}

	return nil
}

// RegexMatch determines whether key1 matches the pattern of key2 in regular expression.
func RegexMatch(key1 string, key2 string) (bool, error) {
	res, err := regexp.MatchString(key2, key1)
	if err != nil {
		return false, err
	}
	return res, nil
}

// RegexMatchFunc is a govaluate function for matching regex,
// it returns true if arg1 matches the regex pattern of arg2.
func RegexMatchFunc(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "regexMatch", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	result, err := RegexMatch(name1, name2)
	if err != nil {
		return false, err
	}

	return result, nil
}

/* end 'with thanks to' */

// THIS got weird when the params were in the other order (e.g. array, regex)
// in this case it flattened the array into the args, e.g.
// Any([1,2], /\d+/)
// args = 1,2,/\d+/
// putting it in regex, array order does not flatten it
// args = /\d+/, [1,2]
//
// Govaluate function to evaluate an embedded govaluate expression (the first argument, a string)
// for an array of 'items' (which should be a map[string]interface{}),
// and if the subexpression returns true for any of the sub items
// then the Any expression also returns true. It will early out once
// it finds a single true result.
func Any(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("Expected 2+ arguments, but got %d", len(args))
	}

	function, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("arg 0 is not a string!")
	}

	arr, ok := args[1].([]interface{})
	if !ok {
		return false, fmt.Errorf("arg 1 is not a slice! ***%v*** %v", args, reflect.TypeOf(args[1]))
	}

	functions := map[string]govaluate.ExpressionFunction{
		"regexMatch": RegexMatchFunc,
	}
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(function, functions)
	if err != nil {
		return false, err
	}

	found := false
	for _, item := range arr {
		params, ok := item.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("item is not a map %v", item)
		}

		result, err := expression.Evaluate(params)
		if err != nil {
			return false, err
		}

		if boolResult, ok := result.(bool); ok {
			if boolResult {
				found = true
				break
			}
		}
	}
	return found, nil
}

func getEvaluationRules() ([]*govaluate.EvaluableExpression, error) {
	createEvaluationRule.Do(func() {
		for _, structExpressionString := range config.Config.EnabledFlagValidationRules {
			functions := map[string]govaluate.ExpressionFunction{
				"regexMatch": RegexMatchFunc,
				"any":        Any,
			}

			expression, err := govaluate.NewEvaluableExpressionWithFunctions(structExpressionString, functions)
			if expression == nil || err != nil {
				evaluationRulesSetupError = err
				return
			}
			evaluationRules = append(evaluationRules, expression)
		}
	})

	return evaluationRules, evaluationRulesSetupError
}

func validateEnabledFlag(f *entity.Flag) (valid bool, err error) {
	if config.Config.EnabledFlagValidationOperation == config.FlagValidationOperationAND {
		valid = true
	}

	evaluationRules, err := getEvaluationRules()
	if err != nil {
		return false, err
	}

	for _, validationRule := range evaluationRules {
		params := structs.Map(f)

		result, err := validationRule.Evaluate(params)
		if err != nil {
			return false, err
		}

		resultBool, ok := result.(bool)
		if !ok {
			return false, fmt.Errorf("Unknown evaluation rule result type (%v)", result)
		}

		if config.Config.EnabledFlagValidationOperation == config.FlagValidationOperationOR {
			valid = valid || resultBool
			if valid {
				break
			}
		} else if config.Config.EnabledFlagValidationOperation == config.FlagValidationOperationAND {
			valid = valid && resultBool
		} else {
			return false, errors.New("Unknown FlagEnabledFlagValidationOperation - check server config")
		}
	}

	return valid, nil
}

func validateFlagIfEnabled(f *entity.Flag) middleware.Responder {
	if f.Enabled {
		valid, err := validateEnabledFlag(f)
		if err != nil {
			return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
		}
		if !valid {
			opMsg := "all"
			if config.Config.EnabledFlagValidationOperation == config.FlagValidationOperationOR {
				opMsg = "one"
			}

			return flag.NewPutFlagDefault(400).WithPayload(ErrorMessage("Flag failed %s of these validation rules: %s", opMsg, config.Config.EnabledFlagValidationRules))
		}
	}
	return nil
}
