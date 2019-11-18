package schema

func (d *ResourceDiff) GetFromConfig(key string) interface{} {
	// TODO: drop feature flag when protocol 4 is stripped from the SDK
	SetProto5()

	return nil
}

func (d *ResourceDiff) GetFromState(key string) interface{} {
	// TODO: drop feature flag when protocol 4 is stripped from the SDK
	SetProto5()

	return nil
}
