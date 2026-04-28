package route

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
)

func ApplyVariables(input any, variables map[string]string) error {
	slog.InfoContext(context.TODO(), "ApplyVariables", "input", input)

	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("input is not a struct")
	}

	for field, fieldValue := range value.Fields() {
		tag := field.Tag.Get("route")
		if tag == "" {
			continue
		}
		slog.InfoContext(context.TODO(), "ApplyVariables", "tag", tag)

		variableName := tag
		variableValue := variables[variableName]

		if !fieldValue.CanSet() {
			slog.WarnContext(context.TODO(), "Field cannot be set.", "field", field.Name)
			continue
		}
		if fieldValue.Kind() != reflect.String {
			slog.WarnContext(context.TODO(), "Field is not a string.", "field", field.Name)
			continue
		}
		slog.InfoContext(context.TODO(), "Setting field value", "field", field.Name, "value", variableValue)
		fieldValue.SetString(variableValue)
	}

	return nil
}
