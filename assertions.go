package testpilot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"strings"
)

type AssertionFunc func(body []byte) error

// AssertEqual asserts that the response body is equal to the given value
// It uses reflect.DeepEqual to compare the response body with the given value
func AssertEqual(response any) AssertionFunc {
	return func(body []byte) error {
		t := reflect.TypeOf(response)
		if t.Kind() == reflect.Ptr {
			t = t.Elem() // If it's a pointer, get the type it points to
		}
		newVar := reflect.New(t).Interface()
		var got any
		if t.Kind() == reflect.Struct || t.Kind() == reflect.Map || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			if err := json.Unmarshal(body, newVar); err != nil {
				return err
			}
			got = reflect.ValueOf(newVar).Elem().Interface()
		} else {
			got = strings.TrimSpace(string(body))
		}
		if !reflect.DeepEqual(response, got) {
			diff := cmp.Diff(response, got)
			coloredDiff := formatDiff(diff)
			return errors.New(fmt.Sprintf("response body does not match. \n "+
				"expected: %v \n "+
				"got     : %v \n\n Diff: \n %v", truncate(response), truncate(got), coloredDiff))
		}
		return nil
	}
}

// AssertPath asserts that the value at the given path in the response body satisfies the given assertion
// the path is a dot separated string representing the path to the value in the response body
// e.g. "data.user.0.name" will navigate to the first user in the data array and check if the name field satisfies the given assertion
func AssertPath[T comparable](path string, assert func(val T) error) AssertionFunc {
	return func(body []byte) error {
		var data any
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		if path[0] == '.' {
			path = strings.Replace(path, ".", "", 1)
		}
		value, err := navigateJSON(data, path)
		if err != nil {
			return err
		}
		v, err := convertToType[T](value)
		if err != nil {
			return err
		}
		if err := assert(v); err != nil {
			return err
		}
		return nil
	}
}

// Equal returns an assertion function that checks if the given value is equal to the expected value
func Equal[T comparable](expected T) func(val T) error {
	return func(val T) error {
		if val != expected {
			return fmt.Errorf("expected %v got %v", expected, val)
		}
		return nil
	}
}
func truncate(data interface{}) string {
	value := fmt.Sprintf("%#v", data)
	maxSize := bufio.MaxScanTokenSize - 100
	if len(value) > maxSize {
		value = value[0:maxSize] + "..."
	}
	return value
}
func formatDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var formattedDiff string
	for _, line := range lines {
		if strings.HasPrefix(line, "-") {
			formattedDiff += fmt.Sprintf("\033[31m%s\033[0m\n", line)
		} else if strings.HasPrefix(line, "+") {
			formattedDiff += fmt.Sprintf("\033[32m%s\033[0m\n", line)
		} else {
			formattedDiff += fmt.Sprintf("%s\n", line)
		}
	}
	return formattedDiff
}

func convertToType[T comparable](value interface{}) (T, error) {
	var result T
	targetType := reflect.TypeOf(result)
	valueReflect := reflect.ValueOf(value)
	if valueReflect.Type().ConvertibleTo(targetType) {
		convertedValue := valueReflect.Convert(targetType)
		result = convertedValue.Interface().(T)
		return result, nil
	}
	if targetType.Kind() == reflect.Struct && valueReflect.Kind() == reflect.Map {
		valueMap := valueReflect.Interface().(map[string]interface{})
		structValue := reflect.ValueOf(&result).Elem()

		for key, val := range valueMap {
			field := structValue.FieldByName(key)
			if field.IsValid() && field.CanSet() {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			}
		}
		return result, nil
	}
	return result, fmt.Errorf("cannot convert value of type %s to type %s", valueReflect.Type(), targetType)
}
