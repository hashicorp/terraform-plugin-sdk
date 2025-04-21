// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseResourceAddress(t *testing.T) {
	cases := map[string]struct {
		Input    string
		Expected *resourceAddress
		Output   string
		Err      bool
	}{
		"implicit primary managed instance, no specific index": {
			"aws_instance.foo",
			&resourceAddress{
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"implicit primary data instance, no specific index": {
			"data.aws_instance.foo",
			&resourceAddress{
				Mode:         DataResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"implicit primary, explicit index": {
			"aws_instance.foo[2]",
			&resourceAddress{
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        2,
			},
			"",
			false,
		},
		"implicit primary, explicit index over ten": {
			"aws_instance.foo[12]",
			&resourceAddress{
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        12,
			},
			"",
			false,
		},
		"explicit primary, explicit index": {
			"aws_instance.foo.primary[2]",
			&resourceAddress{
				Mode:            ManagedResourceMode,
				Type:            "aws_instance",
				Name:            "foo",
				InstanceType:    typePrimary,
				InstanceTypeSet: true,
				Index:           2,
			},
			"",
			false,
		},
		"tainted": {
			"aws_instance.foo.tainted",
			&resourceAddress{
				Mode:            ManagedResourceMode,
				Type:            "aws_instance",
				Name:            "foo",
				InstanceType:    typeTainted,
				InstanceTypeSet: true,
				Index:           -1,
			},
			"",
			false,
		},
		"deposed": {
			"aws_instance.foo.deposed",
			&resourceAddress{
				Mode:            ManagedResourceMode,
				Type:            "aws_instance",
				Name:            "foo",
				InstanceType:    typeDeposed,
				InstanceTypeSet: true,
				Index:           -1,
			},
			"",
			false,
		},
		"with a hyphen": {
			"aws_instance.foo-bar",
			&resourceAddress{
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo-bar",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"managed in a module": {
			"module.child.aws_instance.foo",
			&resourceAddress{
				Path:         []string{"child"},
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"data in a module": {
			"module.child.data.aws_instance.foo",
			&resourceAddress{
				Path:         []string{"child"},
				Mode:         DataResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"nested modules": {
			"module.a.module.b.module.forever.aws_instance.foo",
			&resourceAddress{
				Path:         []string{"a", "b", "forever"},
				Mode:         ManagedResourceMode,
				Type:         "aws_instance",
				Name:         "foo",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"just a module": {
			"module.a",
			&resourceAddress{
				Path:         []string{"a"},
				Type:         "",
				Name:         "",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"just a nested module": {
			"module.a.module.b",
			&resourceAddress{
				Path:         []string{"a", "b"},
				Type:         "",
				Name:         "",
				InstanceType: typePrimary,
				Index:        -1,
			},
			"",
			false,
		},
		"module missing resource type": {
			"module.name.foo",
			nil,
			"",
			true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			out, err := parseResourceAddress(tc.Input)
			if (err != nil) != tc.Err {
				t.Fatalf("%s: unexpected err: %#v", tn, err)
			}
			if tc.Err {
				return
			}

			if !reflect.DeepEqual(out, tc.Expected) {
				t.Fatalf("bad: %q\n\nexpected:\n%#v\n\ngot:\n%#v", tn, tc.Expected, out)
			}

			expected := tc.Input
			if tc.Output != "" {
				expected = tc.Output
			}
			if out.String() != expected {
				t.Fatalf("bad: %q\n\nexpected: %s\n\ngot: %s", tn, expected, out)
			}
		})
	}
}

func TestResourceAddressLess(t *testing.T) {
	tests := []struct {
		A    string
		B    string
		Want bool
	}{
		{
			"foo.bar",
			"module.baz.foo.bar",
			true,
		},
		{
			"module.baz.foo.bar",
			"zzz.bar", // would sort after "module" in lexicographical sort
			false,
		},
		{
			"module.baz.foo.bar",
			"module.baz.foo.bar",
			false,
		},
		{
			"module.baz.foo.bar",
			"module.boz.foo.bar",
			true,
		},
		{
			"module.boz.foo.bar",
			"module.baz.foo.bar",
			false,
		},
		{
			"a.b",
			"b.c",
			true,
		},
		{
			"a.b",
			"a.c",
			true,
		},
		{
			"c.b",
			"b.c",
			false,
		},
		{
			"a.b[9]",
			"a.b[10]",
			true,
		},
		{
			"b.b[9]",
			"a.b[10]",
			false,
		},
		{
			"a.b",
			"a.b.deposed",
			true,
		},
		{
			"a.b.tainted",
			"a.b.deposed",
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s < %s", test.A, test.B), func(t *testing.T) {
			addrA, err := parseResourceAddress(test.A)
			if err != nil {
				t.Fatal(err)
			}
			addrB, err := parseResourceAddress(test.B)
			if err != nil {
				t.Fatal(err)
			}
			got := addrA.Less(addrB)
			invGot := addrB.Less(addrA)
			if got != test.Want {
				t.Errorf(
					"wrong result\ntest: %s < %s\ngot:  %#v\nwant: %#v",
					test.A, test.B, got, test.Want,
				)
			}
			if test.A != test.B { // inverse test doesn't apply when equal
				if invGot != !test.Want {
					t.Errorf(
						"wrong inverse result\ntest: %s < %s\ngot:  %#v\nwant: %#v",
						test.B, test.A, invGot, !test.Want,
					)
				}
			} else {
				if invGot != test.Want {
					t.Errorf(
						"wrong inverse result\ntest: %s < %s\ngot:  %#v\nwant: %#v",
						test.B, test.A, invGot, test.Want,
					)
				}
			}
		})
	}
}
