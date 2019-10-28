package terraform

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/hcl2shim"
)

func TestResourceStateEqual(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *ResourceState
	}{
		// Different types
		{
			false,
			&ResourceState{Type: "foo"},
			&ResourceState{Type: "bar"},
		},

		// Different dependencies
		{
			false,
			&ResourceState{Dependencies: []string{"foo"}},
			&ResourceState{Dependencies: []string{"bar"}},
		},

		{
			false,
			&ResourceState{Dependencies: []string{"foo", "bar"}},
			&ResourceState{Dependencies: []string{"foo"}},
		},

		{
			true,
			&ResourceState{Dependencies: []string{"bar", "foo"}},
			&ResourceState{Dependencies: []string{"foo", "bar"}},
		},

		// Different primaries
		{
			false,
			&ResourceState{Primary: nil},
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
		},

		{
			true,
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
		},

		// Different tainted
		{
			false,
			&ResourceState{
				Primary: &InstanceState{
					ID: "foo",
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},

		{
			true,
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},
	}

	for i, tc := range cases {
		if tc.One.Equal(tc.Two) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
		if tc.Two.Equal(tc.One) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
	}
}

func TestResourceStateTaint(t *testing.T) {
	cases := map[string]struct {
		Input  *ResourceState
		Output *ResourceState
	}{
		"no primary": {
			&ResourceState{},
			&ResourceState{},
		},

		"primary, not tainted": {
			&ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},

		"primary, tainted": {
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},
	}

	for k, tc := range cases {
		tc.Input.Taint()
		if !reflect.DeepEqual(tc.Input, tc.Output) {
			t.Fatalf(
				"Failure: %s\n\nExpected: %#v\n\nGot: %#v",
				k, tc.Output, tc.Input)
		}
	}
}

func TestResourceStateUntaint(t *testing.T) {
	cases := map[string]struct {
		Input          *ResourceState
		ExpectedOutput *ResourceState
	}{
		"no primary, err": {
			Input:          &ResourceState{},
			ExpectedOutput: &ResourceState{},
		},

		"primary, not tainted": {
			Input: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
			ExpectedOutput: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
		},
		"primary, tainted": {
			Input: &ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			ExpectedOutput: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
		},
	}

	for k, tc := range cases {
		tc.Input.Untaint()
		if !reflect.DeepEqual(tc.Input, tc.ExpectedOutput) {
			t.Fatalf(
				"Failure: %s\n\nExpected: %#v\n\nGot: %#v",
				k, tc.ExpectedOutput, tc.Input)
		}
	}
}

func TestInstanceStateEmpty(t *testing.T) {
	cases := map[string]struct {
		In     *InstanceState
		Result bool
	}{
		"nil is empty": {
			nil,
			true,
		},
		"non-nil but without ID is empty": {
			&InstanceState{},
			true,
		},
		"with ID is not empty": {
			&InstanceState{
				ID: "i-abc123",
			},
			false,
		},
	}

	for tn, tc := range cases {
		if tc.In.Empty() != tc.Result {
			t.Fatalf("%q expected %#v to be empty: %#v", tn, tc.In, tc.Result)
		}
	}
}

func TestInstanceStateEqual(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *InstanceState
	}{
		// Nils
		{
			false,
			nil,
			&InstanceState{},
		},

		{
			false,
			&InstanceState{},
			nil,
		},

		// Different IDs
		{
			false,
			&InstanceState{ID: "foo"},
			&InstanceState{ID: "bar"},
		},

		// Different Attributes
		{
			false,
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
			&InstanceState{Attributes: map[string]string{"foo": "baz"}},
		},

		// Different Attribute keys
		{
			false,
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
			&InstanceState{Attributes: map[string]string{"bar": "baz"}},
		},

		{
			false,
			&InstanceState{Attributes: map[string]string{"bar": "baz"}},
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
		},
	}

	for i, tc := range cases {
		if tc.One.Equal(tc.Two) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
	}
}

func TestInstanceState_MergeDiff(t *testing.T) {
	is := InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"foo":  "bar",
			"port": "8000",
		},
	}

	diff := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				Old: "bar",
				New: "baz",
			},
			"bar": &ResourceAttrDiff{
				Old: "",
				New: "foo",
			},
			"baz": &ResourceAttrDiff{
				Old:         "",
				New:         "foo",
				NewComputed: true,
			},
			"port": &ResourceAttrDiff{
				NewRemoved: true,
			},
		},
	}

	is2 := is.MergeDiff(diff)

	expected := map[string]string{
		"foo": "baz",
		"bar": "foo",
		"baz": hcl2shim.UnknownVariableValue,
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

// GH-12183. This tests that a list with a computed set generates the
// right partial state. This never failed but is put here for completion
// of the test case for GH-12183.
func TestInstanceState_MergeDiff_computedSet(t *testing.T) {
	is := InstanceState{}

	diff := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"config.#": &ResourceAttrDiff{
				Old:         "0",
				New:         "1",
				RequiresNew: true,
			},

			"config.0.name": &ResourceAttrDiff{
				Old: "",
				New: "hello",
			},

			"config.0.rules.#": &ResourceAttrDiff{
				Old:         "",
				NewComputed: true,
			},
		},
	}

	is2 := is.MergeDiff(diff)

	expected := map[string]string{
		"config.#":         "1",
		"config.0.name":    "hello",
		"config.0.rules.#": hcl2shim.UnknownVariableValue,
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

func TestInstanceState_MergeDiff_nil(t *testing.T) {
	var is *InstanceState

	diff := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				Old: "",
				New: "baz",
			},
		},
	}

	is2 := is.MergeDiff(diff)

	expected := map[string]string{
		"foo": "baz",
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

func TestInstanceState_MergeDiff_nilDiff(t *testing.T) {
	is := InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"foo": "bar",
		},
	}

	is2 := is.MergeDiff(nil)

	expected := map[string]string{
		"foo": "bar",
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}
