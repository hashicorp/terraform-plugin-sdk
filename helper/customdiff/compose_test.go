// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customdiff

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAll(t *testing.T) {
	var aCalled, bCalled, cCalled bool

	provider := testProvider(
		map[string]*schema.Schema{},
		All(
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				aCalled = true
				return errors.New("A bad")
			},
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				bCalled = true
				return nil
			},
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				cCalled = true
				return errors.New("C bad")
			},
		),
	)

	_, err := testDiff(
		provider,
		map[string]string{
			"foo": "bar",
		},
		map[string]string{
			"foo": "baz",
		},
	)

	if err == nil {
		t.Fatal("Diff succeeded; want error")
	}
	if s, sub := err.Error(), "* A bad"; !strings.Contains(s, sub) {
		t.Errorf("Missing substring %q in error message %q", sub, s)
	}
	if s, sub := err.Error(), "* C bad"; !strings.Contains(s, sub) {
		t.Errorf("Missing substring %q in error message %q", sub, s)
	}

	if !aCalled {
		t.Error("customize callback A was not called")
	}
	if !bCalled {
		t.Error("customize callback B was not called")
	}
	if !cCalled {
		t.Error("customize callback C was not called")
	}
}

func TestSequence(t *testing.T) {
	var aCalled, bCalled, cCalled bool

	provider := testProvider(
		map[string]*schema.Schema{},
		Sequence(
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				aCalled = true
				return nil
			},
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				bCalled = true
				return errors.New("B bad")
			},
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				cCalled = true
				return errors.New("C bad")
			},
		),
	)

	_, err := testDiff(
		provider,
		map[string]string{
			"foo": "bar",
		},
		map[string]string{
			"foo": "baz",
		},
	)

	if err == nil {
		t.Fatal("Diff succeeded; want error")
	}
	if got, want := err.Error(), "B bad"; got != want {
		t.Errorf("Wrong error message %q; want %q", got, want)
	}

	if !aCalled {
		t.Error("customize callback A was not called")
	}
	if !bCalled {
		t.Error("customize callback B was not called")
	}
	if cCalled {
		t.Error("customize callback C was called (should not have been)")
	}
}
