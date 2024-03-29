// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	cases := []struct {
		Input    []string
		Expected []string
	}{
		{
			[]string{},
			[]string{},
		},
		{
			[]string{"x"},
			[]string{"x"},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		{
			[]string{"a", "a", "a"},
			[]string{"a"},
		},
		{
			[]string{"a", "b", "a", "b", "a", "a"},
			[]string{"a", "b"},
		},
		{
			[]string{"c", "b", "a", "c", "b"},
			[]string{"a", "b", "c"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("unique-%d", i), func(t *testing.T) {
			actual := uniqueStrings(tc.Input)
			if !reflect.DeepEqual(tc.Expected, actual) {
				t.Fatalf("Expected: %q\nGot: %q", tc.Expected, actual)
			}
		})
	}
}
