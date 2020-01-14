package validation

import (
	"testing"
)

func TestGUID(t *testing.T) {
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
		"InvalidGuid": {
			Value: "00000000-0000-123-0000-000000000000",
			Error: true,
		},
		"ValidGuidWithOutDashs": {
			Value: "12345678123412341234123456789012",
			Error: true,
		},
		"ValidGuid": {
			Value: "00000000-0000-0000-0000-000000000000",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsBase64(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("%s: GUID(%s) produced an unexpected error", tn, errors)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("%s: GUID(%s) did not error", tn, errors)
			}
		})
	}
}

func TestGUIDorEmpty(t *testing.T) {
	cases := map[string]struct {
		Value interface{}
		Error bool
	}{
		"NotString": {
			Value: 7,
			Error: false,
		},
		"Empty": {
			Value: "",
			Error: true,
		},
		"InvalidGuid": {
			Value: "00000000-0000-123-0000-000000000000",
			Error: true,
		},
		"ValidGuidWithOutDashs": {
			Value: "12345678123412341234123456789012",
			Error: true,
		},
		"ValidGuid": {
			Value: "00000000-0000-0000-0000-000000000000",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsBase64(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("%s: GUIDOrEmpty(%s) produced an unexpected error", tn, errors)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("%s: GUIDOrEmpty(%s) did not error", tn, errors)
			}
		})
	}
}
