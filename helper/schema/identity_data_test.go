// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestIdentityDataGet(t *testing.T) {
	cases := []struct {
		Name           string
		IdentitySchema map[string]*Schema
		State          *terraform.InstanceState
		Diff           *terraform.InstanceDiff
		Key            string
		Value          interface{}
	}{
		{
			Name: "no state, empty diff",
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

		{
			Name: "no state, diff with identity",
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

		// TODO: these numbers are off since i removed some cases -> remove them
		// #3
		{
			Name: "state with identity, no diff",
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

		{
			Name: "state with identity, empty diff",
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
			Value: "",
		},

		// #5
		{
			Name: "int type: state with identity, no diff",
			IdentitySchema: map[string]*Schema{
				"port": {
					Type:              TypeInt,
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"port": "80",
				},
			},

			Diff: nil,

			Key: "port",

			Value: 80,
		},

		{
			Name: "int list type: state with identity, empty diff",
			IdentitySchema: map[string]*Schema{
				"ports": {
					Type:              TypeList,
					Elem:              &Schema{Type: TypeInt},
					RequiredForImport: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "3",
					"ports.0": "1",
					"ports.1": "2",
					"ports.2": "5",
				},
			},

			Key: "ports.1",

			Value: 2,
		},

		{
			Name: "int list type length: state with identity, empty diff",
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

		{
			Name: "int list type length: empty state, empty diff",
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

		{
			Name: "int list type all: state with identity, empty diff",
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

		{
			Name: "full object: empty state, diff with identity",
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

		{
			Name: "float zero: empty state, empty diff", // todo: this should return nothing, right?
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

		{
			Name: "float: state with identity, empty diff",
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

		{
			Name: "float: state with identity, diff with identity",
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

	for i, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
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
				t.Fatalf("Bad: %d\n\n%#v\n\nExpected: %#v", i, v, tc.Value)
			}
		})
	}
}

// func TestIdentityDataGetOk(t *testing.T) {
// 	cases := []struct {
// 		IdentitySchema map[string]*Schema
// 		State          *terraform.InstanceState
// 		Diff           *terraform.InstanceDiff
// 		Key            string
// 		Value          interface{}
// 		Ok             bool
// 	}{
// 		/*
// 		 * Primitives
// 		 */
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: &terraform.InstanceDiff{
// 				Attributes: map[string]*terraform.ResourceAttrDiff{
// 					"region": {
// 						Old: "",
// 						New: "",
// 					},
// 				},
// 			},

// 			Key:   "region",
// 			Value: "",
// 			Ok:    false,
// 		},

// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: &terraform.InstanceDiff{
// 				Attributes: map[string]*terraform.ResourceAttrDiff{
// 					"region": {
// 						Old:         "",
// 						New:         "",
// 						NewComputed: true,
// 					},
// 				},
// 			},

// 			Key:   "region",
// 			Value: "",
// 			Ok:    false,
// 		},

// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: "",
// 			Ok:    false,
// 		},

// 		/*
// 		 * Lists
// 		 */

// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeList,

// 					Elem: &Schema{Type: TypeInt},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports",
// 			Value: []interface{}{},
// 			Ok:    false,
// 		},

// 		/*
// 		 * Map
// 		 */

// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeMap,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports",
// 			Value: map[string]interface{}{},
// 			Ok:    false,
// 		},

// 		/*
// 		 * Set
// 		 */

// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set:  func(a interface{}) int { return a.(int) },
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports",
// 			Value: []interface{}{},
// 			Ok:    false,
// 		},

// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set:  func(a interface{}) int { return a.(int) },
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports.0",
// 			Value: 0,
// 			Ok:    false,
// 		},

// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set:  func(a interface{}) int { return a.(int) },
// 				},
// 			},

// 			State: nil,

// 			Diff: &terraform.InstanceDiff{
// 				Attributes: map[string]*terraform.ResourceAttrDiff{
// 					"ports.#": {
// 						Old: "0",
// 						New: "0",
// 					},
// 				},
// 			},

// 			Key:   "ports",
// 			Value: []interface{}{},
// 			Ok:    false,
// 		},

// 		// Further illustrates and clarifies the GetOk semantics from #933, and
// 		// highlights the limitation that zero-value config is currently
// 		// indistinguishable from unset config.
// 		{
// 			Schema: map[string]*Schema{
// 				"from_port": {
// 					Type: TypeInt,
// 				},
// 			},

// 			State: nil,

// 			Diff: &terraform.InstanceDiff{
// 				Attributes: map[string]*terraform.ResourceAttrDiff{
// 					"from_port": {
// 						Old: "",
// 						New: "0",
// 					},
// 				},
// 			},

// 			Key:   "from_port",
// 			Value: 0,
// 			Ok:    false,
// 		},
// 	}

// 	for i, tc := range cases {
// 		schema := map[string]*Schema{}
// 		d, err := schemaMapWithIdentity{schema, tc.IdentitySchema}.Data(tc.State, tc.Diff)
// 		if err != nil {
// 			t.Fatalf("err: %s", err)
// 		}

// 		v, ok := d.GetOk(tc.Key)
// 		if s, ok := v.(*Set); ok {
// 			v = s.List()
// 		}

// 		if !reflect.DeepEqual(v, tc.Value) {
// 			t.Fatalf("Bad: %d\n\n%#v", i, v)
// 		}
// 		if ok != tc.Ok {
// 			t.Fatalf("%d: expected ok: %t, got: %t", i, tc.Ok, ok)
// 		}
// 	}
// }

// func TestIdentityDataSet(t *testing.T) {
// 	var testNilPtr *string

// 	cases := []struct {
// 		IdentitySchema map[string]*Schema
// 		State          *terraform.InstanceState
// 		Diff           *terraform.InstanceDiff
// 		Key            string
// 		Value          interface{}
// 		Err            bool
// 		GetKey         string
// 		GetValue       interface{}

// 		// GetPreProcess can be set to munge the return value before being
// 		// compared to GetValue
// 		GetPreProcess func(interface{}) interface{}
// 	}{
// 		// #0: Basic good
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: "foo",

// 			GetKey:   "region",
// 			GetValue: "foo",
// 		},

// 		// #1: Basic int
// 		{
// 			Schema: map[string]*Schema{
// 				"port": {
// 					Type: TypeInt,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "port",
// 			Value: 80,

// 			GetKey:   "port",
// 			GetValue: 80,
// 		},

// 		// #2: Basic bool
// 		{
// 			Schema: map[string]*Schema{
// 				"vpc": {
// 					Type: TypeBool,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "vpc",
// 			Value: true,

// 			GetKey:   "vpc",
// 			GetValue: true,
// 		},

// 		// #3
// 		{
// 			Schema: map[string]*Schema{
// 				"vpc": {
// 					Type: TypeBool,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "vpc",
// 			Value: false,

// 			GetKey:   "vpc",
// 			GetValue: false,
// 		},

// 		// #4: Invalid type
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: 80,
// 			Err:   true,

// 			GetKey:   "region",
// 			GetValue: "",
// 		},

// 		// #5: List of primitives, set list
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeList,

// 					Elem: &Schema{Type: TypeInt},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports",
// 			Value: []int{1, 2, 5},

// 			GetKey:   "ports",
// 			GetValue: []interface{}{1, 2, 5},
// 		},

// 		// #6: List of primitives, set list with error
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeList,

// 					Elem: &Schema{Type: TypeInt},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ports",
// 			Value: []interface{}{1, "NOPE", 5},
// 			Err:   true,

// 			GetKey:   "ports",
// 			GetValue: []interface{}{},
// 		},

// 		// #7: Set a list of maps
// 		{
// 			Schema: map[string]*Schema{
// 				"config_vars": {
// 					Type: TypeList,

// 					Elem: &Schema{
// 						Type: TypeMap,
// 					},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key: "config_vars",
// 			Value: []interface{}{
// 				map[string]interface{}{
// 					"foo": "bar",
// 				},
// 				map[string]interface{}{
// 					"bar": "baz",
// 				},
// 			},
// 			Err: false,

// 			GetKey: "config_vars",
// 			GetValue: []interface{}{
// 				map[string]interface{}{
// 					"foo": "bar",
// 				},
// 				map[string]interface{}{
// 					"bar": "baz",
// 				},
// 			},
// 		},

// 		// #8: Set, with list
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set: func(a interface{}) int {
// 						return a.(int)
// 					},
// 				},
// 			},

// 			State: &terraform.InstanceState{
// 				Attributes: map[string]string{
// 					"ports.#": "3",
// 					"ports.0": "100",
// 					"ports.1": "80",
// 					"ports.2": "80",
// 				},
// 			},

// 			Key:   "ports",
// 			Value: []interface{}{100, 125, 125},

// 			GetKey:   "ports",
// 			GetValue: []interface{}{100, 125},
// 		},

// 		// #9: Set, with Set
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set: func(a interface{}) int {
// 						return a.(int)
// 					},
// 				},
// 			},

// 			State: &terraform.InstanceState{
// 				Attributes: map[string]string{
// 					"ports.#":   "3",
// 					"ports.100": "100",
// 					"ports.80":  "80",
// 					"ports.81":  "81",
// 				},
// 			},

// 			Key: "ports",
// 			Value: &Set{
// 				m: map[string]interface{}{
// 					"1": 1,
// 					"2": 2,
// 				},
// 			},

// 			GetKey:   "ports",
// 			GetValue: []interface{}{1, 2},
// 		},

// 		// #10: Set single item
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeInt},
// 					Set: func(a interface{}) int {
// 						return a.(int)
// 					},
// 				},
// 			},

// 			State: &terraform.InstanceState{
// 				Attributes: map[string]string{
// 					"ports.#":   "2",
// 					"ports.100": "100",
// 					"ports.80":  "80",
// 				},
// 			},

// 			Key:   "ports.100",
// 			Value: 256,
// 			Err:   true,

// 			GetKey:   "ports",
// 			GetValue: []interface{}{100, 80},
// 		},

// 		// #11: Set with nested set
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeSet,
// 					Elem: &Resource{
// 						Schema: map[string]*Schema{
// 							"port": {
// 								Type: TypeInt,
// 							},

// 							"set": {
// 								Type: TypeSet,
// 								Elem: &Schema{Type: TypeInt},
// 								Set: func(a interface{}) int {
// 									return a.(int)
// 								},
// 							},
// 						},
// 					},
// 					Set: func(a interface{}) int {
// 						return a.(map[string]interface{})["port"].(int)
// 					},
// 				},
// 			},

// 			State: nil,

// 			Key: "ports",
// 			Value: []interface{}{
// 				map[string]interface{}{
// 					"port": 80,
// 				},
// 			},

// 			GetKey: "ports",
// 			GetValue: []interface{}{
// 				map[string]interface{}{
// 					"port": 80,
// 					"set":  []interface{}{},
// 				},
// 			},

// 			GetPreProcess: func(v interface{}) interface{} {
// 				if v == nil {
// 					return v
// 				}
// 				s, ok := v.([]interface{})
// 				if !ok {
// 					return v
// 				}
// 				for _, v := range s {
// 					m, ok := v.(map[string]interface{})
// 					if !ok {
// 						continue
// 					}
// 					if m["set"] == nil {
// 						continue
// 					}
// 					if s, ok := m["set"].(*Set); ok {
// 						m["set"] = s.List()
// 					}
// 				}

// 				return v
// 			},
// 		},

// 		// #12: List of floats, set list
// 		{
// 			Schema: map[string]*Schema{
// 				"ratios": {
// 					Type: TypeList,

// 					Elem: &Schema{Type: TypeFloat},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ratios",
// 			Value: []float64{1.0, 2.2, 5.5},

// 			GetKey:   "ratios",
// 			GetValue: []interface{}{1.0, 2.2, 5.5},
// 		},

// 		// #12: Set of floats, set list
// 		{
// 			Schema: map[string]*Schema{
// 				"ratios": {
// 					Type: TypeSet,

// 					Elem: &Schema{Type: TypeFloat},
// 					Set: func(a interface{}) int {
// 						return int(math.Float64bits(a.(float64)))
// 					},
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "ratios",
// 			Value: []float64{1.0, 2.2, 5.5},

// 			GetKey:   "ratios",
// 			GetValue: []interface{}{1.0, 2.2, 5.5},
// 		},

// 		// #13: Basic pointer
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: testPtrTo("foo"),

// 			GetKey:   "region",
// 			GetValue: "foo",
// 		},

// 		// #14: Basic nil value
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: testPtrTo(nil),

// 			GetKey:   "region",
// 			GetValue: "",
// 		},

// 		// #15: Basic nil pointer
// 		{
// 			IdentitySchema: map[string]*Schema{
// 				"region": {
// 					Type:              TypeString,
// 					RequiredForImport: true,
// 				},
// 			},

// 			State: nil,

// 			Diff: nil,

// 			Key:   "region",
// 			Value: testNilPtr,

// 			GetKey:   "region",
// 			GetValue: "",
// 		},

// 		// #16: Set in a list
// 		{
// 			Schema: map[string]*Schema{
// 				"ports": {
// 					Type: TypeList,
// 					Elem: &Resource{
// 						Schema: map[string]*Schema{
// 							"set": {
// 								Type: TypeSet,
// 								Elem: &Schema{Type: TypeInt},
// 								Set: func(a interface{}) int {
// 									return a.(int)
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},

// 			State: nil,

// 			Key: "ports",
// 			Value: []interface{}{
// 				map[string]interface{}{
// 					"set": []interface{}{
// 						1,
// 					},
// 				},
// 			},

// 			GetKey: "ports",
// 			GetValue: []interface{}{
// 				map[string]interface{}{
// 					"set": []interface{}{
// 						1,
// 					},
// 				},
// 			},
// 			GetPreProcess: func(v interface{}) interface{} {
// 				if v == nil {
// 					return v
// 				}
// 				s, ok := v.([]interface{})
// 				if !ok {
// 					return v
// 				}
// 				for _, v := range s {
// 					m, ok := v.(map[string]interface{})
// 					if !ok {
// 						continue
// 					}
// 					if m["set"] == nil {
// 						continue
// 					}
// 					if s, ok := m["set"].(*Set); ok {
// 						m["set"] = s.List()
// 					}
// 				}

// 				return v
// 			},
// 		},
// 	}

// 	for i, tc := range cases {
// 		schema := map[string]*Schema{}

// 		d, err := schemaMapWithIdentity{schema, tc.IdentitySchema}.Data(tc.State, tc.Diff)
// 		if err != nil {
// 			t.Fatalf("err: %s", err)
// 		}

// 		err = d.Set(tc.Key, tc.Value)
// 		if err != nil != tc.Err {
// 			t.Fatalf("%d err: %s", i, err)
// 		}

// 		v := d.Get(tc.GetKey)
// 		if s, ok := v.(*Set); ok {
// 			v = s.List()
// 		}

// 		if tc.GetPreProcess != nil {
// 			v = tc.GetPreProcess(v)
// 		}

// 		if !reflect.DeepEqual(v, tc.GetValue) {
// 			t.Fatalf("Get Bad: %d\n\n%#v", i, v)
// 		}
// 	}
// }
