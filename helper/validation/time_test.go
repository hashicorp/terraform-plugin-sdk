// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestValidationIsRFC3339Time(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotString": {
			Value: 7,
			Error: true,
		},
		"ValidDate": {
			Value: "2018-03-01T00:00:00Z",
			Error: false,
		},
		"ValidDateTime": {
			Value: "2018-03-01T00:00:00-05:00",
			Error: false,
		},
		"ValidDateTime2": {
			Value: "2018-03-01T00:00:00+05:00",
			Error: false,
		},
		"InvalidDateWithSlashes": {
			Value: "03/01/2018",
			Error: true,
		},
		"InvalidDateWithDashes": {
			Value: "03-01-2018",
			Error: true,
		},
		"InvalidDateWithDashes2": {
			Value: "2018-03-01",
			Error: true,
		},
		"InvalidDateWithT": {
			Value: "2018-03-01T",
			Error: true,
		},
		"DateTimeWithoutZone": {
			Value: "2018-03-01T00:00:00",
			Error: true,
		},
		"DateTimeWithZZone": {
			Value: "2018-03-01T00:00:00Z05:00",
			Error: true,
		},
		"DateTimeWithZZoneNeg": {
			Value: "2018-03-01T00:00:00Z-05:00",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsRFC3339Time(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsRFC3339Time(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsRFC3339Time(%s) did not error", tc.Value)
			}
		})
	}
}
