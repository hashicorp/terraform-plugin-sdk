package terraform

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
)

func TestDiffEmpty(t *testing.T) {
	var d *diff
	if !d.empty() {
		t.Fatal("should be empty")
	}

	d = new(diff)
	if !d.empty() {
		t.Fatal("should be empty")
	}

	mod := d.addModule(addrs.RootModuleInstance)
	mod.Resources["nodeA"] = &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				Old: "foo",
				New: "bar",
			},
		},
	}

	if d.empty() {
		t.Fatal("should not be empty")
	}
}

func TestDiffEmpty_taintedIsNotEmpty(t *testing.T) {
	diff := new(diff)

	mod := diff.addModule(addrs.RootModuleInstance)
	mod.Resources["nodeA"] = &InstanceDiff{
		DestroyTainted: true,
	}

	if diff.empty() {
		t.Fatal("should not be empty, since DestroyTainted was set")
	}
}

func TestDiffEqual(t *testing.T) {
	cases := map[string]struct {
		D1, D2 *diff
		Equal  bool
	}{
		"nil": {
			nil,
			new(diff),
			false,
		},

		"empty": {
			new(diff),
			new(diff),
			true,
		},

		"different module order": {
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}},
					&moduleDiff{Path: []string{"root", "bar"}},
				},
			},
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "bar"}},
					&moduleDiff{Path: []string{"root", "foo"}},
				},
			},
			true,
		},

		"different module diff destroys": {
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}, Destroy: true},
				},
			},
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}, Destroy: false},
				},
			},
			true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := tc.D1.equal(tc.D2)
			if actual != tc.Equal {
				t.Fatalf("expected: %v\n\n%#v\n\n%#v", tc.Equal, tc.D1, tc.D2)
			}
		})
	}
}

func TestDiffPrune(t *testing.T) {
	cases := map[string]struct {
		D1, D2 *diff
	}{
		"nil": {
			nil,
			nil,
		},

		"empty": {
			new(diff),
			new(diff),
		},

		"empty module": {
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}},
				},
			},
			&diff{},
		},

		"destroy module": {
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}, Destroy: true},
				},
			},
			&diff{
				Modules: []*moduleDiff{
					&moduleDiff{Path: []string{"root", "foo"}, Destroy: true},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.D1.prune()
			if !tc.D1.equal(tc.D2) {
				t.Fatalf("bad:\n\n%#v\n\n%#v", tc.D1, tc.D2)
			}
		})
	}
}

func TestModuleDiff_ChangeType(t *testing.T) {
	cases := []struct {
		diff   *moduleDiff
		Result diffChangeType
	}{
		{
			&moduleDiff{},
			diffNone,
		},
		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo": &InstanceDiff{Destroy: true},
				},
			},
			diffDestroy,
		},
		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"foo": &ResourceAttrDiff{
								Old: "",
								New: "bar",
							},
						},
					},
				},
			},
			diffUpdate,
		},
		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"foo": &ResourceAttrDiff{
								Old:         "",
								New:         "bar",
								RequiresNew: true,
							},
						},
					},
				},
			},
			diffCreate,
		},
		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo": &InstanceDiff{
						Destroy: true,
						Attributes: map[string]*ResourceAttrDiff{
							"foo": &ResourceAttrDiff{
								Old:         "",
								New:         "bar",
								RequiresNew: true,
							},
						},
					},
				},
			},
			diffUpdate,
		},
	}

	for i, tc := range cases {
		actual := tc.diff.changeType()
		if actual != tc.Result {
			t.Fatalf("%d: %#v", i, actual)
		}
	}
}

