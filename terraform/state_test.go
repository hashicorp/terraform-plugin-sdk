// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/hcl2shim"
)

func TestStateValidate(t *testing.T) {
	cases := map[string]struct {
		In  *State
		Err bool
	}{
		"empty state": {
			&State{},
			false,
		},

		"multiple modules": {
			&State{
				Modules: []*ModuleState{
					{
						Path: []string{"root", "foo"},
					},
					{
						Path: []string{"root", "foo"},
					},
				},
			},
			true,
		},
	}

	for name, tc := range cases {
		// Init the state
		tc.In.init()

		err := tc.In.Validate()
		if (err != nil) != tc.Err {
			t.Fatalf("%s: err: %s", name, err)
		}
	}
}

func TestStateAddModule(t *testing.T) {
	cases := []struct {
		In  []addrs.ModuleInstance
		Out [][]string
	}{
		{
			[]addrs.ModuleInstance{
				addrs.RootModuleInstance,
				addrs.RootModuleInstance.Child("child", addrs.NoKey),
			},
			[][]string{
				{"root"},
				{"root", "child"},
			},
		},

		{
			[]addrs.ModuleInstance{
				addrs.RootModuleInstance.Child("foo", addrs.NoKey).Child("bar", addrs.NoKey),
				addrs.RootModuleInstance.Child("foo", addrs.NoKey),
				addrs.RootModuleInstance,
				addrs.RootModuleInstance.Child("bar", addrs.NoKey),
			},
			[][]string{
				{"root"},
				{"root", "bar"},
				{"root", "foo"},
				{"root", "foo", "bar"},
			},
		},
		// Same last element, different middle element
		{
			[]addrs.ModuleInstance{
				addrs.RootModuleInstance.Child("foo", addrs.NoKey).Child("bar", addrs.NoKey), // This one should sort after...
				addrs.RootModuleInstance.Child("foo", addrs.NoKey),
				addrs.RootModuleInstance,
				addrs.RootModuleInstance.Child("bar", addrs.NoKey).Child("bar", addrs.NoKey), // ...this one.
				addrs.RootModuleInstance.Child("bar", addrs.NoKey),
			},
			[][]string{
				{"root"},
				{"root", "bar"},
				{"root", "foo"},
				{"root", "bar", "bar"},
				{"root", "foo", "bar"},
			},
		},
	}

	for _, tc := range cases {
		s := new(State)
		for _, p := range tc.In {
			s.AddModule(p)
		}

		actual := make([][]string, 0, len(tc.In))
		for _, m := range s.Modules {
			actual = append(actual, m.Path)
		}

		if diff := cmp.Diff(tc.Out, actual); diff != "" {
			t.Fatalf("unexpected difference: %s", diff)
		}
	}
}

func TestStateDeepCopy(t *testing.T) {
	cases := []struct {
		State *State
	}{
		// Nil
		{nil},

		// Version
		{
			&State{Version: 5},
		},
		// TFVersion
		{
			&State{TFVersion: "5"},
		},
		// Modules
		{
			&State{
				Version: 6,
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{},
								},
							},
						},
					},
				},
			},
		},
		// Deposed
		// The nil values shouldn't be there if the State was properly init'ed,
		// but the Copy should still work anyway.
		{
			&State{
				Version: 6,
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{},
								},
								Deposed: []*InstanceState{
									{ID: "test"},
									nil,
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("copy-%d", i), func(t *testing.T) {
			actual := tc.State.DeepCopy()
			expected := tc.State
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("Expected: %#v\nRecevied: %#v\n", expected, actual)
			}
		})
	}
}

