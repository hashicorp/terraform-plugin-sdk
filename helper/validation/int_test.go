// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
			expectedErr: regexp.MustCompile(`expected [\w]+ to be in the range \(2 - 3\), got 1`),
		},
		{
			val:         "1",
			f:           IntBetween(2, 3),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be integer`),
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
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(2\), got 1`),
		},
		{
			val:         "1",
			f:           IntAtLeast(2),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be integer`),
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
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at most \(0\), got 1`),
		},
		{
			val:         "1",
			f:           IntAtMost(0),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be integer`),
		},
	})
}

func TestValidationIntDivisibleBy(t *testing.T) {
	cases := map[string]struct {
		Value   interface{}
		Divisor int
		Error   bool
	}{
		"NotInt": {
			Value:   "words",
			Divisor: 2,
			Error:   true,
		},
		"NotDivisible": {
			Value:   15,
			Divisor: 7,
			Error:   true,
		},
		"Divisible": {
			Value:   14,
			Divisor: 7,
			Error:   false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IntDivisibleBy(tc.Divisor)(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IntDivisibleBy(%v) produced an unexpected error for %v", tc.Divisor, tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IntDivisibleBy(%v) did not error for %v", tc.Divisor, tc.Value)
			}
		})
	}
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
			expectedErr: regexp.MustCompile(`expected [\w]+ to be one of \[10 20\], got 42`),
		},
		{
			val:         "InvalidValue",
			f:           IntInSlice([]int{10, 20}),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be integer`),
		},
	})
}

func TestValidationIntNotInSlice(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Slice []int
		Error bool
	}{
		"NotInt": {
			Value: "words",
			Slice: []int{7, 77},
			Error: true,
		},
		"NotInSlice": {
			Value: 1,
			Slice: []int{7, 77},
			Error: false,
		},
		"InSlice": {
			Value: 7,
			Slice: []int{7, 77},
			Error: true,
		},
		"InSliceOfOne": {
			Value: 7,
			Slice: []int{7},
			Error: true,
		},
		"NotInSliceOfOne": {
			Value: 1,
			Slice: []int{7},
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IntNotInSlice(tc.Slice)(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IntNotInSlice(%v) produced an unexpected error for %v", tc.Slice, tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IntNotInSlice(%v) did not error for %v", tc.Slice, tc.Value)
			}
		})
	}
}
