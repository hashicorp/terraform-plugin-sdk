package schema

import "fmt"

// ServiceBuilder is a helper type which allows accessing a canonical list of
// Data Sources and Resources supported by this Provider, or Service Package.
type ServiceBuilder struct {
	services []ServiceRegistration
}

// NewServiceBuilder returns a ServiceBuilder which allows accessing a canonical
// list of Data Sources and Resources from a list of Service Registrations
func NewServiceBuilder(services []ServiceRegistration) ServiceBuilder {
	return ServiceBuilder{
		services: services,
	}
}

// DataSources returns a canonical list of Data Sources supported by the
// Services registered with this Service Builder.
func (s ServiceBuilder) DataSources() (*map[string]*Resource, error) {
	dataSources := make(map[string]*Resource)

	for _, service := range s.services {
		for k, v := range service.SupportedDataSources() {
			if existing := dataSources[k]; existing != nil {
				return nil, fmt.Errorf("An existing Data Source exists for %q", k)
			}

			dataSources[k] = v
		}
	}

	return &dataSources, nil
}

// Resources returns a canonical list of Resources supported by the
// Services registered with this Service Builder.
func (s ServiceBuilder) Resources() (*map[string]*Resource, error) {
	resources := make(map[string]*Resource)

	for _, service := range s.services {
		for k, v := range service.SupportedResources() {
			if existing := resources[k]; existing != nil {
				return nil, fmt.Errorf("An existing Resource exists for %q", k)
			}

			resources[k] = v
		}
	}

	return &resources, nil
}
