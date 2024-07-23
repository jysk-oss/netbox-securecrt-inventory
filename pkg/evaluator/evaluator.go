package evaluator

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

var compiledConditions map[string]*vm.Program = make(map[string]*vm.Program)

func EvaluateCondition(condition string, env *Environment) (bool, error) {
	output, err := EvaluateResult(condition, env)
	if err != nil {
		return false, err
	}

	val, ok := output.(bool)
	if !ok {
		slog.Error("Condition should return true or false", slog.Any("result", val))
		return false, errors.New("condition should return true or false")
	}

	return val, nil
}

func EvaluateResult(condition string, env *Environment) (any, error) {
	if !strings.HasPrefix(condition, "{{") {
		return ApplyTemplate(condition, env), nil
	}

	condition = strings.ReplaceAll(condition, "{{", "")
	condition = strings.ReplaceAll(condition, "}}", "")
	condition = strings.Trim(condition, " ")

	// compile and cache conditions
	if compiledConditions[condition] == nil {
		program, err := expr.Compile(condition)
		if err != nil {
			slog.Error("Failed to compile condition", slog.String("error", err.Error()))
			return false, err
		}

		compiledConditions[condition] = program
	}

	output, err := expr.Run(compiledConditions[condition], env)
	if err != nil {
		slog.Error("Failed to run condition", slog.String("error", err.Error()))
		return false, err
	}

	slog.Debug("Evaluation Result", slog.String("device", env.DeviceName), slog.String("condition", condition), slog.Any("result", output))
	return output, nil
}

func ApplyTemplate(template string, env *Environment) string {
	oldTemplate := template
	v := reflect.ValueOf(env).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.String {
			tag := v.Type().Field(i).Tag.Get("expr")
			template = strings.ReplaceAll(template, fmt.Sprintf("{%s}", tag), v.Field(i).String())
		}
	}

	slog.Debug("Evaluation Template Result", slog.String("device", env.DeviceName), slog.String("template", oldTemplate), slog.String("result", template))
	return template
}
