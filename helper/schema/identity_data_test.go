// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestIdentityDataGet(t *testing.T) {
	cases := map[string]struct {
		IdentitySchema map[string]*Schema
		State          *terraform.InstanceState
		Diff           *terraform.InstanceDiff
		Key            string
		Value          interface{}
	}{
		"no state, empty diff": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: &terraform.InstanceDiff{},

			Key:   "region",
			Value: "",
		},

		"no state, diff with identity": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: &terraform.InstanceDiff{
				Identity: map[string]string{
					"region": "foo",
				},
			},

			Key:   "region",
			Value: "foo",
		},

		"state with identity, no diff": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"region": "bar",
				},
			},

			Diff: nil,

			Key: "region",

			Value: "bar",
		},

		"state with identity, empty diff": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"region": "foo",
				},
			},

			Diff: &terraform.InstanceDiff{},

			Key:   "region",
			Value: "foo", // This is different than for resource data – which would be empty
		},

		"int type: state with identity, no diff": {
			IdentitySchema: map[string]*Schema{
				"port": {
					Type:              TypeInt,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"port": "80",
				},
			},

			Diff: nil,

			Key: "port",

			Value: 80,
		},

		"int list type: state with identity, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type:              TypeList,
					Elem:              &Schema{Type: TypeInt},
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"ports.#": "3",
					"ports.0": "1",
					"ports.1": "2",
					"ports.2": "5",
				},
			},

			Key: "ports.1",

			Value: 2,
		},

		"int list type length: state with identity, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type: TypeList,
					Elem: &Schema{Type: TypeInt},
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"ports.#": "3",
					"ports.0": "1",
					"ports.1": "2",
					"ports.2": "5",
				},
			},

			Key: "ports.#",

			Value: 3,
		},

		"int list type length: empty state, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type:              TypeList,
					Elem:              &Schema{Type: TypeInt},
					RequiredForImport: true,
				},
			},

			State: nil,

			Key: "ports.#",

			Value: 0,
		},

		"int list type all: state with identity, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type:              TypeList,
					Elem:              &Schema{Type: TypeInt},
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"ports.#": "3",
					"ports.0": "1",
					"ports.1": "2",
					"ports.2": "5",
				},
			},

			Key: "ports",

			Value: []interface{}{1, 2, 5},
		},

		"full object: empty state, diff with identity": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: &terraform.InstanceDiff{
				Identity: map[string]string{
					"region": "foo",
				},
			},

			Key: "",

			Value: map[string]interface{}{
				"region": "foo",
			},
		},

		"float zero: empty state, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ratio": {
					Type:              TypeFloat,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key: "ratio",

			Value: 0.0,
		},

		"float: state with identity, empty diff": {
			IdentitySchema: map[string]*Schema{
				"ratio": {
					Type:              TypeFloat,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"ratio": "0.5",
				},
			},

			Diff: nil,

			Key: "ratio",

			Value: 0.5,
		},

		"float: state with identity, diff with identity": {
			IdentitySchema: map[string]*Schema{
				"ratio": {
					Type:              TypeFloat,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Identity: map[string]string{
					"ratio": "-0.5",
				},
			},

			Diff: &terraform.InstanceDiff{
				Identity: map[string]string{
					"ratio": "33.0",
				},
			},

			Key: "ratio",

			Value: 33.0,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			schema := map[string]*Schema{}
			d, err := schemaMapWithIdentity{schema, tc.IdentitySchema}.Data(tc.State, tc.Diff)
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			v := identity.Get(tc.Key)

			if !reflect.DeepEqual(v, tc.Value) {
				t.Fatalf("Bad: %s\n\n%#v\n\nExpected: %#v", name, v, tc.Value)
			}
		})
	}
}