func TestDiff_DeepCopy(t *testing.T) {
	cases := map[string]*diff{
		"empty": &diff{},

		"basic diff": &diff{
			Modules: []*moduleDiff{
				&moduleDiff{
					Path: []string{"root"},
					Resources: map[string]*InstanceDiff{
						"aws_instance.foo": &InstanceDiff{
							Attributes: map[string]*ResourceAttrDiff{
								"num": &ResourceAttrDiff{
									Old: "0",
									New: "2",
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dup := tc.deepCopy()
			if !reflect.DeepEqual(dup, tc) {
				t.Fatalf("\n%#v\n\n%#v", dup, tc)
			}
		})
	}
}

func TestModuleDiff_Empty(t *testing.T) {
	diff := new(moduleDiff)
	if !diff.empty() {
		t.Fatal("should be empty")
	}

	diff.Resources = map[string]*InstanceDiff{
		"nodeA": &InstanceDiff{},
	}

	if !diff.empty() {
		t.Fatal("should be empty")
	}

	diff.Resources["nodeA"].Attributes = map[string]*ResourceAttrDiff{
		"foo": &ResourceAttrDiff{
			Old: "foo",
			New: "bar",
		},
	}

	if diff.empty() {
		t.Fatal("should not be empty")
	}

	diff.Resources["nodeA"].Attributes = nil
	diff.Resources["nodeA"].Destroy = true

	if diff.empty() {
		t.Fatal("should not be empty")
	}
}

func TestModuleDiff_String(t *testing.T) {
	diff := &moduleDiff{
		Resources: map[string]*InstanceDiff{
			"nodeA": &InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "foo",
						New: "bar",
					},
					"bar": &ResourceAttrDiff{
						Old:         "foo",
						NewComputed: true,
					},
					"longfoo": &ResourceAttrDiff{
						Old:         "foo",
						New:         "bar",
						RequiresNew: true,
					},
					"secretfoo": &ResourceAttrDiff{
						Old:       "foo",
						New:       "bar",
						Sensitive: true,
					},
				},
			},
		},
	}

	actual := strings.TrimSpace(diff.String())
	expected := strings.TrimSpace(moduleDiffStrBasic)
	if actual != expected {
		t.Fatalf("bad:\n%s", actual)
	}
}

func TestInstanceDiff_ChangeType(t *testing.T) {
	cases := []struct {
		diff   *InstanceDiff
		Result diffChangeType
	}{
		{
			&InstanceDiff{},
			diffNone,
		},
		{
			&InstanceDiff{Destroy: true},
			diffDestroy,
		},
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "",
						New: "bar",
					},
				},
			},
			diffUpdate,
		},
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},
			diffCreate,
		},
		{
			&InstanceDiff{
				Destroy: true,
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},
			diffDestroyCreate,
		},
		{
			&InstanceDiff{
				DestroyTainted: true,
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},
			diffDestroyCreate,
		},
	}

	for i, tc := range cases {
		actual := tc.diff.ChangeType()
		if actual != tc.Result {
			t.Fatalf("%d: %#v", i, actual)
		}
	}
}

func TestInstanceDiff_Empty(t *testing.T) {
	var rd *InstanceDiff

	if !rd.Empty() {
		t.Fatal("should be empty")
	}

	rd = new(InstanceDiff)

	if !rd.Empty() {
		t.Fatal("should be empty")
	}

	rd = &InstanceDiff{Destroy: true}

	if rd.Empty() {
		t.Fatal("should not be empty")
	}

	rd = &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				New: "bar",
			},
		},
	}

	if rd.Empty() {
		t.Fatal("should not be empty")
	}
}

