package validation

import (
	"regexp"
	"testing"
)

func TestValidateRFC3339TimeString(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "2018-03-01T00:00:00Z",
			f:   ValidateRFC3339TimeString,
		},
		{
			val: "2018-03-01T00:00:00-05:00",
			f:   ValidateRFC3339TimeString,
		},
		{
			val: "2018-03-01T00:00:00+05:00",
			f:   ValidateRFC3339TimeString,
		},
		{
			val:         "03/01/2018",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "03-01-2018",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "2018-03-01",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "2018-03-01T",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "2018-03-01T00:00:00",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "2018-03-01T00:00:00Z05:00",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
		{
			val:         "2018-03-01T00:00:00Z-05:00",
			f:           ValidateRFC3339TimeString,
			expectedErr: regexp.MustCompile(regexp.QuoteMeta(`invalid RFC3339 timestamp`)),
		},
	})
}
