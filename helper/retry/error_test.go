// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"errors"
	"testing"
)

func TestUnexpectedStateErrorError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *UnexpectedStateError
		want string
	}{
		{
			name: "with last error",
			err: &UnexpectedStateError{
				State:         "failed",
				ExpectedState: []string{"running", "ready"},
				LastError:     errors.New("compute quota exceeded"),
			},
			want: "unexpected state 'failed', wanted target 'running, ready'. last error: compute quota exceeded",
		},
		{
			name: "nil last error",
			err: &UnexpectedStateError{
				State:         "failed",
				ExpectedState: []string{"running"},
			},
			want: "unexpected state 'failed', wanted target 'running'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
