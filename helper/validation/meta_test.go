package validation

import (
	"regexp"
	"testing"
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
			expectedErr: regexp.MustCompile("expected length of [\\w]+ to be in the range \\(5 - 42\\), got foo"),
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
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at least \\(42\\), got 7"),
		},
		{
			val: 7,
			f: Any(
				IntAtLeast(42),
				IntAtMost(5),
			),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at most \\(5\\), got 7"),
		},
	})
}

func TestToDiagFunc(t *testing.T) {
	runDiagTestCases(t, []diagTestCase{
		{
			val: 43,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
		},
		{
			val: "foo",
			f: ToDiagFunc(All(
				StringLenBetween(1, 10),
				StringIsNotWhiteSpace,
			)),
		},
		{
			val: 7,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at least \\(42\\), got 7"),
		},
		{
			val: 7,
			f: ToDiagFunc(Any(
				IntAtLeast(42),
				IntAtMost(5),
			)),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at most \\(5\\), got 7"),
		},
	})
}
