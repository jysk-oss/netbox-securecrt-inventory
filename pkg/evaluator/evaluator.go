package evaluator

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/expr-lang/expr"
)

func addFunctionsToEnv(env map[string]interface{}) {
	env["findTag"] = FindTag

}

func EvaluateCondition(condition string, env map[string]interface{}) (bool, error) {
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

func EvaluateResult(condition string, env map[string]interface{}) (any, error) {
	if !strings.HasPrefix(condition, "{{") {
		return ApplyTemplate(condition, env), nil
	}

	addFunctionsToEnv(env)
	condition = strings.ReplaceAll(condition, "{{", "")
	condition = strings.ReplaceAll(condition, "}}", "")
	condition = strings.Trim(condition, " ")
	program, err := expr.Compile(condition, expr.Env(env))
	if err != nil {
		slog.Error("Failed to compile condition", slog.String("error", err.Error()))
		return false, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		slog.Error("Failed to run condition", slog.String("error", err.Error()))
		return false, err
	}

	slog.Debug("Evaluation Result", slog.String("device", env["device_name"].(string)), slog.String("condition", condition), slog.Any("result", output))
	return output, nil
}

func ApplyTemplate(template string, env map[string]interface{}) string {
	oldTemplate := template
	for key, value := range env {
		// TODO: should we support more types here?
		val, ok := value.(string)
		if ok {
			template = strings.ReplaceAll(template, fmt.Sprintf("{%s}", key), val)
		}
	}

	slog.Debug("Evaluation Template Result", slog.String("device", env["device_name"].(string)), slog.String("template", oldTemplate), slog.String("result", template))
	return template
}
