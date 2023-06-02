// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestValidationListOfUniqueStrings(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotList": {
			Value: "the list is a lie",
			Error: true,
		},
		"NotListOfString": {
			Value: []interface{}{"seven", 7},
			Error: true,
		},
		"NonUniqueStrings": {
			Value: []interface{}{"kt", "is", "kt"},
			Error: true,
		},
		"UniqueStrings": {
			Value: []interface{}{"thanks", "for", "all", "the", "fish"},
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := ListOfUniqueStrings(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("ListOfUniqueStrings(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("ListOfUniqueStrings(%s) did not error", tc.Value)
			}
		})
	}
}
