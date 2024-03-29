// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestValidationIsUUID(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotString": {
			Value: 7,
			Error: true,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"InvalidUuid": {
			Value: "00000000-0000-123-0000-000000000000",
			Error: true,
		},
		"ValidUuidWithOutDashs": {
			Value: "12345678123412341234123456789012",
			Error: true,
		},
		"ValidUuid": {
			Value: "00000000-0000-0000-0000-000000000000",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsUUID(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsUUID(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsUUID(%s) did not error", tc.Value)
			}
		})
	}
}