func TestModuleDiff_Instances(t *testing.T) {
	yesDiff := &InstanceDiff{Destroy: true}
	noDiff := &InstanceDiff{Destroy: true, DestroyTainted: true}

	cases := []struct {
		diff   *moduleDiff
		Id     string
		Result []*InstanceDiff
	}{
		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo": yesDiff,
					"bar": noDiff,
				},
			},
			"foo",
			[]*InstanceDiff{
				yesDiff,
			},
		},

		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo":   yesDiff,
					"foo.0": yesDiff,
					"bar":   noDiff,
				},
			},
			"foo",
			[]*InstanceDiff{
				yesDiff,
				yesDiff,
			},
		},

		{
			&moduleDiff{
				Resources: map[string]*InstanceDiff{
					"foo":     yesDiff,
					"foo.0":   yesDiff,
					"foo_bar": noDiff,
					"bar":     noDiff,
				},
			},
			"foo",
			[]*InstanceDiff{
				yesDiff,
				yesDiff,
			},
		},
	}

	for i, tc := range cases {
		actual := tc.diff.instances(tc.Id)
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d: %#v", i, actual)
		}
	}
}

func TestInstanceDiff_RequiresNew(t *testing.T) {
	rd := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{},
		},
	}

	if rd.RequiresNew() {
		t.Fatal("should not require new")
	}

	rd.Attributes["foo"].RequiresNew = true

	if !rd.RequiresNew() {
		t.Fatal("should require new")
	}
}

func TestInstanceDiff_RequiresNew_nil(t *testing.T) {
	var rd *InstanceDiff

	if rd.RequiresNew() {
		t.Fatal("should not require new")
	}
}

