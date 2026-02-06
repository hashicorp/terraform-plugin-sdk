// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import "testing"

func TestResourceIdentity_SchemaMap_handles_nil_identity(t *testing.T) {
	var ri *ResourceIdentity
	if ri.SchemaMap() != nil {
		t.Fatal("expected nil schema map")
	}
}
