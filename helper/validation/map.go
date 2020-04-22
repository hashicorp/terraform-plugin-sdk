package validation

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MapKeyLenBetween returns a SchemaValidateDiagFunc which tests if the provided value
// is of type map and the length of all keys are between min and max (inclusive)
func MapKeyLenBetween(min, max int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		for _, key := range sortedKeys(v.(map[string]interface{})) {
			len := len(key)
			if len < min || len > max {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Bad map key length",
					Detail:        fmt.Sprintf("Map key lengths should be in the range (%d - %d): %s (length = %d)", min, max, key, len),
					AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
				})
			}
		}

		return diags
	}
}

// MapValueLenBetween returns a SchemaValidateDiagFunc which tests if the provided value
// is of type map and the length of all values are between min and max (inclusive)
func MapValueLenBetween(min, max int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		m := v.(map[string]interface{})

		for _, key := range sortedKeys(m) {
			val := m[key]

			if _, ok := val.(string); !ok {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Bad map value type",
					Detail:        fmt.Sprintf("Map values should be strings: %s => %v (type = %T)", key, val, val),
					AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
				})
				continue
			}

			len := len(val.(string))
			if len < min || len > max {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Bad map value length",
					Detail:        fmt.Sprintf("Map value lengths should be in the range (%d - %d): %s => %v (length = %d)", min, max, key, val, len),
					AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
				})
			}
		}

		return diags
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

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))

	i := 0
	for key := range m {
		keys[i] = key
		i++
	}

	sort.Strings(keys)

	return keys
}