func TestStateEqual(t *testing.T) {
	cases := []struct {
		Name     string
		Result   bool
		One, Two *State
	}{
		// Nils
		{
			"one nil",
			false,
			nil,
			&State{Version: 2},
		},

		{
			"both nil",
			true,
			nil,
			nil,
		},

		// Different versions
		{
			"different state versions",
			false,
			&State{Version: 5},
			&State{Version: 2},
		},

		// Different modules
		{
			"different module states",
			false,
			&State{
				Modules: []*ModuleState{
					{
						Path: []string{"root"},
					},
				},
			},
			&State{},
		},

		{
			"same module states",
			true,
			&State{
				Modules: []*ModuleState{
					{
						Path: []string{"root"},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: []string{"root"},
					},
				},
			},
		},

		// Meta differs
		{
			"differing meta values with primitives",
			false,
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"schema_version": "1",
									},
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"schema_version": "2",
									},
								},
							},
						},
					},
				},
			},
		},

		// Meta with complex types
		{
			"same meta with complex types",
			true,
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"timeouts": map[string]interface{}{
											"create": 42,
											"read":   "27",
										},
									},
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"timeouts": map[string]interface{}{
											"create": 42,
											"read":   "27",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// Meta with complex types that have been altered during serialization
		{
			"same meta with complex types that have been json-ified",
			true,
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"timeouts": map[string]interface{}{
											"create": int(42),
											"read":   "27",
										},
									},
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Primary: &InstanceState{
									Meta: map[string]interface{}{
										"timeouts": map[string]interface{}{
											"create": float64(42),
											"read":   "27",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			if tc.One.Equal(tc.Two) != tc.Result {
				t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
			}
			if tc.Two.Equal(tc.One) != tc.Result {
				t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
			}
		})
	}
}

func TestStateCompareAges(t *testing.T) {
	cases := []struct {
		Result   StateAgeComparison
		Err      bool
		One, Two *State
	}{
		{
			StateAgeEqual, false,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "1",
				Serial:  2,
			},
		},
		{
			StateAgeReceiverOlder, false,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "1",
				Serial:  3,
			},
		},
		{
			StateAgeReceiverNewer, false,
			&State{
				Lineage: "1",
				Serial:  3,
			},
			&State{
				Lineage: "1",
				Serial:  2,
			},
		},
		{
			StateAgeEqual, true,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "2",
				Serial:  2,
			},
		},
		{
			StateAgeEqual, true,
			&State{
				Lineage: "1",
				Serial:  3,
			},
			&State{
				Lineage: "2",
				Serial:  2,
			},
		},
	}

	for i, tc := range cases {
		result, err := tc.One.CompareAges(tc.Two)

		if err != nil && !tc.Err {
			t.Errorf(
				"%d: got error, but want success\n\n%s\n\n%s",
				i, tc.One, tc.Two,
			)
			continue
		}

		if err == nil && tc.Err {
			t.Errorf(
				"%d: got success, but want error\n\n%s\n\n%s",
				i, tc.One, tc.Two,
			)
			continue
		}

		if result != tc.Result {
			t.Errorf(
				"%d: got result %d, but want %d\n\n%s\n\n%s",
				i, result, tc.Result, tc.One, tc.Two,
			)
			continue
		}
	}
}

func TestStateSameLineage(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *State
	}{
		{
			true,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "1",
			},
		},
		{
			// Empty lineage is compatible with all
			true,
			&State{
				Lineage: "",
			},
			&State{
				Lineage: "1",
			},
		},
		{
			// Empty lineage is compatible with all
			true,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "",
			},
		},
		{
			false,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "2",
			},
		},
	}

	for i, tc := range cases {
		result := tc.One.SameLineage(tc.Two)

		if result != tc.Result {
			t.Errorf(
				"%d: got %v, but want %v\n\n%s\n\n%s",
				i, result, tc.Result, tc.One, tc.Two,
			)
			continue
		}
	}
}

