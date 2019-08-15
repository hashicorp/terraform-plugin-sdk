package terraform

import (
	"fmt"

	"github.com/hashicorp/terraform/providers"
)

// contextComponentFactory is the interface that Context uses
// to initialize various components such as providers.
// This factory gets more information than the raw maps using to initialize
// a Context. This information is used for debugging.
type contextComponentFactory interface {
	// ResourceProvider creates a new ResourceProvider with the given
	// type. The "uid" is a unique identifier for this provider being
	// initialized that can be used for internal tracking.
	ResourceProvider(typ, uid string) (providers.Interface, error)
	ResourceProviders() []string
}

// basicComponentFactory just calls a factory from a map directly.
type basicComponentFactory struct {
	providers map[string]providers.Factory
}

func (c *basicComponentFactory) ResourceProviders() []string {
	result := make([]string, len(c.providers))
	for k := range c.providers {
		result = append(result, k)
	}

	return result
}

func (c *basicComponentFactory) ResourceProvider(typ, uid string) (providers.Interface, error) {
	f, ok := c.providers[typ]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", typ)
	}

	return f()
}
