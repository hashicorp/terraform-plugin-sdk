package validation

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MapKeyLenBetween returns a SchemaValidateFunc which tests if the provided value
// is of type map and the length of all keys are between min and max (inclusive)
func MapKeyLenBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %[1]q to be Map, got %[1]T", k))
			return warnings, errors
		}

		for key := range v {
			len := len(key)
			if len < min || len > max {
				errors = append(errors, fmt.Errorf("expected the length of all keys of %q to be in the range (%d - %d), got %q (length = %d)", k, min, max, key, len))
				return warnings, errors
			}
		}

		return warnings, errors
	}
}

// MapValueLenBetween returns a SchemaValidateFunc which tests if the provided value
// is of type map and the length of all values are between min and max (inclusive)
func MapValueLenBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %[1]q to be Map, got %[1]T", k))
			return warnings, errors
		}

		for _, val := range v {
			if _, ok := val.(string); !ok {
				errors = append(errors, fmt.Errorf("expected all values of %[1]q to be strings, found %[2]v (type = %[2]T)", k, val))
				return warnings, errors
			}
		}

		for _, val := range v {
			len := len(val.(string))
			if len < min || len > max {
				errors = append(errors, fmt.Errorf("expected the length of all values of %q to be in the range (%d - %d), got %q (length = %d)", k, min, max, val, len))
				return warnings, errors
			}
		}

		return warnings, errors
	}
}

// MapKeyMatch returns a SchemaValidateFunc which tests if the provided value
// is of type map and all keys match a given regexp. Optionally an error message
// can be provided to return something friendlier than "expected to match some globby regexp".
func MapKeyMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %[1]q to be Map, got %[1]T", k))
			return warnings, errors
		}

		for key := range v {
			if ok := r.MatchString(key); !ok {
				if message != "" {
					errors = append(errors, fmt.Errorf("invalid key %q for %q (%s)", key, k, message))
					return warnings, errors
				}

				errors = append(errors, fmt.Errorf("invalid key %q for %q (expected to match regular expression %q)", key, k, r))
				return warnings, errors
			}
		}

		return warnings, errors
	}
}

// MapValueMatch returns a SchemaValidateFunc which tests if the provided value
// is of type map and all values match a given regexp. Optionally an error message
// can be provided to return something friendlier than "expected to match some globby regexp".
func MapValueMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %[1]q to be Map, got %[1]T", k))
			return warnings, errors
		}

		for _, val := range v {
			if _, ok := val.(string); !ok {
				errors = append(errors, fmt.Errorf("expected all values of %[1]q to be strings, found %[2]v (type = %[2]T)", k, val))
				return warnings, errors
			}
		}

		for _, val := range v {
			if ok := r.MatchString(val.(string)); !ok {
				if message != "" {
					errors = append(errors, fmt.Errorf("invalid value %q for %q (%s)", val, k, message))
					return warnings, errors
				}

				errors = append(errors, fmt.Errorf("invalid value %q for %q (expected to match regular expression %q)", val, k, r))
				return warnings, errors
			}
		}

		return warnings, errors
	}
}
