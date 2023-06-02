// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestValidationMapKeyLenBetween(t *testing.T) {
	cases := map[string]struct {
		Value         interface{}
		ExpectedDiags diag.Diagnostics
	}{
		"TooLong": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": "123456",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"TooShort": {
			Value: map[string]interface{}{
				"ABC": "123",
				"U":   "1",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("U")}),
				},
			},
		},
		"TooLongAndTooShort": {
			Value: map[string]interface{}{
				"UVWXYZ": "123456",
				"ABC":    "123",
				"U":      "1",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("U")}),
				},
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"AllGood": {
			Value: map[string]interface{}{
				"AB":    "12",
				"UVWXY": "12345",
			},
			ExpectedDiags: nil,
		},
	}

	fn := MapKeyLenBetween(2, 5)

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			diags := fn(tc.Value, cty.Path{})

			checkDiagnostics(t, tn, diags, tc.ExpectedDiags)
		})
	}
}

func TestValidationMapValueLenBetween(t *testing.T) {
	cases := map[string]struct {
		Value         interface{}
		ExpectedDiags diag.Diagnostics
	}{
		"NotStringValue": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": 123456,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"TooLong": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": "123456",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"TooShort": {
			Value: map[string]interface{}{
				"ABC": "123",
				"U":   "1",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("U")}),
				},
			},
		},
		"TooLongAndTooShort": {
			Value: map[string]interface{}{
				"UVWXYZ": "123456",
				"ABC":    "123",
				"U":      "1",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("U")}),
				},
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"AllGood": {
			Value: map[string]interface{}{
				"AB":    "12",
				"UVWXY": "12345",
			},
			ExpectedDiags: nil,
		},
	}

	fn := MapValueLenBetween(2, 5)

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			diags := fn(tc.Value, cty.Path{})

			checkDiagnostics(t, tn, diags, tc.ExpectedDiags)
		})
	}
}

func TestValidationMapKeyMatch(t *testing.T) {
	cases := map[string]struct {
		Value         interface{}
		ExpectedDiags diag.Diagnostics
	}{
		"NoMatch": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": "123456",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"AllGood": {
			Value: map[string]interface{}{
				"AB":    "12",
				"UVABY": "12345",
			},
			ExpectedDiags: nil,
		},
	}

	fn := MapKeyMatch(regexp.MustCompile(".*AB.*"), "")

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			diags := fn(tc.Value, cty.Path{})

			checkDiagnostics(t, tn, diags, tc.ExpectedDiags)
		})
	}
}

func TestValidationValueKeyMatch(t *testing.T) {
	cases := map[string]struct {
		Value         interface{}
		ExpectedDiags diag.Diagnostics
	}{
		"NotStringValue": {
			Value: map[string]interface{}{
				"MNO":    "ABC",
				"UVWXYZ": 123456,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"NoMatch": {
			Value: map[string]interface{}{
				"MNO":    "ABC",
				"UVWXYZ": "UVWXYZ",
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"BothBad": {
			Value: map[string]interface{}{
				"MNO":    "123",
				"UVWXYZ": 123456,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("MNO")}),
				},
				{
					Severity:      diag.Error,
					AttributePath: append(cty.Path{}, cty.IndexStep{Key: cty.StringVal("UVWXYZ")}),
				},
			},
		},
		"AllGood": {
			Value: map[string]interface{}{
				"MNO":    "ABC",
				"UVWXYZ": "UVABY",
			},
			ExpectedDiags: nil,
		},
	}

	fn := MapValueMatch(regexp.MustCompile(".*AB.*"), "")

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			diags := fn(tc.Value, cty.Path{})

			checkDiagnostics(t, tn, diags, tc.ExpectedDiags)
		})
	}
}

func checkDiagnostics(t *testing.T, tn string, got, expected diag.Diagnostics) {
	if len(got) != len(expected) {
		t.Fatalf("%s: wrong number of diags, expected %d, got %d", tn, len(expected), len(got))
	}
	for j := range got {
		if got[j].Severity != expected[j].Severity {
			t.Fatalf("%s: expected severity %v, got %v", tn, expected[j].Severity, got[j].Severity)
		}
		if !got[j].AttributePath.Equals(expected[j].AttributePath) {
			t.Fatalf("%s: attribute paths do not match expected: %v, got %v", tn, expected[j].AttributePath, got[j].AttributePath)
		}
	}
}
