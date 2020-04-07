package schema

import "testing"

func TestInternalValidate(t *testing.T) {
	r := &ResourceImporter{
		State:        ImportStatePassthrough,
		StateContext: ImportStatePassthroughContext,
	}
	if err := r.InternalValidate(); err == nil {
		t.Fatal("ResourceImporter should not allow State and StateContext to be set")
	}
}
