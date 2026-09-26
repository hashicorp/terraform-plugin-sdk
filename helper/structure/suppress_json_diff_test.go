// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package structure

import (
	"testing"
)

func TestSuppressJsonDiff(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		oldValue string
		newValue string
		expected bool
	}{
		"different-structure": {
			oldValue: `{ "enabled": true }`,
			newValue: `{ "enabled": true, "world": "round" }`,
			expected: false,
		},
		"different-value": {
			oldValue: `{ "enabled": true }`,
			newValue: `{ "enabled": false }`,
			expected: false,
		},
		"same": {
			oldValue: `{ "enabled": true }`,
			newValue: `{ "enabled": true }`,
			expected: true,
		},
		"same-whitespace": {
			oldValue: `{
				"enabled": true
			}`,
			newValue: `{ "enabled": true }`,
			expected: true,
		},
		"different-structure (array)": {
			oldValue: `[{ "enabled": true }]`,
			newValue: `[{ "enabled": true, "world": "round" }]`,
			expected: false,
		},
		"different-value (array)": {
			oldValue: `[{ "enabled": true }]`,
			newValue: `[{ "enabled": false }]`,
			expected: false,
		},
		"different order (array)": {
			oldValue: `[{ "enabled": true }, { "enabled": false }]`,
			newValue: `[{ "enabled": false }, { "enabled": true }]`,
			expected: false,
		},
		"same (array)": {
			oldValue: `[{ "enabled": true }]`,
			newValue: `[{ "enabled": true }]`,
			expected: true,
		},
		"same-whitespace (array)": {
			oldValue: `[{
				"enabled": true
			}]`,
			newValue: `[{ "enabled": true }]`,
			expected: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := SuppressJsonDiff("test", testCase.oldValue, testCase.newValue, nil)

			if actual != testCase.expected {
				t.Fatalf("expected %t, got %t", testCase.expected, actual)
			}
		})
	}
}
