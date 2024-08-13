// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"regexp"
	"testing"
)

func TestValidationStringIsNotEmpty(t *testing.T) {
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
		"SingleSpace": {
			Value: " ",
			Error: false,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: false,
		},
		"NewLine": {
			Value: "\n",
			Error: false,
		},
		"MultipleSymbols": {
			Value: "-_-",
			Error: false,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: false,
		},
		"StartsWithWhitespace": {
			Value: "  7",
			Error: false,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsNotEmpty(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsNotEmpty(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsNotEmpty(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsNotWhitespace(t *testing.T) {
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
		"SingleSpace": {
			Value: " ",
			Error: true,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: true,
		},
		"CarriageReturn": {
			Value: "\r",
			Error: true,
		},
		"NewLine": {
			Value: "\n",
			Error: true,
		},
		"Tab": {
			Value: "\t",
			Error: true,
		},
		"FormFeed": {
			Value: "\f",
			Error: true,
		},
		"VerticalTab": {
			Value: "\v",
			Error: true,
		},
		"SingleChar": {
			Value: "\v",
			Error: true,
		},
		"MultipleChars": {
			Value: "-_-",
			Error: false,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: false,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: false,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsNotWhiteSpace(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsNotWhiteSpace(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsNotWhiteSpace(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsEmpty(t *testing.T) {
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
			Error: false,
		},
		"SingleSpace": {
			Value: " ",
			Error: true,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: true,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: true,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: true,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsEmpty(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsEmpty(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsEmpty(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsWhiteSpace(t *testing.T) {
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
			Error: false,
		},
		"SingleSpace": {
			Value: " ",
			Error: false,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: false,
		},
		"MultipleWhitespace": {
			Value: "  \t\n\f   ",
			Error: false,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: true,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: true,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsWhiteSpace(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsWhiteSpace(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsWhiteSpace(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringBytesBetween(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Min   int
		Max   int
		Error bool
	}{
		"NotStringNil": {
			Value: nil,
			Error: true,
		},
		"NotStringBool": {
			Value: bool(true),
			Error: true,
		},
		"NotStringInt": {
			Value: int(-1),
			Error: true,
		},
		"NotStringUint": {
			Value: uint(1),
			Error: true,
		},
		"NotStringByteSlice": {
			Value: []byte("hello"),
			Error: true,
		},
		"NotStringRuneSlice": {
			Value: []rune("こんにちは"),
			Error: true,
		},
		"NotStringFloat32": {
			Value: float32(1.23),
			Error: true,
		},
		"NotStringFloat64": {
			Value: float32(-1.23),
			Error: true,
		},
		"MinNegativeNumber": {
			Value: "MinNegativeNumber",
			Min:   -1,
			Max:   17,
			Error: true,
		},
		"MinZero": {
			Value: "MinZero",
			Min:   0,
			Max:   7,
			Error: false,
		},
		"MinPositiveNumber": {
			Value: "MinPositiveNumber",
			Min:   1,
			Max:   17,
			Error: false,
		},
		"MaxNegativeNumber": {
			Value: "MaxNegativeNumber",
			Min:   0,
			Max:   -1,
			Error: true,
		},
		"MaxZero": {
			Value: "",
			Min:   0,
			Max:   0,
			Error: false,
		},
		"MaxPositiveNumber": {
			Value: "MaxPositiveNumber",
			Min:   17,
			Max:   17,
			Error: false,
		},
		"MinLowerThanByteLength": {
			Value: "MinLowerThanByteLength",
			Min:   21,
			Max:   2147483647,
			Error: false,
		},
		"MinEqualByteLength": {
			Value: "MinEqualByteLength",
			Min:   18,
			Max:   2147483647,
			Error: false,
		},
		"MinGreaterThanByteLength": {
			Value: "MinGreaterThanByteLength",
			Min:   25,
			Max:   2147483647,
			Error: true,
		},
		"MaxLowerThanByteLength": {
			Value: "MaxLowerThanByteLength",
			Min:   0,
			Max:   21,
			Error: true,
		},
		"MaxEqualByteLength": {
			Value: "MaxEqualByteLength",
			Min:   0,
			Max:   18,
			Error: false,
		},
		"MaxGreaterThanByteLength": {
			Value: "MaxGreaterThanByteLength",
			Min:   0,
			Max:   25,
			Error: false,
		},
		"Empty": {
			Value: "",
			Min:   0,
			Max:   0,
			Error: false,
		},
		"WhiteSpace": {
			Value: " ",
			Min:   1,
			Max:   1,
			Error: false,
		},
		"Tab": {
			Value: "\t",
			Min:   1,
			Max:   1,
			Error: false,
		},
		"1ByteString": {
			Value: "Hello world!",
			Min:   12,
			Max:   12,
			Error: false,
		},
		"2BytesString": {
			Value: "αβγ",
			Min:   6,
			Max:   6,
			Error: false,
		},
		"3BytesString": {
			Value: "こんにちは世界！",
			Min:   24,
			Max:   24,
			Error: false,
		},
		"4BytesString": {
			Value: "👍",
			Min:   4,
			Max:   4,
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			v := StringBytesBetween(tc.Min, tc.Max)
			_, errors := v(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringBytesBetween(%d, %d) with '%v' produced an unexpected error", tc.Min, tc.Max, tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringBytesBetween(%d, %d) with '%v' did not error", tc.Min, tc.Max, tc.Value)
			}
		})
	}
}

func TestValidationStringRuneCountBetween(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Min   int
		Max   int
		Error bool
	}{
		"NotStringNil": {
			Value: nil,
			Error: true,
		},
		"NotStringBool": {
			Value: bool(true),
			Error: true,
		},
		"NotStringInt": {
			Value: int(-1),
			Error: true,
		},
		"NotStringUint": {
			Value: uint(1),
			Error: true,
		},
		"NotStringByteSlice": {
			Value: []byte("hello"),
			Error: true,
		},
		"NotStringRuneSlice": {
			Value: []rune("こんにちは"),
			Error: true,
		},
		"NotStringFloat32": {
			Value: float32(1.23),
			Error: true,
		},
		"NotStringFloat64": {
			Value: float32(-1.23),
			Error: true,
		},
		"MinNegativeNumber": {
			Value: "MinNegativeNumber",
			Min:   -1,
			Max:   17,
			Error: true,
		},
		"MinZero": {
			Value: "MinZero",
			Min:   0,
			Max:   7,
			Error: false,
		},
		"MinPositiveNumber": {
			Value: "MinPositiveNumber",
			Min:   1,
			Max:   17,
			Error: false,
		},
		"MaxNegativeNumber": {
			Value: "MaxNegativeNumber",
			Min:   0,
			Max:   -1,
			Error: true,
		},
		"MaxZero": {
			Value: "",
			Min:   0,
			Max:   0,
			Error: false,
		},
		"MaxPositiveNumber": {
			Value: "MaxPositiveNumber",
			Min:   17,
			Max:   17,
			Error: false,
		},
		"MinLowerThanByteLength": {
			Value: "MinLowerThanByteLength",
			Min:   21,
			Max:   2147483647,
			Error: false,
		},
		"MinEqualByteLength": {
			Value: "MinEqualByteLength",
			Min:   18,
			Max:   2147483647,
			Error: false,
		},
		"MinGreaterThanByteLength": {
			Value: "MinGreaterThanByteLength",
			Min:   25,
			Max:   2147483647,
			Error: true,
		},
		"MaxLowerThanByteLength": {
			Value: "MaxLowerThanByteLength",
			Min:   0,
			Max:   21,
			Error: true,
		},
		"MaxEqualByteLength": {
			Value: "MaxEqualByteLength",
			Min:   0,
			Max:   18,
			Error: false,
		},
		"MaxGreaterThanByteLength": {
			Value: "MaxGreaterThanByteLength",
			Min:   0,
			Max:   25,
			Error: false,
		},
		"Empty": {
			Value: "",
			Min:   0,
			Max:   0,
			Error: false,
		},
		"WhiteSpace": {
			Value: " ",
			Min:   1,
			Max:   1,
			Error: false,
		},
		"Tab": {
			Value: "\t",
			Min:   1,
			Max:   1,
			Error: false,
		},
		"1ByteString": {
			Value: "Hello world!",
			Min:   12,
			Max:   12,
			Error: false,
		},
		"2BytesString": {
			Value: "αβγ",
			Min:   3,
			Max:   3,
			Error: false,
		},
		"3BytesString": {
			Value: "こんにちは世界！",
			Min:   8,
			Max:   8,
			Error: false,
		},
		"4BytesString": {
			Value: "👍",
			Min:   1,
			Max:   1,
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			v := StringRuneCountBetween(tc.Min, tc.Max)
			_, errors := v(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringRuneCountBetween(%d, %d) with '%v' produced an unexpected error", tc.Min, tc.Max, tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringRuneCountBetween(%d, %d) with '%v' did not error", tc.Min, tc.Max, tc.Value)
			}
		})
	}
}

func TestValidationStringIsBase64(t *testing.T) {
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
		"NotBase64": {
			Value: "Do'h!",
			Error: true,
		},
		"Base64": {
			Value: "RG8naCE=",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsBase64(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsBase64(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsBase64(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringInSlice(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "ValidValue",
			f:   StringInSlice([]string{"ValidValue", "AnotherValidValue"}, false),
		},
		// ignore case
		{
			val: "VALIDVALUE",
			f:   StringInSlice([]string{"ValidValue", "AnotherValidValue"}, true),
		},
		{
			val:         "VALIDVALUE",
			f:           StringInSlice([]string{"ValidValue", "AnotherValidValue"}, false),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be one of \["ValidValue" "AnotherValidValue"\], got VALIDVALUE`),
		},
		{
			val:         "InvalidValue",
			f:           StringInSlice([]string{"ValidValue", "AnotherValidValue"}, false),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be one of \["ValidValue" "AnotherValidValue"\], got InvalidValue`),
		},
		{
			val:         1,
			f:           StringInSlice([]string{"ValidValue", "AnotherValidValue"}, false),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationStringNotInSlice(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "ValidValue",
			f:   StringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, false),
		},
		// ignore case
		{
			val: "VALIDVALUE",
			f:   StringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, true),
		},
		{
			val:         "AnotherInvalidValue",
			f:           StringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, false),
			expectedErr: regexp.MustCompile(`expected [\w]+ to not be any of \[InvalidValue AnotherInvalidValue\], got AnotherInvalidValue`),
		},
		// ignore case
		{
			val:         "INVALIDVALUE",
			f:           StringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, true),
			expectedErr: regexp.MustCompile(`expected [\w]+ to not be any of \[InvalidValue AnotherInvalidValue\], got INVALIDVALUE`),
		},
		{
			val:         1,
			f:           StringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, false),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationStringMatch(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "foobar",
			f:   StringMatch(regexp.MustCompile(".*foo.*"), ""),
		},
		{
			val:         "bar",
			f:           StringMatch(regexp.MustCompile(".*foo.*"), ""),
			expectedErr: regexp.MustCompile(`expected value of [\w]+ to match regular expression ` + regexp.QuoteMeta(`".*foo.*"`)),
		},
		{
			val:         "bar",
			f:           StringMatch(regexp.MustCompile(".*foo.*"), "value must contain foo"),
			expectedErr: regexp.MustCompile(`invalid value for [\w]+ \(value must contain foo\)`),
		},
	})
}

func TestValidationStringDoesNotMatch(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "foobar",
			f:   StringDoesNotMatch(regexp.MustCompile(".*baz.*"), ""),
		},
		{
			val:         "bar",
			f:           StringDoesNotMatch(regexp.MustCompile(".*bar.*"), ""),
			expectedErr: regexp.MustCompile(`expected value of [\w]+ to not match regular expression ` + regexp.QuoteMeta(`".*bar.*"`)),
		},
		{
			val:         "bar",
			f:           StringDoesNotMatch(regexp.MustCompile(".*bar.*"), "value must not contain foo"),
			expectedErr: regexp.MustCompile(`invalid value for [\w]+ \(value must not contain foo\)`),
		},
	})
}

func TestStringIsJSON(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    `{0:"1"}`,
			ErrCount: 1,
		},
		{
			Value:    `{'abc':1}`,
			ErrCount: 1,
		},
		{
			Value:    `{"def":}`,
			ErrCount: 1,
		},
		{
			Value:    `{"xyz":[}}`,
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := StringIsJSON(tc.Value, "json")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    ``,
			ErrCount: 0,
		},
		{
			Value:    `{}`,
			ErrCount: 0,
		},
		{
			Value:    `{"abc":["1","2"]}`,
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := StringIsJSON(tc.Value, "json")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}
}

func TestStringDoesNotContainAny(t *testing.T) {
	chars := "|:/"

	validStrings := []string{
		"HelloWorld",
		"ABC_*&^%123",
	}
	for _, v := range validStrings {
		_, errors := StringDoesNotContainAny(chars)(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should not contain any of %q", v, chars)
		}
	}

	invalidStrings := []string{
		"Hello/World",
		"ABC|123",
		"This will fail:",
		chars,
	}
	for _, v := range invalidStrings {
		_, errors := StringDoesNotContainAny(chars)(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should contain one of %q", v, chars)
		}
	}
}

func TestStringIsValidRegExp(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: ".*foo.*",
			f:   StringIsValidRegExp,
		},
		{
			val:         "foo(bar",
			f:           StringIsValidRegExp,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("error parsing regexp: missing closing ): `foo(bar`")),
		},
	})
}
