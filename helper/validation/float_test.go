// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateFloatBetween(t *testing.T) {
	cases := map[string]struct {
		Value                  interface{}
		ValidateFunc           schema.SchemaValidateFunc
		ExpectValidationErrors bool
	}{
		"accept valid value": {
			Value:                  1.5,
			ValidateFunc:           FloatBetween(1.0, 2.0),
			ExpectValidationErrors: false,
		},
		"accept valid value inclusive upper bound": {
			Value:                  1.0,
			ValidateFunc:           FloatBetween(0.0, 1.0),
			ExpectValidationErrors: false,
		},
		"accept valid value inclusive lower bound": {
			Value:                  0.0,
			ValidateFunc:           FloatBetween(0.0, 1.0),
			ExpectValidationErrors: false,
		},
		"reject out of range value": {
			Value:                  -1.0,
			ValidateFunc:           FloatBetween(0.0, 1.0),
			ExpectValidationErrors: true,
		},
		"reject incorrectly typed value": {
			Value:                  1,
			ValidateFunc:           FloatBetween(0.0, 1.0),
			ExpectValidationErrors: true,
		},
	}

	for tn, tc := range cases {
		_, errors := tc.ValidateFunc(tc.Value, tn)
		if len(errors) > 0 && !tc.ExpectValidationErrors {
			t.Errorf("%s: unexpected errors %s", tn, errors)
		} else if len(errors) == 0 && tc.ExpectValidationErrors {
			t.Errorf("%s: expected errors but got none", tn)
		}
	}
}

func TestValidateFloatAtLeast(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 2.5,
			f:   FloatAtLeast(1.5),
		},
		{
			val: -1.0,
			f:   FloatAtLeast(-1.5),
		},
		{
			val:         1.5,
			f:           FloatAtLeast(2.5),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(2\.5\d*\), got 1\.5\d*`),
		},
		{
			val:         "2.5",
			f:           FloatAtLeast(1.5),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be float`),
		},
	})
}

func TestValidateFloatAtMost(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 2.5,
			f:   FloatAtMost(3.5),
		},
		{
			val: -1.0,
			f:   FloatAtMost(-0.5),
		},
		{
			val:         2.5,
			f:           FloatAtMost(1.5),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at most \(1\.5\d*\), got 2\.5\d*`),
		},
		{
			val:         "2.5",
			f:           FloatAtMost(3.5),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be float`),
		},
	})
}
