// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidationNoZeroValues(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "foo",
			f:   NoZeroValues,
		},
		{
			val: 1,
			f:   NoZeroValues,
		},
		{
			val: float64(1),
			f:   NoZeroValues,
		},
		{
			val:         "",
			f:           NoZeroValues,
			expectedErr: regexp.MustCompile("must not be empty"),
		},
		{
			val:         0,
			f:           NoZeroValues,
			expectedErr: regexp.MustCompile("must not be zero"),
		},
		{
			val:         float64(0),
			f:           NoZeroValues,
			expectedErr: regexp.MustCompile("must not be zero"),
		},
	})
}

func TestValidationAll(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "valid",
			f: All(
				StringLenBetween(5, 42),
				StringMatch(regexp.MustCompile(`[a-zA-Z0-9]+`), "value must be alphanumeric"),
			),
		},
		{
			val: "foo",
			f: All(
				StringLenBetween(5, 42),
				StringMatch(regexp.MustCompile(`[a-zA-Z0-9]+`), "value must be alphanumeric"),
			),
			expectedErr: regexp.MustCompile(`expected length of [\w]+ to be in the range \(5 - 42\), got foo`),
		},
		{
			val: "!!!!!",
			f: All(
				StringLenBetween(5, 42),
				StringMatch(regexp.MustCompile(`[a-zA-Z0-9]+`), "value must be alphanumeric"),
			),
			expectedErr: regexp.MustCompile("value must be alphanumeric"),
		},
	})
}

func TestValidationAny(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 43,
			f: Any(
				IntAtLeast(42),
				IntAtMost(5),
			),
		},
		{
			val: 4,
			f: Any(
				IntAtLeast(42),
				IntAtMost(5),
			),
		},
		{
			val: 7,
			f: Any(
				IntAtLeast(42),
				IntAtMost(5),
			),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(42\), got 7`),
		},
		{
			val: 7,
			f: Any(
				IntAtLeast(42),
				IntAtMost(5),
			),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at most \(5\), got 7`),
		},
	})
}

func TestToDiagFunc(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path        cty.Path
		val         interface{}
		f           schema.SchemaValidateDiagFunc
		expectedErr *regexp.Regexp
	}{
		"success-GetAttrStep-int": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
			},
			val: 43,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
		},
		"success-GetAttrStep-string": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
			},
			val: "foo",
			f: ToDiagFunc(All(
				StringLenBetween(1, 10),
				StringIsNotWhiteSpace,
			)),
		},
		"success-IndexStep-string": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
			},
			val: "foo",
			f:   ToDiagFunc(StringLenBetween(1, 10)),
		},
		"error-GetAttrStep-int-first": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
			},
			val: 7,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
			expectedErr: regexp.MustCompile(`expected test_property to be at least \(42\), got 7`),
		},
		"error-GetAttrStep-int-second": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
			},
			val: 7,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
			expectedErr: regexp.MustCompile(`expected test_property to be at most \(5\), got 7`),
		},
		"error-IndexStep-int": {
			path: cty.Path{
				cty.GetAttrStep{Name: "test_property"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
			},
			val:         7,
			f:           ToDiagFunc(IntAtLeast(42)),
			expectedErr: regexp.MustCompile(`expected test_property to be at least \(42\), got 7`),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.f(testCase.val, testCase.path)

			if !diags.HasError() && testCase.expectedErr == nil {
				return
			}

			if diags.HasError() && testCase.expectedErr == nil {
				t.Fatalf("expected to produce no errors, got %v", diags)
			}

			if !matchAnyDiagSummary(diags, testCase.expectedErr) {
				t.Fatalf("expected to produce error matching %q, got %v", testCase.expectedErr, diags)
			}
		})
	}
}
