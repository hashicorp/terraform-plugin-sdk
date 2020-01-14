package validation

import (
	"regexp"
	"testing"
)

func TestValidationSingleIP(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "172.10.10.10",
			f:   SingleIP(),
		},
		{
			val:         "1.1.1",
			f:           SingleIP(),
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("expected test_property to contain a valid IP, got:")),
		},
		{
			val:         "1.1.1.0/20",
			f:           SingleIP(),
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("expected test_property to contain a valid IP, got:")),
		},
		{
			val:         "256.1.1.1",
			f:           SingleIP(),
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("expected test_property to contain a valid IP, got:")),
		},
	})
}

func TestValidationIPRange(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "172.10.10.10-172.10.10.12",
			f:   IPRange(),
		},
		{
			val:         "172.10.10.20",
			f:           IPRange(),
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("expected test_property to contain a valid IP range, got:")),
		},
		{
			val:         "172.10.10.20-172.10.10.12",
			f:           IPRange(),
			expectedErr: regexp.MustCompile(regexp.QuoteMeta("expected test_property to contain a valid IP range, got:")),
		},
	})
}
