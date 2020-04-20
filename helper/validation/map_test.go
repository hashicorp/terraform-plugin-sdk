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

			if len(diags) != len(tc.ExpectedDiags) {
				t.Fatalf("%s: wrong number of diags, expected %d, got %d", tn, len(tc.ExpectedDiags), len(diags))
			}
			for j := range diags {
				if diags[j].Severity != tc.ExpectedDiags[j].Severity {
					t.Fatalf("%s: expected severity %v, got %v", tn, tc.ExpectedDiags[j].Severity, diags[j].Severity)
				}
				if !diags[j].AttributePath.Equals(tc.ExpectedDiags[j].AttributePath) {
					t.Fatalf("%s: attribute paths do not match expected: %v, got %v", tn, tc.ExpectedDiags[j].AttributePath, diags[j].AttributePath)
				}
			}
		})
	}
}

func TestValidationMapValueLenBetween(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotMap": {
			Value: "the map is a lie",
			Error: true,
		},
		"NotStringValue": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": 123456,
			},
			Error: true,
		},
		"TooLong": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": "123456",
			},
			Error: true,
		},
		"TooShort": {
			Value: map[string]interface{}{
				"ABC": "123",
				"U":   "1",
			},
			Error: true,
		},
		"AllGood": {
			Value: map[string]interface{}{
				"AB":    "12",
				"UVWXY": "12345",
			},
			Error: false,
		},
	}

	fn := MapValueLenBetween(2, 5)

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := fn(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("MapValueLenBetween(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("MapValueLenBetween(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationMapKeyMatch(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotMap": {
			Value: "the map is a lie",
			Error: true,
		},
		"NoMatch": {
			Value: map[string]interface{}{
				"ABC":    "123",
				"UVWXYZ": "123456",
			},
			Error: true,
		},
		"AllGood": {
			Value: map[string]interface{}{
				"AB":    "12",
				"UVABY": "12345",
			},
			Error: false,
		},
	}

	fn := MapKeyMatch(regexp.MustCompile(".*AB.*"), "")

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := fn(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("MapKeyMatch(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("MapKeyMatch(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationValueKeyMatch(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotMap": {
			Value: "the map is a lie",
			Error: true,
		},
		"NotStringValue": {
			Value: map[string]interface{}{
				"MNO":    "123",
				"UVWXYZ": 123456,
			},
			Error: true,
		},
		"NoMatch": {
			Value: map[string]interface{}{
				"MNO":    "ABC",
				"UVWXYZ": "UVWXYZ",
			},
			Error: true,
		},
		"AllGood": {
			Value: map[string]interface{}{
				"MNO":    "ABC",
				"UVWXYZ": "UVABY",
			},
			Error: false,
		},
	}

	fn := MapValueMatch(regexp.MustCompile(".*AB.*"), "")

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := fn(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("MapValueMatch(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("MapValueMatch(%s) did not error", tc.Value)
			}
		})
	}
}
