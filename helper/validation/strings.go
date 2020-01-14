package validation

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// NoEmptyStrings validates that the string is not just whitespace characters (equal to [\r\n\t\f\v ])
func StringIsNotEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) == "" {
		return nil, []error{fmt.Errorf("expected %q it not be an empty string (whitespace not allowed)", k)}
	}

	return nil, nil
}

func StringIsBase64(i interface{}, k string) (warnings []string, errors []error) {
	// Empty string is not allowed
	if warnings, errors = StringIsNotEmpty(i, k); len(errors) > 0 {
		return
	}

	// NoEmptyStrings checks it is a string
	v, _ := i.(string)
	if _, err := base64.StdEncoding.DecodeString(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %w to be a base64 string, got %v", k, v))
	}

	return
}
