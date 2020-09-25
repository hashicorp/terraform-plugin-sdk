package acctest

import (
	"fmt"
	"regexp"
	"testing"
)

func TestSkipOnErrorContains(t *testing.T) {
	testCases := []struct {
		s        string
		err      error
		expected bool
	}{
		{
			s:        "Operations related to PublicDNS are not supported in this aws partition",
			err:      fmt.Errorf("Error: error creating Route53 Hosted Zone: InvalidInput: Operations related to PublicDNS are not supported in this aws partition.\nstatus code: 400, request id: 395ef7ef-be89-48a1-98ec-0bcb0a517825"),
			expected: true,
		},
		{
			s:        "Operations related to PublicDNS are not supported in this aws partition",
			err:      fmt.Errorf("Error: unrelated error: Smog Patrol. Had your emissions checked?\nstatus code: 400, request id: 395ef7ef-be89-48a1-98ec-0bcb0a517825"),
			expected: false,
		},
		{
			s:        "Operations related to PublicDNS are not supported in this aws partition",
			expected: false,
		},
	}

	for i, tc := range testCases {
		f := SkipOnErrorContains(tc.s)
		if f(tc.err) != tc.expected {
			t.Fatalf("expected test case %d to be %v but was %v (error portion %s, error message %s)", i, tc.expected, f(tc.err), tc.s, tc.err)
		}
	}
}

func TestSkipOnErrorRegexp(t *testing.T) {
	testCases := []struct {
		re       *regexp.Regexp
		err      error
		expected bool
	}{
		{
			re:       regexp.MustCompile(`Operations related to [a-zA-Z]+ are not supported in this aws partition`),
			err:      fmt.Errorf("Error: error creating Route53 Hosted Zone: InvalidInput: Operations related to PublicDNS are not supported in this aws partition.\nstatus code: 400, request id: 395ef7ef-be89-48a1-98ec-0bcb0a517825"),
			expected: true,
		},
		{
			re:       regexp.MustCompile(`Operations related to [a-zA-Z]+ are not supported in this aws partition`),
			err:      fmt.Errorf("Error: unrelated error, You on a scavenger hunt, or did I forget to pay my dinner check?\nstatus code: 400, request id: 395ef7ef-be89-48a1-98ec-0bcb0a517825"),
			expected: false,
		},
		{
			re:       regexp.MustCompile(`Operations related to [a-zA-Z]+ are not supported in this aws partition`),
			expected: false,
		},
	}

	for i, tc := range testCases {
		f := SkipOnErrorRegexp(tc.re)
		if f(tc.err) != tc.expected {
			t.Fatalf("expected test case %d to be %v but was %v (regexp %s, error message %s)", i, tc.expected, f(tc.err), tc.re, tc.err)
		}
	}
}