func TestInstanceDiffSame(t *testing.T) {
	cases := []struct {
		One, Two *InstanceDiff
		Same     bool
		Reason   string
	}{
		{
			&InstanceDiff{},
			&InstanceDiff{},
			true,
			"",
		},

		{
			nil,
			nil,
			true,
			"",
		},

		{
			&InstanceDiff{Destroy: false},
			&InstanceDiff{Destroy: true},
			false,
			"diff: Destroy; old: false, new: true",
		},

		{
			&InstanceDiff{Destroy: true},
			&InstanceDiff{Destroy: true},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{},
				},
			},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"bar": &ResourceAttrDiff{},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{},
				},
			},
			false,
			"attribute mismatch: bar",
		},

		// Extra attributes
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{},
					"bar": &ResourceAttrDiff{},
				},
			},
			false,
			"extra attributes: bar",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{RequiresNew: true},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{RequiresNew: false},
				},
			},
			false,
			"diff RequiresNew; old: true, new: false",
		},

		// NewComputed on primitive
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old:         "",
						New:         "${var.foo}",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
				},
			},
			true,
			"",
		},

		// NewComputed on primitive, removed
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old:         "",
						New:         "${var.foo}",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{},
			},
			true,
			"",
		},

		// NewComputed on set, removed
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "",
						New:         "",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.1": &ResourceAttrDiff{
						Old:        "foo",
						New:        "",
						NewRemoved: true,
					},
					"foo.2": &ResourceAttrDiff{
						Old: "",
						New: "bar",
					},
				},
			},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{NewComputed: true},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.0": &ResourceAttrDiff{
						Old: "",
						New: "12",
					},
				},
			},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~35964334.bar": &ResourceAttrDiff{
						Old: "",
						New: "${var.foo}",
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.87654323.bar": &ResourceAttrDiff{
						Old: "",
						New: "12",
					},
				},
			},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "0",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{},
			},
			true,
			"",
		},

		// Computed can change RequiresNew by removal, and that's okay
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "0",
						NewComputed: true,
						RequiresNew: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{},
			},
			true,
			"",
		},

		// Computed can change Destroy by removal, and that's okay
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "0",
						NewComputed: true,
						RequiresNew: true,
					},
				},

				Destroy: true,
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{},
			},
			true,
			"",
		},

		// Computed can change Destroy by elements
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "0",
						NewComputed: true,
						RequiresNew: true,
					},
				},

				Destroy: true,
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "1",
						New: "1",
					},
					"foo.12": &ResourceAttrDiff{
						Old:         "4",
						New:         "12",
						RequiresNew: true,
					},
				},

				Destroy: true,
			},
			true,
			"",
		},

		// Computed sets may not contain all fields in the original diff, and
		// because multiple entries for the same set can compute to the same
		// hash before the values are computed or interpolated, the overall
		// count can change as well.
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~35964334.bar": &ResourceAttrDiff{
						Old: "",
						New: "${var.foo}",
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "2",
					},
					"foo.87654323.bar": &ResourceAttrDiff{
						Old: "",
						New: "12",
					},
					"foo.87654325.bar": &ResourceAttrDiff{
						Old: "",
						New: "12",
					},
					"foo.87654325.baz": &ResourceAttrDiff{
						Old: "",
						New: "12",
					},
				},
			},
			true,
			"",
		},

		// Computed values in maps will fail the "Same" check as well
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.%": &ResourceAttrDiff{
						Old:         "",
						New:         "",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.%": &ResourceAttrDiff{
						Old:         "0",
						New:         "1",
						NewComputed: false,
					},
					"foo.val": &ResourceAttrDiff{
						Old: "",
						New: "something",
					},
				},
			},
			true,
			"",
		},

		// In a DESTROY/CREATE scenario, the plan diff will be run against the
		// state of the old instance, while the apply diff will be run against an
		// empty state (because the state is cleared when the destroy runs.)
		// For complex attributes, this can result in keys that seem to disappear
		// between the two diffs, when in reality everything is working just fine.
		//
		// Same() needs to take into account this scenario by analyzing NewRemoved
		// and treating as "Same" a diff that does indeed have that key removed.
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"somemap.oldkey": &ResourceAttrDiff{
						Old:        "long ago",
						New:        "",
						NewRemoved: true,
					},
					"somemap.newkey": &ResourceAttrDiff{
						Old: "",
						New: "brave new world",
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"somemap.newkey": &ResourceAttrDiff{
						Old: "",
						New: "brave new world",
					},
				},
			},
			true,
			"",
		},

		// Another thing that can occur in DESTROY/CREATE scenarios is that list
		// values that are going to zero have diffs that show up at plan time but
		// are gone at apply time. The NewRemoved handling catches the fields and
		// treats them as OK, but it also needs to treat the .# field itself as
		// okay to be present in the old diff but not in the new one.
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"reqnew": &ResourceAttrDiff{
						Old:         "old",
						New:         "new",
						RequiresNew: true,
					},
					"somemap.#": &ResourceAttrDiff{
						Old: "1",
						New: "0",
					},
					"somemap.oldkey": &ResourceAttrDiff{
						Old:        "long ago",
						New:        "",
						NewRemoved: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"reqnew": &ResourceAttrDiff{
						Old:         "",
						New:         "new",
						RequiresNew: true,
					},
				},
			},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"reqnew": &ResourceAttrDiff{
						Old:         "old",
						New:         "new",
						RequiresNew: true,
					},
					"somemap.%": &ResourceAttrDiff{
						Old: "1",
						New: "0",
					},
					"somemap.oldkey": &ResourceAttrDiff{
						Old:        "long ago",
						New:        "",
						NewRemoved: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"reqnew": &ResourceAttrDiff{
						Old:         "",
						New:         "new",
						RequiresNew: true,
					},
				},
			},
			true,
			"",
		},

		// Innner computed set should allow outer change in key
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~1.outer_val": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
					"foo.~1.inner.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~1.inner.~2.value": &ResourceAttrDiff{
						Old:         "",
						New:         "${var.bar}",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.12.outer_val": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
					"foo.12.inner.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.12.inner.42.value": &ResourceAttrDiff{
						Old: "",
						New: "baz",
					},
				},
			},
			true,
			"",
		},

		// Innner computed list should allow outer change in key
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~1.outer_val": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
					"foo.~1.inner.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.~1.inner.0.value": &ResourceAttrDiff{
						Old:         "",
						New:         "${var.bar}",
						NewComputed: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.12.outer_val": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
					"foo.12.inner.#": &ResourceAttrDiff{
						Old: "0",
						New: "1",
					},
					"foo.12.inner.0.value": &ResourceAttrDiff{
						Old: "",
						New: "baz",
					},
				},
			},
			true,
			"",
		},

		// When removing all collection items, the diff is allowed to contain
		// nothing when re-creating the resource. This should be the "Same"
		// since we said we were going from 1 to 0.
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.%": &ResourceAttrDiff{
						Old:         "1",
						New:         "0",
						RequiresNew: true,
					},
					"foo.bar": &ResourceAttrDiff{
						Old:         "baz",
						New:         "",
						NewRemoved:  true,
						RequiresNew: true,
					},
				},
			},
			&InstanceDiff{},
			true,
			"",
		},

		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo.#": &ResourceAttrDiff{
						Old:         "1",
						New:         "0",
						RequiresNew: true,
					},
					"foo.0": &ResourceAttrDiff{
						Old:         "baz",
						New:         "",
						NewRemoved:  true,
						RequiresNew: true,
					},
				},
			},
			&InstanceDiff{},
			true,
			"",
		},

		// Make sure that DestroyTainted diffs pass as well, especially when diff
		// two works off of no state.
		{
			&InstanceDiff{
				DestroyTainted: true,
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "foo",
						New: "foo",
					},
				},
			},
			&InstanceDiff{
				DestroyTainted: true,
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
				},
			},
			true,
			"",
		},
		// RequiresNew in different attribute
		{
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "foo",
						New: "foo",
					},
					"bar": &ResourceAttrDiff{
						Old:         "bar",
						New:         "baz",
						RequiresNew: true,
					},
				},
			},
			&InstanceDiff{
				Attributes: map[string]*ResourceAttrDiff{
					"foo": &ResourceAttrDiff{
						Old: "",
						New: "foo",
					},
					"bar": &ResourceAttrDiff{
						Old:         "",
						New:         "baz",
						RequiresNew: true,
					},
				},
			},
			true,
			"",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			same, reason := tc.One.Same(tc.Two)
			if same != tc.Same {
				t.Fatalf("%d: expected same: %t, got %t (%s)\n\n one: %#v\n\ntwo: %#v",
					i, tc.Same, same, reason, tc.One, tc.Two)
			}
			if reason != tc.Reason {
				t.Fatalf(
					"%d: bad reason\n\nexpected: %#v\n\ngot: %#v", i, tc.Reason, reason)
			}
		})
	}
}

