package validation

import (
	"regexp"
	"testing"
)

func TestValidationIntBetween(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 1,
			f:   IntBetween(1, 1),
		},
		{
			val: 1,
			f:   IntBetween(0, 2),
		},
		{
			val:         1,
			f:           IntBetween(2, 3),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be in the range \\(2 - 3\\), got 1"),
		},
		{
			val:         "1",
			f:           IntBetween(2, 3),
			expectedErr: regexp.MustCompile("expected type of [\\w]+ to be int"),
		},
	})
}

func TestValidationIntAtLeast(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 1,
			f:   IntAtLeast(1),
		},
		{
			val: 1,
			f:   IntAtLeast(0),
		},
		{
			val:         1,
			f:           IntAtLeast(2),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at least \\(2\\), got 1"),
		},
		{
			val:         "1",
			f:           IntAtLeast(2),
			expectedErr: regexp.MustCompile("expected type of [\\w]+ to be int"),
		},
	})
}

func TestValidationIntAtMost(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 1,
			f:   IntAtMost(1),
		},
		{
			val: 1,
			f:   IntAtMost(2),
		},
		{
			val:         1,
			f:           IntAtMost(0),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be at most \\(0\\), got 1"),
		},
		{
			val:         "1",
			f:           IntAtMost(0),
			expectedErr: regexp.MustCompile("expected type of [\\w]+ to be int"),
		},
	})
}

func TestValidationIntInSlice(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: 42,
			f:   IntInSlice([]int{1, 42}),
		},
		{
			val:         42,
			f:           IntInSlice([]int{10, 20}),
			expectedErr: regexp.MustCompile("expected [\\w]+ to be one of \\[10 20\\], got 42"),
		},
		{
			val:         "InvalidValue",
			f:           IntInSlice([]int{10, 20}),
			expectedErr: regexp.MustCompile("expected type of [\\w]+ to be an integer"),
		},
	})
}
