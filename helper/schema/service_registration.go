package schema

// ServiceRegistration returns the MetaData for a given Service, which
// is a collection of related Data Sources and Resources.
//
// Services allow larger providers to become more maintainable by
// grouping relevant functionality together - which allows helper methods
// to be better scoped and rebuilding only the components that have changed.
type ServiceRegistration interface {
	// Name is the name of this Service
	Name() string

	// WebsiteCategories returns a list of categories used in the Website
	WebsiteCategories() []string

	// SupportedDataSources returns the supported Data Sources supported by this Service
	SupportedDataSources() map[string]*Resource

	// SupportedResources returns the supported Resources supported by this Service
	SupportedResources() map[string]*Resource
}