const moduleDiffStrBasic = `
CREATE: nodeA
  bar:       "foo" => "<computed>"
  foo:       "foo" => "bar"
  longfoo:   "foo" => "bar" (forces new resource)
  secretfoo: "<sensitive>" => "<sensitive>" (attribute changed)
`

func TestCountFlatmapContainerValues(t *testing.T) {
	for i, tc := range []struct {
		attrs map[string]string
		key   string
		count string
	}{
		{
			attrs: map[string]string{"set.2.list.#": "9999", "set.2.list.0": "x", "set.2.list.0.z": "y", "set.2.attr": "bar", "set.#": "9999"},
			key:   "set.2.list.#",
			count: "1",
		},
		{
			attrs: map[string]string{"set.2.list.#": "9999", "set.2.list.0": "x", "set.2.list.0.z": "y", "set.2.attr": "bar", "set.#": "9999"},
			key:   "set.#",
			count: "1",
		},
		{
			attrs: map[string]string{"set.2.list.0": "x", "set.2.list.0.z": "y", "set.2.attr": "bar", "set.#": "9999"},
			key:   "set.#",
			count: "1",
		},
		{
			attrs: map[string]string{"map.#": "3", "map.a": "b", "map.a.#": "0", "map.b": "4"},
			key:   "map.#",
			count: "2",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			count := countFlatmapContainerValues(tc.key, tc.attrs)
			if count != tc.count {
				t.Fatalf("expected %q, got %q", tc.count, count)
			}
		})
	}
}
