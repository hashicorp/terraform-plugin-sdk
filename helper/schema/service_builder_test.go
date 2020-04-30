package schema

import (
	"testing"
)

func TestServiceBuilder_DataSources(t *testing.T) {
	testCases := []struct {
		name                string
		services            []ServiceRegistration
		expectedDataSources map[string]*Resource
		shouldError         bool
	}{
		{
			name: "none defined",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{},
				},
			},
			expectedDataSources: map[string]*Resource{},
			shouldError:         false,
		},
		{
			name: "single item",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello_world": {},
					},
				},
			},
			expectedDataSources: map[string]*Resource{
				"hello_world": {},
			},
			shouldError: false,
		},
		{
			name: "multiple items same service",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello": {},
						"world": {},
					},
				},
			},
			expectedDataSources: map[string]*Resource{
				"hello": {},
				"world": {},
			},
			shouldError: false,
		},
		{
			name: "multiple items different service",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello": {},
						"world": {},
					},
				},
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"rick":  {},
						"morty": {},
					},
				},
			},
			expectedDataSources: map[string]*Resource{
				"hello": {},
				"world": {},
				"rick":  {},
				"morty": {},
			},
			shouldError: false,
		},
		{
			name: "conflicting items different services",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello": {},
					},
				},
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello": {},
					},
				},
			},
			expectedDataSources: map[string]*Resource{},
			shouldError:         true,
		},
	}

	for _, testCase := range testCases {
		t.Logf("Testing %q..", testCase.name)
		builder := NewServiceBuilder(testCase.services)
		dataSources, err := builder.DataSources()
		if testCase.shouldError {
			if err != nil {
				continue
			}

			t.Fatalf("Expected an error but didn't get one!")
		}
		if !testCase.shouldError {
			if err == nil {
				continue
			}

			t.Fatalf("Expected no error but got: %+v", err)
		}

		if len(*dataSources) != len(testCase.expectedDataSources) {
			t.Fatalf("Expected %d Data Sources but got %d", len(testCase.expectedDataSources), len(*dataSources))
		}
		for k := range *dataSources {
			if _, ok := testCase.expectedDataSources[k]; !ok {
				t.Fatalf("Expected %q to be present but it wasn't!", k)
			}
		}
	}
}

func TestServiceBuilder_Resources(t *testing.T) {
	testCases := []struct {
		name              string
		services          []ServiceRegistration
		expectedResources map[string]*Resource
		shouldError       bool
	}{
		{
			name: "none defined",
			services: []ServiceRegistration{
				testServiceRegistration{
					resources: map[string]*Resource{},
				},
			},
			expectedResources: map[string]*Resource{},
			shouldError:       false,
		},
		{
			name: "single item",
			services: []ServiceRegistration{
				testServiceRegistration{
					resources: map[string]*Resource{
						"hello_world": {},
					},
				},
			},
			expectedResources: map[string]*Resource{
				"hello_world": {},
			},
			shouldError: false,
		},
		{
			name: "multiple items same service",
			services: []ServiceRegistration{
				testServiceRegistration{
					resources: map[string]*Resource{
						"hello": {},
						"world": {},
					},
				},
			},
			expectedResources: map[string]*Resource{
				"hello": {},
				"world": {},
			},
			shouldError: false,
		},
		{
			name: "multiple items different service",
			services: []ServiceRegistration{
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"hello": {},
						"world": {},
					},
				},
				testServiceRegistration{
					dataSources: map[string]*Resource{
						"rick":  {},
						"morty": {},
					},
				},
			},
			expectedResources: map[string]*Resource{
				"hello": {},
				"world": {},
				"rick":  {},
				"morty": {},
			},
			shouldError: false,
		},
		{
			name: "conflicting items different services",
			services: []ServiceRegistration{
				testServiceRegistration{
					resources: map[string]*Resource{
						"hello": {},
					},
				},
				testServiceRegistration{
					resources: map[string]*Resource{
						"hello": {},
					},
				},
			},
			expectedResources: map[string]*Resource{},
			shouldError:       true,
		},
	}

	for _, testCase := range testCases {
		t.Logf("Testing %q..", testCase.name)
		builder := NewServiceBuilder(testCase.services)
		resources, err := builder.Resources()
		if testCase.shouldError {
			if err != nil {
				continue
			}

			t.Fatalf("Expected an error but didn't get one!")
		}
		if !testCase.shouldError {
			if err == nil {
				continue
			}

			t.Fatalf("Expected no error but got: %+v", err)
		}

		if len(*resources) != len(testCase.expectedResources) {
			t.Fatalf("Expected %d Resources but got %d", len(testCase.expectedResources), len(*resources))
		}
		for k := range *resources {
			if _, ok := testCase.expectedResources[k]; !ok {
				t.Fatalf("Expected %q to be present but it wasn't!", k)
			}
		}
	}
}

type testServiceRegistration struct {
	dataSources map[string]*Resource
	resources   map[string]*Resource
}

func (r testServiceRegistration) Name() string {
	return "test"
}
func (r testServiceRegistration) WebsiteCategories() []string {
	return []string{}
}
func (r testServiceRegistration) SupportedDataSources() map[string]*Resource {
	return r.dataSources
}
func (r testServiceRegistration) SupportedResources() map[string]*Resource {
	return r.resources
}
