package validation

import (
	"regexp"
	"testing"
)

func TestValidateCIDR(t *testing.T) {
	cases := map[string]struct {
		Value string
		Error bool
	}{
		"NotString": {
			Value: "777",
			Error: true,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"Zeros": {
			Value: "0.0.0.0",
			Error: true,
		},
		"Slash8": {
			Value: "127.0.0.1/8",
			Error: false,
		},
		"Slash33": {
			Value: "127.0.0.1/33",
			Error: true,
		},
		"Slash-1": {
			Value: "127.0.0.1/-1",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsCIDR(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsCIDR(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsCIDR(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidateIsIPv6Address(t *testing.T) {
	cases := map[string]struct {
		Value string
		Error bool
	}{
		"NotString": {
			Value: "777",
			Error: true,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"ZeroIpv4": {
			Value: "0.0.0.0",
			Error: false,
		},
		"NotARealAddress": {
			Value: "not:a:real:address:1:2:3:4",
			Error: true,
		},
		"Text": {
			Value: "text",
			Error: true,
		},
		"Colons": {
			Value: "::",
			Error: false,
		},
		"ZeroIPv6": {
			Value: "0:0:0:0:0:0:0:0",
			Error: false,
		},
		"Valid1": {
			Value: "2001:0db8:85a3:0:0:8a2e:0370:7334",
			Error: false,
		},
		"Valid2": {
			Value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsIPv6Address(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsIPv6Address(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsIPv6Address(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidateIsIPv4Address(t *testing.T) {
	cases := map[string]struct {
		Value string
		Error bool
	}{
		"NotString": {
			Value: "777",
			Error: true,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"Zeros": {
			Value: "0.0.0.0",
			Error: false,
		},
		"Chars": {
			Value: "1.2.3.no",
			Error: true,
		},
		"Text": {
			Value: "text",
			Error: true,
		},
		"Valid": {
			Value: "1.2.3.4",
			Error: false,
		},
		"Valid10s": {
			Value: "12.34.43.21",
			Error: false,
		},
		"Valid100s": {
			Value: "100.123.199.0",
			Error: false,
		},
		"Valid255": {
			Value: "255.255.255.255",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsIPv4Address(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsIPv4Address(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsIPv4Address(%s) did not error", tc.Value)
			}
		})
	}
}

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

func TestValidationMACAddress(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotString": {
			Value: "777",
			Error: true,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"Text": {
			Value: "text d",
			Error: true,
		},
		"Gibberish": {
			Value: "12:34:no",
			Error: true,
		},
		"InvalidOctetSize": {
			Value: "123:34:56:78:90:ab",
			Error: true,
		},
		"InvalidOctetChars": {
			Value: "12:34:56:78:90:NO",
			Error: true,
		},
		"ValidLowercase": {
			Value: "12:34:56:78:90:ab",
			Error: false,
		},
		"ValidUppercase": {
			Value: "ab:cd:ef:AB:CD:EF",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsMACAddress(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsMACAddress(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsMACAddress(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationPortNumber(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotInt": {
			Value: "kt",
			Error: true,
		},
		"Negative": {
			Value: -1,
			Error: true,
		},
		"Zero": {
			Value: 0,
			Error: true,
		},
		"One": {
			Value: 1,
			Error: false,
		},
		"Valid": {
			Value: 8477,
			Error: false,
		},
		"MaxPort": {
			Value: 65535,
			Error: false,
		},
		"OneToHigh": {
			Value: 65536,
			Error: true,
		},
		"HugeNumber": {
			Value: 7000000,
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsPortNumber(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsPortNumber(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsPortNumber(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationPortNumberOrZero(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotInt": {
			Value: "kt",
			Error: true,
		},
		"Negative": {
			Value: -1,
			Error: true,
		},
		"Zero": {
			Value: 0,
			Error: false,
		},
		"One": {
			Value: 1,
			Error: false,
		},
		"Valid": {
			Value: 8477,
			Error: false,
		},
		"MaxPort": {
			Value: 65535,
			Error: false,
		},
		"OneToHigh": {
			Value: 65536,
			Error: true,
		},
		"HugeNumber": {
			Value: 7000000,
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := IsPortNumberOrZero(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("IsPortNumberOrZero(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("IsPortNumberOrZero(%s) did not error", tc.Value)
			}
		})
	}
}
