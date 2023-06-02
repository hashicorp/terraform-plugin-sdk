// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestValidationIsURLWithHTTPS(t *testing.T) {
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
		"NotUrl": {
			Value: "this is not a url",
			Error: true,
		},
		"BareUrl": {
			Value: "www.example.com",
			Error: true,
		},
		"FtpUrl": {
			Value: "ftp://www.example.com",
			Error: true,
		},
		"HttpUrl": {
			Value: "http://www.example.com",
			Error: true,
		},
		"HttpsUrl": {
			Value: "https://www.example.com",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsURLWithHTTPS(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsURLWithHTTPS(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsURLWithHTTPS(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationIsURLWithHTTPorHTTPS(t *testing.T) {
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
		"NotUrl": {
			Value: "this is not a url",
			Error: true,
		},
		"BareUrl": {
			Value: "www.example.com",
			Error: true,
		},
		"FtpUrl": {
			Value: "ftp://www.example.com",
			Error: true,
		},
		"HttpUrl": {
			Value: "http://www.example.com",
			Error: false,
		},
		"HttpsUrl": {
			Value: "https://www.example.com",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsURLWithHTTPorHTTPS(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsURLWithHTTPorHTTPS(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsURLWithHTTPorHTTPS(%s) did not error", tc.Value)
			}
		})
	}
}
