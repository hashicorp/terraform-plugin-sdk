package validation

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// StringIsNotEmpty is a ValidateFunc that ensures a string is not empty or consisting entirely of whitespace characters
func StringIsNotEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string", k)}
	}

	return nil, nil
}

// StringIsNotEmpty is a ValidateFunc that ensures a string is not empty or consisting entirely of whitespace characters
func StringIsNotWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string or whitespace", k)}
	}

	return nil, nil
}

// StringIsEmpty is a ValidateFunc that ensures a string has no characters
func StringIsEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string: got %v", k, v)}
	}

	return nil, nil
}

// StringIsEmpty is a ValidateFunc that ensures a string is composed of entirely whitespace
func StringIsWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string or whitespace: got %v", k, v)}
	}

	return nil, nil
}

// StringIsBase64 is a ValidateFunc that ensures a string can be parsed as Base64
func StringIsBase64(i interface{}, k string) (warnings []string, errors []error) {
	// Empty string is not allowed
	if warnings, errors = StringIsNotEmpty(i, k); len(errors) > 0 {
		return
	}

	// NoEmptyStrings checks it is a string
	v, _ := i.(string)
	if _, err := base64.StdEncoding.DecodeString(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a base64 string, got %v", k, v))
	}

	return
}
