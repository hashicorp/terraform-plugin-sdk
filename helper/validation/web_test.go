package validation

import (
	"testing"
)

func TestURLIsHTTPS(t *testing.T) {
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
			_, errors := URLIsHTTPS(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("URLIsHTTPS(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("URLIsHTTPS(%s) did not error", tc.Value)
			}
		})
	}
}

func TestURLIsHTTPOrHTTPS(t *testing.T) {
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
			_, errors := URLIsHTTPOrHTTPS(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("URLIsHTTPOrHTTPS(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("URLIsHTTPOrHTTPS(%s) did not error", tc.Value)
			}
		})
	}
}
