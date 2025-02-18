package schema

type IdentityData struct {
	// raw identity data will be stored internally
}

// Reading/writing data will be similar to the *schema.ResourceData flatmap
func (d *IdentityData) Get(key string) interface{} {
	panic("not implemented")
}
func (d *IdentityData) GetOk(key string) (interface{}, bool) {
	panic("not implemented")
}
func (d *IdentityData) Set(key string, value interface{}) {
	panic("not implemented")
}
