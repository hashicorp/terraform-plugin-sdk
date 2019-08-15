package terraform

func simpleTestSchemas() *Schemas {
	provider := simpleMockProvider()
	return &Schemas{
		Providers: map[string]*ProviderSchema{
			"test": provider.GetSchemaReturn,
		},
	}
}
