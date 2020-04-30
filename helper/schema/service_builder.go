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
	dataSourceRegistrationSource := make(map[string]string)

	for _, service := range s.services {
		for k, v := range service.SupportedDataSources() {
			if existing := dataSources[k]; existing != nil {
				existingRegName := dataSourceRegistrationSource[k]
				return nil, fmt.Errorf("Both %q and %q register Data Source %q!", existingRegName, service.Name(), k)
			}

			dataSources[k] = v
			dataSourceRegistrationSource[k] = service.Name()
		}
	}

	return &dataSources, nil
}

// Resources returns a canonical list of Resources supported by the
// Services registered with this Service Builder.
func (s ServiceBuilder) Resources() (*map[string]*Resource, error) {
	resources := make(map[string]*Resource)
	resourceRegistrationSource := make(map[string]string)

	for _, service := range s.services {
		for k, v := range service.SupportedResources() {
			if existing := resources[k]; existing != nil {
				existingRegName := resourceRegistrationSource[k]
				return nil, fmt.Errorf("Both %q and %q register Resource %q!", existingRegName, service.Name(), k)
			}

			resources[k] = v
			resourceRegistrationSource[k] = service.Name()
		}
	}

	return &resources, nil
}
