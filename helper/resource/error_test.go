package resource

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "not found error",
			Err:      &NotFoundError{LastError: errors.New("test")},
			Expected: true,
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name: "wrapped not found error",
			Err:  fmt.Errorf("test: %w", &NotFoundError{LastError: errors.New("test")}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := IsNotFoundError(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
