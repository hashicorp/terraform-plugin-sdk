package validation

import (
	"testing"
)

func TestValidationStringIsNotEmpty(t *testing.T) {
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
		"SingleSpace": {
			Value: " ",
			Error: true,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: true,
		},
		"CarriageReturn": {
			Value: "\r",
			Error: true,
		},
		"NewLine": {
			Value: "\n",
			Error: true,
		},
		"Tab": {
			Value: "\t",
			Error: true,
		},
		"FormFeed": {
			Value: "\f",
			Error: true,
		},
		"VerticalTab": {
			Value: "\v",
			Error: true,
		},
		"SingleChar": {
			Value: "\v",
			Error: true,
		},
		"MultipleChars": {
			Value: "-_-",
			Error: false,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: false,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: false,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsNotEmpty(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsNotEmpty(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsNotEmpty(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsBase64(t *testing.T) {
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
		"NotBase64": {
			Value: "Do'h!",
			Error: true,
		},
		"Base64": {
			Value: "RG8naCE=",
			Error: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsBase64(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsBase64(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsBase64(%s) did not error", tc.Value)
			}
		})
	}
}