func TestIdentityDataGetOk(t *testing.T) {
	cases := []struct {
		IdentitySchema map[string]*Schema
		State          *terraform.InstanceState
		Diff           *terraform.InstanceDiff
		Key            string
		Value          interface{}
		Ok             bool
	}{
		/*
		 * Primitives
		 */
		{
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: &terraform.InstanceDiff{
				Identity: map[string]string{
					"region": "",
				},
			},

			Key:   "region",
			Value: "",
			Ok:    false,
		},

		{
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: "",
			Ok:    false,
		},

		/*
		 * Lists
		 */

		{
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type:              TypeList,
					Elem:              &Schema{Type: TypeInt},
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "ports",
			Value: []interface{}{},
			Ok:    false,
		},
	}

	for i, tc := range cases {
		schema := map[string]*Schema{}
		d, err := schemaMapWithIdentity{schema, tc.IdentitySchema}.Data(tc.State, tc.Diff)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		identity, err := d.Identity()
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v, ok := identity.GetOk(tc.Key)
		if s, ok := v.(*Set); ok {
			v = s.List()
		}

		if !reflect.DeepEqual(v, tc.Value) {
			t.Fatalf("Bad: %d\n\n%#v", i, v)
		}
		if ok != tc.Ok {
			t.Fatalf("%d: expected ok: %t, got: %t", i, tc.Ok, ok)
		}
	}
}

func TestIdentityDataSet(t *testing.T) {
	var testNilPtr *string

	cases := map[string]struct {
		IdentitySchema map[string]*Schema
		State          *terraform.InstanceState
		Diff           *terraform.InstanceDiff
		Key            string
		Value          interface{}
		Err            bool
		GetKey         string
		GetValue       interface{}
	}{
		"basic string": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: "foo",

			GetKey:   "region",
			GetValue: "foo",
		},

		"basic int": {
			IdentitySchema: map[string]*Schema{
				"port": {
					Type: TypeInt,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "port",
			Value: 80,

			GetKey:   "port",
			GetValue: 80,
		},

		"basic bool": {
			IdentitySchema: map[string]*Schema{
				"vpc": {
					Type: TypeBool,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "vpc",
			Value: true,

			GetKey:   "vpc",
			GetValue: true,
		},

		"basic bool false": {
			IdentitySchema: map[string]*Schema{
				"vpc": {
					Type: TypeBool,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "vpc",
			Value: false,

			GetKey:   "vpc",
			GetValue: false,
		},

		"invalid type": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: 80,
			Err:   true,

			GetKey:   "region",
			GetValue: "",
		},

		"list of primitives - set list": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type: TypeList,

					Elem: &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Diff: nil,

			Key:   "ports",
			Value: []int{1, 2, 5},

			GetKey:   "ports",
			GetValue: []interface{}{1, 2, 5},
		},

		"list of primitives - set list with error": {
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type: TypeList,

					Elem: &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Diff: nil,

			Key:   "ports",
			Value: []interface{}{1, "NOPE", 5},
			Err:   true,

			GetKey:   "ports",
			GetValue: []interface{}{},
		},

		"list of floats - set list": {
			IdentitySchema: map[string]*Schema{
				"ratios": {
					Type: TypeList,

					Elem: &Schema{Type: TypeFloat},
				},
			},

			State: nil,

			Diff: nil,

			Key:   "ratios",
			Value: []float64{1.0, 2.2, 5.5},

			GetKey:   "ratios",
			GetValue: []interface{}{1.0, 2.2, 5.5},
		},

		"basic pointer": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: testPtrTo("foo"),

			GetKey:   "region",
			GetValue: "foo",
		},

		"basic nil value": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: testPtrTo(nil),

			GetKey:   "region",
			GetValue: "",
		},

		"basic nil pointer": {
			IdentitySchema: map[string]*Schema{
				"region": {
					Type:              TypeString,
					RequiredForImport: true,
				},
			},

			State: nil,

			Diff: nil,

			Key:   "region",
			Value: testNilPtr,

			GetKey:   "region",
			GetValue: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			schema := map[string]*Schema{}

			d, err := schemaMapWithIdentity{schema, tc.IdentitySchema}.Data(tc.State, tc.Diff)
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			err = identity.Set(tc.Key, tc.Value)
			if err != nil != tc.Err {
				t.Fatalf("%s err: %s", name, err)
			}

			// we retrieve a new identity to ensure memoization is working
			identity, err = d.Identity()
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			v := identity.Get(tc.GetKey)

			if !reflect.DeepEqual(v, tc.GetValue) {
				t.Fatalf("Get Bad: %s\n\n%#v", name, v)
			}
		})
	}
}
