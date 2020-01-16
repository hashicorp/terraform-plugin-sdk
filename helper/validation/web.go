package validation

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// IsURLWithHTTPS is a SchemaValidateFunc which tests if the provided value is of type string and a valid IPv6 address
func IsURLWithHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return IsURLWithScheme([]string{"https"})(i, k)
}

// IsURLWithHTTPorHTTPS is a SchemaValidateFunc which tests if the provided value is of type string and a valid IPv6 address
func IsURLWithHTTPorHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return IsURLWithScheme([]string{"http", "https"})(i, k)
}

// IsURLWithScheme is a SchemaValidateFunc which tests if the provided value is of type string and a valid IPv6 address
func IsURLWithScheme(validSchemes []string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
			return
		}

		if v == "" {
			errors = append(errors, fmt.Errorf("expected %q url to not be empty", k))
			return
		}

		u, err := url.Parse(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q url is in an invalid format: %q (%+v)", k, v, err))
			return
		}

		if u.Host == "" {
			errors = append(errors, fmt.Errorf("%q url has no host: %q", k, v))
			return
		}

		for _, s := range validSchemes {
			if u.Scheme == s {
				return //last check so just return
			}
		}

		errors = append(errors, fmt.Errorf("expected %q url %q to have a schema of: %q", k, v, strings.Join(validSchemes, ",")))
		return
	}
}
