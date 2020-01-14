package validation

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-uuid"
)

// UUIDRegExp is a Regular Expression that can be used to validate UUIDs
var UUIDRegExp = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")

// UUID is a ValidateFunc that ensures a string can be parsed as UUID
func UUID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := uuid.ParseUUID(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid UUID, got %v", k, v))
	}

	return warnings, errors
}
