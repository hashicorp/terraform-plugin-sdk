// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

import (
	"errors"
	"strings"
	"testing"
)

func TestDiagnosticsErr(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var diags Diagnostics
		err := diags.Err()
		if err != nil {
			t.Errorf("got non-nil error %#v; want nil", err)
		}
	})
	t.Run("warning only", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, SimpleWarning("bad"))
		err := diags.Err()
		if err != nil {
			t.Errorf("got non-nil error %#v; want nil", err)
		}
	})
	t.Run("one error", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		err := diags.Err()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		if got, want := err.Error(), "didn't work"; got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("two errors", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, FromError(errors.New("didn't work either")))
		err := diags.Err()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("error and warning", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, SimpleWarning("didn't work either"))
		err := diags.Err()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		// Since this "as error" mode is just a fallback for
		// non-diagnostics-aware situations like tests, we don't actually
		// distinguish warnings and errors here since the point is to just
		// get the messages rendered. User-facing code should be printing
		// each diagnostic separately, so won't enter this codepath,
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
}

func TestDiagnosticsErrWithWarnings(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var diags Diagnostics
		err := diags.ErrWithWarnings()
		if err != nil {
			t.Errorf("got non-nil error %#v; want nil", err)
		}
	})
	t.Run("warning only", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, SimpleWarning("bad"))
		err := diags.ErrWithWarnings()
		if err == nil {
			t.Errorf("got nil error; want NonFatalError")
			return
		}
		if _, ok := err.(NonFatalError); !ok {
			t.Errorf("got %T; want NonFatalError", err)
		}
	})
	t.Run("one error", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		err := diags.ErrWithWarnings()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		if got, want := err.Error(), "didn't work"; got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("two errors", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, FromError(errors.New("didn't work either")))
		err := diags.ErrWithWarnings()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
	t.Run("error and warning", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, SimpleWarning("didn't work either"))
		err := diags.ErrWithWarnings()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		// Since this "as error" mode is just a fallback for
		// non-diagnostics-aware situations like tests, we don't actually
		// distinguish warnings and errors here since the point is to just
		// get the messages rendered. User-facing code should be printing
		// each diagnostic separately, so won't enter this codepath,
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
	})
}

func TestDiagnosticsNonFatalErr(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var diags Diagnostics
		err := diags.NonFatalErr()
		if err != nil {
			t.Errorf("got non-nil error %#v; want nil", err)
		}
	})
	t.Run("warning only", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, SimpleWarning("bad"))
		err := diags.NonFatalErr()
		if err == nil {
			t.Errorf("got nil error; want NonFatalError")
			return
		}
		if _, ok := err.(NonFatalError); !ok {
			t.Errorf("got %T; want NonFatalError", err)
		}
	})
	t.Run("one error", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		err := diags.NonFatalErr()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		if got, want := err.Error(), "didn't work"; got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
		if _, ok := err.(NonFatalError); !ok {
			t.Errorf("got %T; want NonFatalError", err)
		}
	})
	t.Run("two errors", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, FromError(errors.New("didn't work either")))
		err := diags.NonFatalErr()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
		if _, ok := err.(NonFatalError); !ok {
			t.Errorf("got %T; want NonFatalError", err)
		}
	})
	t.Run("error and warning", func(t *testing.T) {
		var diags Diagnostics
		diags = append(diags, FromError(errors.New("didn't work")))
		diags = append(diags, SimpleWarning("didn't work either"))
		err := diags.NonFatalErr()
		if err == nil {
			t.Fatalf("got nil error %#v; want non-nil", err)
		}
		// Since this "as error" mode is just a fallback for
		// non-diagnostics-aware situations like tests, we don't actually
		// distinguish warnings and errors here since the point is to just
		// get the messages rendered. User-facing code should be printing
		// each diagnostic separately, so won't enter this codepath,
		want := strings.TrimSpace(`
2 problems:

- didn't work
- didn't work either
`)
		if got := err.Error(); got != want {
			t.Errorf("wrong error message\ngot:  %s\nwant: %s", got, want)
		}
		if _, ok := err.(NonFatalError); !ok {
			t.Errorf("got %T; want NonFatalError", err)
		}
	})
}
