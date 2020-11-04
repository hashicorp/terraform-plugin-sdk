package schema

import (
	"errors"
	"fmt"
)

var ProviderReconfiguredError = errors.New("reconfigured")

// ProviderConfigurationError is the error type for provider configuration
type ProviderConfigurationError struct {
	Provider *Provider
	Err      error
}

func (e *ProviderConfigurationError) Error() string {
	return fmt.Sprintf("error configuring provider: %s", e.Err.Error())
}

func (e *ProviderConfigurationError) Unwrap() error {
	return e.Err
}
