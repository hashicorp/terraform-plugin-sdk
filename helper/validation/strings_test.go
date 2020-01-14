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
			Error: false,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: false,
		},
		"NewLine": {
			Value: "\n",
			Error: false,
		},
		"MultipleSymbols": {
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

func TestValidationStringIsNotWhitespace(t *testing.T) {
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
			_, errors := StringIsNotWhiteSpace(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsNotWhiteSpace(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsNotWhiteSpace(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsEmpty(t *testing.T) {
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
			Error: false,
		},
		"SingleSpace": {
			Value: " ",
			Error: true,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: true,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: true,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: true,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsEmpty(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsEmpty(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsEmpty(%s) did not error", tc.Value)
			}
		})
	}
}

func TestValidationStringIsWhiteSpace(t *testing.T) {
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
			Error: false,
		},
		"SingleSpace": {
			Value: " ",
			Error: false,
		},
		"MultipleSpaces": {
			Value: "     ",
			Error: false,
		},
		"MultipleWhitespace": {
			Value: "  \t\n\f   ",
			Error: false,
		},
		"Sentence": {
			Value: "Hello kt's sentence.",
			Error: true,
		},

		"StartsWithWhitespace": {
			Value: "  7",
			Error: true,
		},
		"EndsWithWhitespace": {
			Value: "7 ",
			Error: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := StringIsWhiteSpace(tc.Value, tn)

			if len(errors) > 0 && !tc.Error {
				t.Errorf("StringIsWhiteSpace(%s) produced an unexpected error", tc.Value)
			} else if len(errors) == 0 && tc.Error {
				t.Errorf("StringIsWhiteSpace(%s) did not error", tc.Value)
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
