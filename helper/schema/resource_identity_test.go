// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package schema

import "testing"

func TestResourceIdentity_SchemaMap_handles_nil_identity(t *testing.T) {
	var ri *ResourceIdentity
	if ri.SchemaMap() != nil {
		t.Fatal("expected nil schema map")
	}
}