func TestStateRemove(t *testing.T) {
	cases := map[string]struct {
		Address  string
		One, Two *State
	}{
		"simple resource": {
			"test_instance.foo",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.bar": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"single instance": {
			"test_instance.foo.primary",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path:      rootModulePath,
						Resources: map[string]*ResourceState{},
					},
				},
			},
		},

		"single instance in multi-count": {
			"test_instance.foo[0]",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.0": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.foo.1": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.1": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"single resource, multi-count": {
			"test_instance.foo",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.0": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.foo.1": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path:      rootModulePath,
						Resources: map[string]*ResourceState{},
					},
				},
			},
		},

		"full module": {
			"module.foo",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"module and children": {
			"module.foo",
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					{
						Path: []string{"root", "foo", "bar"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": {
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},
	}

	for k, tc := range cases {
		if err := tc.One.Remove(tc.Address); err != nil {
			t.Fatalf("bad: %s\n\n%s", k, err)
		}

		if !tc.One.Equal(tc.Two) {
			t.Fatalf("Bad: %s\n\n%s\n\n%s", k, tc.One.String(), tc.Two.String())
		}
	}
}

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

func TestStateEmpty(t *testing.T) {
	cases := []struct {
		In     *State
		Result bool
	}{
		{
			nil,
			true,
		},
		{
			&State{},
			true,
		},
		{
			&State{
				Remote: &RemoteState{Type: "foo"},
			},
			true,
		},
		{
			&State{
				Modules: []*ModuleState{
					{},
				},
			},
			false,
		},
	}

	for i, tc := range cases {
		if tc.In.Empty() != tc.Result {
			t.Fatalf("bad %d %#v:\n\n%#v", i, tc.Result, tc.In)
		}
	}
}

func TestStateHasResources(t *testing.T) {
	cases := []struct {
		In     *State
		Result bool
	}{
		{
			nil,
			false,
		},
		{
			&State{},
			false,
		},
		{
			&State{
				Remote: &RemoteState{Type: "foo"},
			},
			false,
		},
		{
			&State{
				Modules: []*ModuleState{
					{},
				},
			},
			false,
		},
		{
			&State{
				Modules: []*ModuleState{
					{},
					{},
				},
			},
			false,
		},
		{
			&State{
				Modules: []*ModuleState{
					{},
					{
						Resources: map[string]*ResourceState{
							"foo.foo": {},
						},
					},
				},
			},
			true,
		},
	}

	for i, tc := range cases {
		if tc.In.HasResources() != tc.Result {
			t.Fatalf("bad %d %#v:\n\n%#v", i, tc.Result, tc.In)
		}
	}
}

func TestStateIsRemote(t *testing.T) {
	cases := []struct {
		In     *State
		Result bool
	}{
		{
			nil,
			false,
		},
		{
			&State{},
			false,
		},
		{
			&State{
				Remote: &RemoteState{Type: "foo"},
			},
			true,
		},
	}

	for i, tc := range cases {
		if tc.In.IsRemote() != tc.Result {
			t.Fatalf("bad %d %#v:\n\n%#v", i, tc.Result, tc.In)
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
			"foo": {
				Old: "bar",
				New: "baz",
			},
			"bar": {
				Old: "",
				New: "foo",
			},
			"baz": {
				Old:         "",
				New:         "foo",
				NewComputed: true,
			},
			"port": {
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
			"config.#": {
				Old:         "0",
				New:         "1",
				RequiresNew: true,
			},

			"config.0.name": {
				Old: "",
				New: "hello",
			},

			"config.0.rules.#": {
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
			"foo": {
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

func TestParseResourceStateKey(t *testing.T) {
	cases := []struct {
		Input       string
		Expected    *ResourceStateKey
		ExpectedErr bool
	}{
		{
			Input: "aws_instance.foo.3",
			Expected: &ResourceStateKey{
				Mode:  ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: 3,
			},
		},
		{
			Input: "aws_instance.foo.0",
			Expected: &ResourceStateKey{
				Mode:  ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: 0,
			},
		},
		{
			Input: "aws_instance.foo",
			Expected: &ResourceStateKey{
				Mode:  ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: -1,
			},
		},
		{
			Input: "data.aws_ami.foo",
			Expected: &ResourceStateKey{
				Mode:  DataResourceMode,
				Type:  "aws_ami",
				Name:  "foo",
				Index: -1,
			},
		},
		{
			Input:       "aws_instance.foo.malformed",
			ExpectedErr: true,
		},
		{
			Input:       "aws_instance.foo.malformedwithnumber.123",
			ExpectedErr: true,
		},
		{
			Input:       "malformed",
			ExpectedErr: true,
		},
	}
	for _, tc := range cases {
		rsk, err := parseResourceStateKey(tc.Input)
		if rsk != nil && tc.Expected != nil && !rsk.Equal(tc.Expected) {
			t.Fatalf("%s: expected %s, got %s", tc.Input, tc.Expected, rsk)
		}
		if (err != nil) != tc.ExpectedErr {
			t.Fatalf("%s: expected err: %t, got %s", tc.Input, tc.ExpectedErr, err)
		}
	}
}

func TestResourceNameSort(t *testing.T) {
	names := []string{
		"a",
		"b",
		"a.0",
		"a.c",
		"a.d",
		"c",
		"a.b.0",
		"a.b.1",
		"a.b.10",
		"a.b.2",
	}

	sort.Sort(resourceNameSort(names))

	expected := []string{
		"a",
		"a.0",
		"a.b.0",
		"a.b.1",
		"a.b.2",
		"a.b.10",
		"a.c",
		"a.d",
		"b",
		"c",
	}

	if !reflect.DeepEqual(names, expected) {
		t.Fatalf("got: %q\nexpected: %q\n", names, expected)
	}
}
