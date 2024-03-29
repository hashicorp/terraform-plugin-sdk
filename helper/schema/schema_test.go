// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/hcl2shim"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/diagutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestEnvDefaultFunc(t *testing.T) {
	key := "TF_TEST_ENV_DEFAULT_FUNC"
	defer os.Unsetenv(key)

	f := EnvDefaultFunc(key, "42")
	t.Setenv(key, "foo")

	actual, err := f()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != "foo" {
		t.Fatalf("bad: %#v", actual)
	}

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual, err = f()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != "42" {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestMultiEnvDefaultFunc(t *testing.T) {
	keys := []string{
		"TF_TEST_MULTI_ENV_DEFAULT_FUNC1",
		"TF_TEST_MULTI_ENV_DEFAULT_FUNC2",
	}
	defer func() {
		for _, k := range keys {
			os.Unsetenv(k)
		}
	}()

	// Test that the first key is returned first
	f := MultiEnvDefaultFunc(keys, "42")
	t.Setenv(keys[0], "foo")

	actual, err := f()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != "foo" {
		t.Fatalf("bad: %#v", actual)
	}

	if err := os.Unsetenv(keys[0]); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test that the second key is returned if the first one is empty
	f = MultiEnvDefaultFunc(keys, "42")
	t.Setenv(keys[1], "foo")

	actual, err = f()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != "foo" {
		t.Fatalf("bad: %#v", actual)
	}

	if err := os.Unsetenv(keys[1]); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test that the default value is returned when no keys are set
	actual, err = f()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != "42" {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestValueType_Zero(t *testing.T) {
	cases := []struct {
		Type  ValueType
		Value interface{}
	}{
		{TypeBool, false},
		{TypeInt, 0},
		{TypeFloat, 0.0},
		{TypeString, ""},
		{TypeList, []interface{}{}},
		{TypeMap, map[string]interface{}{}},
		{TypeSet, new(Set)},
	}

	for i, tc := range cases {
		actual := tc.Type.Zero()
		if !reflect.DeepEqual(actual, tc.Value) {
			t.Fatalf("%d: %#v != %#v", i, actual, tc.Value)
		}
	}
}

func TestSchemaMap_Diff(t *testing.T) {
	cases := []struct {
		Name          string
		Schema        map[string]*Schema
		State         *terraform.InstanceState
		Config        map[string]interface{}
		CustomizeDiff CustomizeDiffFunc
		Diff          *terraform.InstanceDiff
		Err           bool
	}{
		{
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         "foo",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						NewComputed: true,
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: &terraform.InstanceState{
				ID: "foo",
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Computed, but set in config",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "foo",
				},
			},

			Config: map[string]interface{}{
				"availability_zone": "bar",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "foo",
						New: "bar",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Default",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Default:  "foo",
				},
			},

			State: nil,

			Config: nil,

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "",
						New: "foo",
					},
				},
			},

			Err: false,
		},

		{
			Name: "DefaultFunc, value",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "foo", nil
					},
				},
			},

			State: nil,

			Config: nil,

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "",
						New: "foo",
					},
				},
			},

			Err: false,
		},

		{
			Name: "DefaultFunc, configuration set",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "foo", nil
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "bar",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "",
						New: "bar",
					},
				},
			},

			Err: false,
		},

		{
			Name: "String with StateFunc",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					StateFunc: func(a interface{}) string {
						return a.(string) + "!"
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:      "",
						New:      "foo!",
						NewExtra: "foo",
					},
				},
			},

			Err: false,
		},

		{
			Name: "StateFunc not called with nil value",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					StateFunc: func(a interface{}) string {
						t.Fatalf("should not get here!")
						return ""
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Variable computed",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": hcl2shim.UnknownVariableValue,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         hcl2shim.UnknownVariableValue,
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Int decode",
			Schema: map[string]*Schema{
				"port": {
					Type:     TypeInt,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"port": 27,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"port": {
						Old:         "",
						New:         "27",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "bool decode",
			Schema: map[string]*Schema{
				"port": {
					Type:     TypeBool,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"port": false,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"port": {
						Old:         "",
						New:         "false",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Bool",
			Schema: map[string]*Schema{
				"delete": {
					Type:     TypeBool,
					Optional: true,
					Default:  false,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"delete": "false",
				},
			},

			Config: nil,

			Diff: nil,

			Err: false,
		},

		{
			Name: "List decode",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, 2, 5},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "0",
						New: "3",
					},
					"ports.0": {
						Old: "",
						New: "1",
					},
					"ports.1": {
						Old: "",
						New: "2",
					},
					"ports.2": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, 2, 5},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "0",
						New: "3",
					},
					"ports.0": {
						Old: "",
						New: "1",
					},
					"ports.1": {
						Old: "",
						New: "2",
					},
					"ports.2": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, hcl2shim.UnknownVariableValue, 5},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "0",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
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

			Config: map[string]interface{}{
				"ports": []interface{}{1, 2, 5},
			},

			Diff: nil,

			Err: false,
		},

		{
			Name: "",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "2",
					"ports.0": "1",
					"ports.1": "2",
				},
			},

			Config: map[string]interface{}{
				"ports": []interface{}{1, 2, 5},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "2",
						New: "3",
					},
					"ports.2": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Name: "",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, 2, 5},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "0",
						New:         "3",
						RequiresNew: true,
					},
					"ports.0": {
						Old:         "",
						New:         "1",
						RequiresNew: true,
					},
					"ports.1": {
						Old:         "",
						New:         "2",
						RequiresNew: true,
					},
					"ports.2": {
						Old:         "",
						New:         "5",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeList,
					Optional: true,
					Computed: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "List with computed set",
			Schema: map[string]*Schema{
				"config": {
					Type:     TypeList,
					Optional: true,
					ForceNew: true,
					MinItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"name": {
								Type:     TypeString,
								Required: true,
							},

							"rules": {
								Type:     TypeSet,
								Computed: true,
								Elem:     &Schema{Type: TypeString},
								Set:      HashString,
							},
						},
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{
						"name": "hello",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
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
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{5, 2, 1},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "0",
						New: "3",
					},
					"ports.1": {
						Old: "",
						New: "1",
					},
					"ports.2": {
						Old: "",
						New: "2",
					},
					"ports.5": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Computed: true,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "0",
				},
			},

			Config: nil,

			Diff: nil,

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: nil,

			Config: nil,

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{"2", "5", 1},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "0",
						New: "3",
					},
					"ports.1": {
						Old: "",
						New: "1",
					},
					"ports.2": {
						Old: "",
						New: "2",
					},
					"ports.5": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, hcl2shim.UnknownVariableValue, "5"},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "2",
					"ports.1": "1",
					"ports.2": "2",
				},
			},

			Config: map[string]interface{}{
				"ports": []interface{}{5, 2, 1},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "2",
						New: "3",
					},
					"ports.1": {
						Old: "1",
						New: "1",
					},
					"ports.2": {
						Old: "2",
						New: "2",
					},
					"ports.5": {
						Old: "",
						New: "5",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "2",
					"ports.1": "1",
					"ports.2": "2",
				},
			},

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "2",
						New: "0",
					},
					"ports.1": {
						Old:        "1",
						New:        "0",
						NewRemoved: true,
					},
					"ports.2": {
						Old:        "2",
						New:        "0",
						NewRemoved: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "bar",
					"ports.#":           "1",
					"ports.80":          "80",
				},
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeSet,
					Required: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem:     &Schema{Type: TypeInt},
							},
						},
					},
					Set: func(v interface{}) int {
						m := v.(map[string]interface{})
						ps := m["ports"].([]interface{})
						result := 0
						for _, p := range ps {
							result += p.(int)
						}
						return result
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ingress.#":           "2",
					"ingress.80.ports.#":  "1",
					"ingress.80.ports.0":  "80",
					"ingress.443.ports.#": "1",
					"ingress.443.ports.0": "443",
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"ports": []interface{}{443},
					},
					map[string]interface{}{
						"ports": []interface{}{80},
					},
				},
			},

			Diff: nil,

			Err: false,
		},

		{
			Name: "List of structure decode",
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Required: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"from": 8080,
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ingress.#": {
						Old: "0",
						New: "1",
					},
					"ingress.0.from": {
						Old: "",
						New: "8080",
					},
				},
			},

			Err: false,
		},

		{
			Name: "ComputedWhen",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:         TypeString,
					Computed:     true,
					ComputedWhen: []string{"port"},
				},

				"port": {
					Type:     TypeInt,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "foo",
					"port":              "80",
				},
			},

			Config: map[string]interface{}{
				"port": 80,
			},

			Diff: nil,

			Err: false,
		},

		{
			Name: "",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:         TypeString,
					Computed:     true,
					ComputedWhen: []string{"port"},
				},

				"port": {
					Type:     TypeInt,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"port": "80",
				},
			},

			Config: map[string]interface{}{
				"port": 80,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		/* TODO
		{
			Schema: map[string]*Schema{
				"availability_zone": &Schema{
					Type:         TypeString,
					Computed:     true,
					ComputedWhen: []string{"port"},
				},

				"port": &Schema{
					Type:     TypeInt,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "foo",
					"port":              "80",
				},
			},

			Config: map[string]interface{}{
				"port": 8080,
			},

			Diff: &terraform.ResourceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": &terraform.ResourceAttrDiff{
						Old:         "foo",
						NewComputed: true,
					},
					"port": &terraform.ResourceAttrDiff{
						Old: "80",
						New: "8080",
					},
				},
			},

			Err: false,
		},
		*/

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"config_vars": {
					Type: TypeMap,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"config_vars": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"config_vars.%": {
						Old: "0",
						New: "1",
					},

					"config_vars.bar": {
						Old: "",
						New: "baz",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"config_vars": {
					Type: TypeMap,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"config_vars.foo": "bar",
				},
			},

			Config: map[string]interface{}{
				"config_vars": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"config_vars.foo": {
						Old:        "bar",
						NewRemoved: true,
					},
					"config_vars.bar": {
						Old: "",
						New: "baz",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"vars": {
					Type:     TypeMap,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"vars.foo": "bar",
				},
			},

			Config: map[string]interface{}{
				"vars": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"vars.foo": {
						Old:        "bar",
						New:        "",
						NewRemoved: true,
					},
					"vars.bar": {
						Old: "",
						New: "baz",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"vars": {
					Type:     TypeMap,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"vars.foo": "bar",
				},
			},

			Config: nil,

			Diff: nil,

			Err: false,
		},

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"config_vars": {
					Type: TypeList,
					Elem: &Schema{Type: TypeMap},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"config_vars.#":     "1",
					"config_vars.0.foo": "bar",
				},
			},

			Config: map[string]interface{}{
				"config_vars": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"config_vars.0.foo": {
						Old:        "bar",
						NewRemoved: true,
					},
					"config_vars.0.bar": {
						Old: "",
						New: "baz",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Maps",
			Schema: map[string]*Schema{
				"config_vars": {
					Type: TypeList,
					Elem: &Schema{Type: TypeMap},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"config_vars.#":     "1",
					"config_vars.0.foo": "bar",
					"config_vars.0.bar": "baz",
				},
			},

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"config_vars.#": {
						Old: "1",
						New: "0",
					},
					"config_vars.0.%": {
						Old: "2",
						New: "0",
					},
					"config_vars.0.foo": {
						Old:        "bar",
						NewRemoved: true,
					},
					"config_vars.0.bar": {
						Old:        "baz",
						NewRemoved: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "ForceNews",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					ForceNew: true,
				},

				"address": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "bar",
					"address":           "foo",
				},
			},

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "bar",
						New:         "foo",
						RequiresNew: true,
					},

					"address": {
						Old:         "foo",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					ForceNew: true,
				},

				"ports": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"availability_zone": "bar",
					"ports.#":           "1",
					"ports.80":          "80",
				},
			},

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "bar",
						New:         "foo",
						RequiresNew: true,
					},

					"ports.#": {
						Old:         "1",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"instances": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
					Computed: true,
					Set: func(v interface{}) int {
						return len(v.(string))
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"instances.#": "0",
				},
			},

			Config: map[string]interface{}{
				"instances": []interface{}{hcl2shim.UnknownVariableValue},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"instances.#": {
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"route": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"index": {
								Type:     TypeInt,
								Required: true,
							},

							"gateway": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
					Set: func(v interface{}) int {
						m := v.(map[string]interface{})
						return m["index"].(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"route": []interface{}{
					map[string]interface{}{
						"index":   "1",
						"gateway": hcl2shim.UnknownVariableValue,
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"route.#": {
						Old: "0",
						New: "1",
					},
					"route.~1.index": {
						Old: "",
						New: "1",
					},
					"route.~1.gateway": {
						Old:         "",
						New:         hcl2shim.UnknownVariableValue,
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set",
			Schema: map[string]*Schema{
				"route": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"index": {
								Type:     TypeInt,
								Required: true,
							},

							"gateway": {
								Type:     TypeSet,
								Optional: true,
								Elem:     &Schema{Type: TypeInt},
								Set: func(a interface{}) int {
									return a.(int)
								},
							},
						},
					},
					Set: func(v interface{}) int {
						m := v.(map[string]interface{})
						return m["index"].(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"route": []interface{}{
					map[string]interface{}{
						"index": "1",
						"gateway": []interface{}{
							hcl2shim.UnknownVariableValue,
						},
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"route.#": {
						Old: "0",
						New: "1",
					},
					"route.~1.index": {
						Old: "",
						New: "1",
					},
					"route.~1.gateway.#": {
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Computed maps",
			Schema: map[string]*Schema{
				"vars": {
					Type:     TypeMap,
					Computed: true,
				},
			},

			State: nil,

			Config: nil,

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"vars.%": {
						Old:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Computed maps",
			Schema: map[string]*Schema{
				"vars": {
					Type:     TypeMap,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"vars.%": "0",
				},
			},

			Config: map[string]interface{}{
				"vars": map[string]interface{}{
					"bar": hcl2shim.UnknownVariableValue,
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"vars.%": {
						Old:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name:   " - Empty",
			Schema: map[string]*Schema{},

			State: &terraform.InstanceState{},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Float",
			Schema: map[string]*Schema{
				"some_threshold": {
					Type: TypeFloat,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"some_threshold": "567.8",
				},
			},

			Config: map[string]interface{}{
				"some_threshold": 12.34,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"some_threshold": {
						Old: "567.8",
						New: "12.34",
					},
				},
			},

			Err: false,
		},

		{
			Name: "https://github.com/hashicorp/terraform-plugin-sdk/issues/824",
			Schema: map[string]*Schema{
				"block_device": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"device_name": {
								Type:     TypeString,
								Required: true,
							},
							"delete_on_termination": {
								Type:     TypeBool,
								Optional: true,
								Default:  true,
							},
						},
					},
					Set: func(v interface{}) int {
						var buf bytes.Buffer
						m := v.(map[string]interface{})
						buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
						buf.WriteString(fmt.Sprintf("%t-", m["delete_on_termination"].(bool)))
						return hashcode.String(buf.String())
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"block_device.#": "2",
					"block_device.616397234.delete_on_termination":  "true",
					"block_device.616397234.device_name":            "/dev/sda1",
					"block_device.2801811477.delete_on_termination": "true",
					"block_device.2801811477.device_name":           "/dev/sdx",
				},
			},

			Config: map[string]interface{}{
				"block_device": []interface{}{
					map[string]interface{}{
						"device_name": "/dev/sda1",
					},
					map[string]interface{}{
						"device_name": "/dev/sdx",
					},
				},
			},
			Diff: nil,
			Err:  false,
		},

		{
			Name: "Zero value in state shouldn't result in diff",
			Schema: map[string]*Schema{
				"port": {
					Type:     TypeBool,
					Optional: true,
					ForceNew: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"port": "false",
				},
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Same as prev, but for sets",
			Schema: map[string]*Schema{
				"route": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"index": {
								Type:     TypeInt,
								Required: true,
							},

							"gateway": {
								Type:     TypeSet,
								Optional: true,
								Elem:     &Schema{Type: TypeInt},
								Set: func(a interface{}) int {
									return a.(int)
								},
							},
						},
					},
					Set: func(v interface{}) int {
						m := v.(map[string]interface{})
						return m["index"].(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"route.#": "0",
				},
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "A set computed element shouldn't cause a diff",
			Schema: map[string]*Schema{
				"active": {
					Type:     TypeBool,
					Computed: true,
					ForceNew: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"active": "true",
				},
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "An empty set should show up in the diff",
			Schema: map[string]*Schema{
				"instances": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
					ForceNew: true,
					Set: func(v interface{}) int {
						return len(v.(string))
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"instances.#": "1",
					"instances.3": "foo",
				},
			},

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"instances.#": {
						Old:         "1",
						New:         "0",
						RequiresNew: true,
					},
					"instances.3": {
						Old:         "foo",
						New:         "",
						NewRemoved:  true,
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Map with empty value",
			Schema: map[string]*Schema{
				"vars": {
					Type: TypeMap,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"vars": map[string]interface{}{
					"foo": "",
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"vars.%": {
						Old: "0",
						New: "1",
					},
					"vars.foo": {
						Old: "",
						New: "",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Unset bool, not in state",
			Schema: map[string]*Schema{
				"force": {
					Type:     TypeBool,
					Optional: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Unset set, not in state",
			Schema: map[string]*Schema{
				"metadata_keys": {
					Type:     TypeSet,
					Optional: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeInt},
					Set:      func(interface{}) int { return 0 },
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Unset list in state, should not show up computed",
			Schema: map[string]*Schema{
				"metadata_keys": {
					Type:     TypeList,
					Optional: true,
					Computed: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeInt},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"metadata_keys.#": "0",
				},
			},

			Config: map[string]interface{}{},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Set element computed element",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ports": []interface{}{1, hcl2shim.UnknownVariableValue},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Computed map without config that's known to be empty does not generate diff",
			Schema: map[string]*Schema{
				"tags": {
					Type:     TypeMap,
					Computed: true,
				},
			},

			Config: nil,

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"tags.%": "0",
				},
			},

			Diff: nil,

			Err: false,
		},

		{
			Name: "Set with hyphen keys",
			Schema: map[string]*Schema{
				"route": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"index": {
								Type:     TypeInt,
								Required: true,
							},

							"gateway-name": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
					Set: func(v interface{}) int {
						m := v.(map[string]interface{})
						return m["index"].(int)
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"route": []interface{}{
					map[string]interface{}{
						"index":        "1",
						"gateway-name": "hello",
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"route.#": {
						Old: "0",
						New: "1",
					},
					"route.1.index": {
						Old: "",
						New: "1",
					},
					"route.1.gateway-name": {
						Old: "",
						New: "hello",
					},
				},
			},

			Err: false,
		},

		{
			Name: ": StateFunc in nested set (#1759)",
			Schema: map[string]*Schema{
				"service_account": {
					Type:     TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"scopes": {
								Type:     TypeSet,
								Required: true,
								ForceNew: true,
								Elem: &Schema{
									Type: TypeString,
									StateFunc: func(v interface{}) string {
										return v.(string) + "!"
									},
								},
								Set: func(v interface{}) int {
									i, err := strconv.Atoi(v.(string))
									if err != nil {
										t.Fatalf("err: %s", err)
									}
									return i
								},
							},
						},
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"service_account": []interface{}{
					map[string]interface{}{
						"scopes": []interface{}{"123"},
					},
				},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"service_account.#": {
						Old:         "0",
						New:         "1",
						RequiresNew: true,
					},
					"service_account.0.scopes.#": {
						Old:         "0",
						New:         "1",
						RequiresNew: true,
					},
					"service_account.0.scopes.123": {
						Old:         "",
						New:         "123!",
						NewExtra:    "123",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Removing set elements",
			Schema: map[string]*Schema{
				"instances": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
					ForceNew: true,
					Set: func(v interface{}) int {
						return len(v.(string))
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"instances.#": "2",
					"instances.3": "333",
					"instances.2": "22",
				},
			},

			Config: map[string]interface{}{
				"instances": []interface{}{"333", "4444"},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"instances.#": {
						Old: "2",
						New: "2",
					},
					"instances.2": {
						Old:         "22",
						New:         "",
						NewRemoved:  true,
						RequiresNew: true,
					},
					"instances.3": {
						Old: "333",
						New: "333",
					},
					"instances.4": {
						Old:         "",
						New:         "4444",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Bools can be set with 0/1 in config, still get true/false",
			Schema: map[string]*Schema{
				"one": {
					Type:     TypeBool,
					Optional: true,
				},
				"two": {
					Type:     TypeBool,
					Optional: true,
				},
				"three": {
					Type:     TypeBool,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"one":   "false",
					"two":   "true",
					"three": "true",
				},
			},

			Config: map[string]interface{}{
				"one": "1",
				"two": "0",
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"one": {
						Old: "false",
						New: "true",
					},
					"two": {
						Old: "true",
						New: "false",
					},
					"three": {
						Old:        "true",
						New:        "false",
						NewRemoved: true,
					},
				},
			},

			Err: false,
		},

		{
			Name:   "tainted in state w/ no attr changes is still a replacement",
			Schema: map[string]*Schema{},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"id": "someid",
				},
				Tainted: true,
			},

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes:     map[string]*terraform.ResourceAttrDiff{},
				DestroyTainted: true,
			},

			Err: false,
		},

		{
			Name: "Set ForceNew only marks the changing element as ForceNew",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "3",
					"ports.1": "1",
					"ports.2": "2",
					"ports.4": "4",
				},
			},

			Config: map[string]interface{}{
				"ports": []interface{}{5, 2, 1},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "3",
						New: "3",
					},
					"ports.1": {
						Old: "1",
						New: "1",
					},
					"ports.2": {
						Old: "2",
						New: "2",
					},
					"ports.5": {
						Old:         "",
						New:         "5",
						RequiresNew: true,
					},
					"ports.4": {
						Old:         "4",
						New:         "0",
						NewRemoved:  true,
						RequiresNew: true,
					},
				},
			},
		},

		{
			Name: "removed optional items should trigger ForceNew",
			Schema: map[string]*Schema{
				"description": {
					Type:     TypeString,
					ForceNew: true,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"description": "foo",
				},
			},

			Config: map[string]interface{}{},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"description": {
						Old:         "foo",
						New:         "",
						RequiresNew: true,
						NewRemoved:  true,
					},
				},
			},

			Err: false,
		},

		// GH-7715
		{
			Name: "computed value for boolean field",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeBool,
					ForceNew: true,
					Computed: true,
					Optional: true,
				},
			},

			State: &terraform.InstanceState{},

			Config: map[string]interface{}{
				"foo": hcl2shim.UnknownVariableValue,
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"foo": {
						Old:         "",
						New:         "false",
						NewComputed: true,
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set ForceNew marks count as ForceNew if computed",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "3",
					"ports.1": "1",
					"ports.2": "2",
					"ports.4": "4",
				},
			},

			Config: map[string]interface{}{
				"ports": []interface{}{hcl2shim.UnknownVariableValue, 2, 1},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old:         "3",
						New:         "",
						NewComputed: true,
						RequiresNew: true,
					},
				},
			},
		},

		{
			Name: "List with computed schema and ForceNew",
			Schema: map[string]*Schema{
				"config": {
					Type:     TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &Schema{
						Type: TypeString,
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"config.#": "2",
					"config.0": "a",
					"config.1": "b",
				},
			},

			Config: map[string]interface{}{
				"config": []interface{}{hcl2shim.UnknownVariableValue, hcl2shim.UnknownVariableValue},
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"config.#": {
						Old:         "2",
						New:         "",
						RequiresNew: true,
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "overridden diff with a CustomizeDiff function, ForceNew not in schema",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNew("availability_zone", "bar"); err != nil {
					return err
				}
				if err := d.ForceNew("availability_zone"); err != nil {
					return err
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			// NOTE: This case is technically impossible in the current
			// implementation, because optional+computed values never show up in the
			// diff. In the event behavior changes this test should ensure that the
			// intended diff still shows up.
			Name: "overridden removed attribute diff with a CustomizeDiff function, ForceNew not in schema",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNew("availability_zone", "bar"); err != nil {
					return err
				}
				if err := d.ForceNew("availability_zone"); err != nil {
					return err
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{

			Name: "overridden diff with a CustomizeDiff function, ForceNew in schema",
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNew("availability_zone", "bar"); err != nil {
					return err
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old:         "",
						New:         "bar",
						RequiresNew: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "required field with computed diff added with CustomizeDiff function",
			Schema: map[string]*Schema{
				"ami_id": {
					Type:     TypeString,
					Required: true,
				},
				"instance_id": {
					Type:     TypeString,
					Computed: true,
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"ami_id": "foo",
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNew("instance_id", "bar"); err != nil {
					return err
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ami_id": {
						Old: "",
						New: "foo",
					},
					"instance_id": {
						Old: "",
						New: "bar",
					},
				},
			},

			Err: false,
		},

		{
			Name: "Set ForceNew only marks the changing element as ForceNew - CustomizeDiffFunc edition",
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"ports.#": "3",
					"ports.1": "1",
					"ports.2": "2",
					"ports.4": "4",
				},
			},

			Config: map[string]interface{}{
				"ports": []interface{}{5, 2, 6},
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNew("ports", []interface{}{5, 2, 1}); err != nil {
					return err
				}
				if err := d.ForceNew("ports"); err != nil {
					return err
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"ports.#": {
						Old: "3",
						New: "3",
					},
					"ports.1": {
						Old: "1",
						New: "1",
					},
					"ports.2": {
						Old: "2",
						New: "2",
					},
					"ports.5": {
						Old:         "",
						New:         "5",
						RequiresNew: true,
					},
					"ports.4": {
						Old:         "4",
						New:         "0",
						NewRemoved:  true,
						RequiresNew: true,
					},
				},
			},
		},

		{
			Name:   "tainted resource does not run CustomizeDiffFunc",
			Schema: map[string]*Schema{},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"id": "someid",
				},
				Tainted: true,
			},

			Config: map[string]interface{}{},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				return errors.New("diff customization should not have run")
			},

			Diff: &terraform.InstanceDiff{
				Attributes:     map[string]*terraform.ResourceAttrDiff{},
				DestroyTainted: true,
			},

			Err: false,
		},

		{
			Name: "NewComputed based on a conditional with CustomizeDiffFunc",
			Schema: map[string]*Schema{
				"etag": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
				"version_id": {
					Type:     TypeString,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"etag":       "foo",
					"version_id": "1",
				},
			},

			Config: map[string]interface{}{
				"etag": "bar",
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if d.HasChange("etag") {
					if err := d.SetNewComputed("version_id"); err != nil {
						return fmt.Errorf("unexpected SetNewComputed error: %w", err)
					}
				}
				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"etag": {
						Old: "foo",
						New: "bar",
					},
					"version_id": {
						Old:         "1",
						New:         "",
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "NewComputed should always propagate with CustomizeDiff",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"foo": "",
				},
				ID: "pre-existing",
			},

			Config: map[string]interface{}{},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				if err := d.SetNewComputed("foo"); err != nil {
					return fmt.Errorf("unexpected SetNewComputed error: %w", err)
				}

				return nil
			},

			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"foo": {
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		{
			Name: "vetoing a diff",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"foo": "bar",
				},
			},

			Config: map[string]interface{}{
				"foo": "baz",
			},

			CustomizeDiff: func(_ context.Context, d *ResourceDiff, meta interface{}) error {
				return fmt.Errorf("diff vetoed")
			},

			Err: true,
		},

		// A lot of resources currently depended on using the empty string as a
		// nil/unset value.
		// FIXME: We want this to eventually produce a diff, since there
		// technically is a new value in the config.
		{
			Name: "optional, computed, empty string",
			Schema: map[string]*Schema{
				"attr": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"attr": "bar",
				},
			},

			Config: map[string]interface{}{
				"attr": "",
			},
		},

		{
			Name: "optional, computed, empty string should not crash in CustomizeDiff",
			Schema: map[string]*Schema{
				"unrelated_set": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"stream_enabled": {
					Type:     TypeBool,
					Optional: true,
				},
				"stream_view_type": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
				},
			},

			State: &terraform.InstanceState{
				Attributes: map[string]string{
					"unrelated_set.#":  "0",
					"stream_enabled":   "true",
					"stream_view_type": "KEYS_ONLY",
				},
			},
			Config: map[string]interface{}{
				"stream_enabled":   false,
				"stream_view_type": "",
			},
			CustomizeDiff: func(_ context.Context, diff *ResourceDiff, _ interface{}) error {
				v, ok := diff.GetOk("unrelated_set")
				if ok {
					return fmt.Errorf("Didn't expect unrelated_set: %#v", v)
				}
				return nil
			},
			Diff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"stream_enabled": {
						Old: "true",
						New: "false",
					},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)

			d, err := schemaMap(tc.Schema).Diff(context.Background(), tc.State, c, tc.CustomizeDiff, nil, true)
			if err != nil != tc.Err {
				t.Fatalf("err: %s", err)
			}

			if !reflect.DeepEqual(tc.Diff, d) {
				t.Fatalf("expected:\n%#v\n\ngot:\n%#v", tc.Diff, d)
			}
		})
	}
}

func TestSchemaMap_InternalValidate(t *testing.T) {
	cases := map[string]struct {
		In  map[string]*Schema
		Err bool
	}{
		"nothing": {
			nil,
			false,
		},

		"Both optional and required": {
			map[string]*Schema{
				"foo": {
					Type:     TypeInt,
					Optional: true,
					Required: true,
				},
			},
			true,
		},

		"No optional and no required": {
			map[string]*Schema{
				"foo": {
					Type: TypeInt,
				},
			},
			true,
		},

		"Missing Type": {
			map[string]*Schema{
				"foo": {
					Required: true,
				},
			},
			true,
		},

		"Required but computed": {
			map[string]*Schema{
				"foo": {
					Type:     TypeInt,
					Required: true,
					Computed: true,
				},
			},
			true,
		},

		"Looks good": {
			map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Required: true,
				},
			},
			false,
		},

		"Computed but has default": {
			map[string]*Schema{
				"foo": {
					Type:     TypeInt,
					Optional: true,
					Computed: true,
					Default:  "foo",
				},
			},
			true,
		},

		"Required but has default": {
			map[string]*Schema{
				"foo": {
					Type:     TypeInt,
					Optional: true,
					Required: true,
					Default:  "foo",
				},
			},
			true,
		},

		"List element not set": {
			map[string]*Schema{
				"foo": {
					Type: TypeList,
				},
			},
			true,
		},

		"List default": {
			map[string]*Schema{
				"foo": {
					Type:    TypeList,
					Elem:    &Schema{Type: TypeInt},
					Default: "foo",
				},
			},
			true,
		},

		"List element computed": {
			map[string]*Schema{
				"foo": {
					Type:     TypeList,
					Optional: true,
					Elem: &Schema{
						Type:     TypeInt,
						Computed: true,
					},
				},
			},
			true,
		},

		"List element with Set set": {
			map[string]*Schema{
				"foo": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeInt},
					Set:      func(interface{}) int { return 0 },
					Optional: true,
				},
			},
			true,
		},

		"Set element with no Set set": {
			map[string]*Schema{
				"foo": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeInt},
					Optional: true,
				},
			},
			false,
		},

		"Required but computedWhen": {
			map[string]*Schema{
				"foo": {
					Type:         TypeInt,
					Required:     true,
					ComputedWhen: []string{"foo"},
				},
			},
			true,
		},

		"Conflicting attributes cannot be required": {
			map[string]*Schema{
				"blacklist": {
					Type:     TypeBool,
					Required: true,
				},
				"whitelist": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"blacklist"},
				},
			},
			true,
		},

		"Attribute with conflicts cannot be required": {
			map[string]*Schema{
				"whitelist": {
					Type:          TypeBool,
					Required:      true,
					ConflictsWith: []string{"blacklist"},
				},
			},
			true,
		},

		"ConflictsWith cannot be used w/ ComputedWhen": {
			map[string]*Schema{
				"blacklist": {
					Type:         TypeBool,
					ComputedWhen: []string{"foor"},
				},
				"whitelist": {
					Type:          TypeBool,
					Required:      true,
					ConflictsWith: []string{"blacklist"},
				},
			},
			true,
		},

		"AtLeastOneOf list index syntax with self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								AtLeastOneOf: []string{"config_block_attr.0.nested_attr"},
							},
						},
					},
				},
			},
			false,
		},

		"AtLeastOneOf list index syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.0.nested_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf list index syntax with list configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf list index syntax with list configuration block missing MaxItems": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf list index syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.0.nested_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf list index syntax with set configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf map key syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf map key syntax with list configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								AtLeastOneOf: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"AtLeastOneOf map key syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"AtLeastOneOf map key syntax with set configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								AtLeastOneOf: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"AtLeastOneOf map key syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"map_attr.some_key"},
				},
			},
			true,
		},

		"AtLeastOneOf string syntax with list attribute": {
			map[string]*Schema{
				"list_attr": {
					Type:     TypeList,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"list_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf string syntax with list configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf string syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"map_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf string syntax with set attribute": {
			map[string]*Schema{
				"set_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"set_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf string syntax with set configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"config_block_attr"},
				},
			},
			false,
		},

		"AtLeastOneOf string syntax with self reference": {
			map[string]*Schema{
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"test"},
				},
			},
			false,
		},

		"ConflictsWith list index syntax with self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"config_block_attr.0.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"ConflictsWith list index syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.0.nested_attr"},
				},
			},
			false,
		},

		"ConflictsWith list index syntax with list configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ConflictsWith list index syntax with list configuration block missing MaxItems": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ConflictsWith list index syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.0.nested_attr"},
				},
			},
			true,
		},

		"ConflictsWith list index syntax with set configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ConflictsWith map key syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"ConflictsWith map key syntax with list configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"ConflictsWith map key syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"ConflictsWith map key syntax with set configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"ConflictsWith map key syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"map_attr.some_key"},
				},
			},
			true,
		},

		"ConflictsWith string syntax with list attribute": {
			map[string]*Schema{
				"list_attr": {
					Type:     TypeList,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"list_attr"},
				},
			},
			false,
		},

		"ConflictsWith string syntax with list configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr"},
				},
			},
			false,
		},

		"ConflictsWith string syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"map_attr"},
				},
			},
			false,
		},

		"ConflictsWith string syntax with set attribute": {
			map[string]*Schema{
				"set_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"set_attr"},
				},
			},
			false,
		},

		"ConflictsWith string syntax with set configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"config_block_attr"},
				},
			},
			false,
		},

		"ConflictsWith string syntax with self reference": {
			map[string]*Schema{
				"test": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"test"},
				},
			},
			true,
		},

		"ExactlyOneOf list index syntax with self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								ExactlyOneOf: []string{"config_block_attr.0.nested_attr"},
							},
						},
					},
				},
			},
			false,
		},

		"ExactlyOneOf list index syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.0.nested_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf list index syntax with list configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf list index syntax with list configuration block missing MaxItems": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf list index syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.0.nested_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf list index syntax with set configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf map key syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf map key syntax with list configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								ExactlyOneOf: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"ExactlyOneOf map key syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"ExactlyOneOf map key syntax with set configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								ExactlyOneOf: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"ExactlyOneOf map key syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"map_attr.some_key"},
				},
			},
			true,
		},

		"ExactlyOneOf string syntax with list attribute": {
			map[string]*Schema{
				"list_attr": {
					Type:     TypeList,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"list_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf string syntax with list configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf string syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"map_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf string syntax with set attribute": {
			map[string]*Schema{
				"set_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"set_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf string syntax with set configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"config_block_attr"},
				},
			},
			false,
		},

		"ExactlyOneOf string syntax with self reference": {
			map[string]*Schema{
				"test": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"test"},
				},
			},
			false,
		},

		"RequiredWith list index syntax with self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								RequiredWith: []string{"config_block_attr.0.nested_attr"},
							},
						},
					},
				},
			},
			false,
		},

		"RequiredWith list index syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.0.nested_attr"},
				},
			},
			false,
		},

		"RequiredWith list index syntax with list configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"RequiredWith list index syntax with list configuration block missing MaxItems": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"RequiredWith list index syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.0.nested_attr"},
				},
			},
			true,
		},

		"RequiredWith list index syntax with set configuration block missing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.0.missing_attr"},
				},
			},
			true,
		},

		"RequiredWith map key syntax with list configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"RequiredWith map key syntax with list configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								RequiredWith: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"RequiredWith map key syntax with set configuration block existing attribute": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr.nested_attr"},
				},
			},
			true,
		},

		"RequiredWith map key syntax with set configuration block self reference": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:         TypeString,
								Optional:     true,
								RequiredWith: []string{"config_block_attr.nested_attr"},
							},
						},
					},
				},
			},
			true,
		},

		"RequiredWith map key syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"map_attr.some_key"},
				},
			},
			true,
		},

		"RequiredWith string syntax with list attribute": {
			map[string]*Schema{
				"list_attr": {
					Type:     TypeList,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"list_attr"},
				},
			},
			false,
		},

		"RequiredWith string syntax with list configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr"},
				},
			},
			false,
		},

		"RequiredWith string syntax with map attribute": {
			map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"map_attr"},
				},
			},
			false,
		},

		"RequiredWith string syntax with set attribute": {
			map[string]*Schema{
				"set_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"set_attr"},
				},
			},
			false,
		},

		"RequiredWith string syntax with set configuration block": {
			map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"config_block_attr"},
				},
			},
			false,
		},

		"RequiredWith string syntax with self reference": {
			map[string]*Schema{
				"test": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"test"},
				},
			},
			false,
		},

		"Sub-resource invalid": {
			map[string]*Schema{
				"foo": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"foo": new(Schema),
						},
					},
				},
			},
			true,
		},

		"Sub-resource valid": {
			map[string]*Schema{
				"foo": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"foo": {
								Type:     TypeInt,
								Optional: true,
							},
						},
					},
				},
			},
			false,
		},

		"ValidateFunc on non-primitive": {
			map[string]*Schema{
				"foo": {
					Type:     TypeSet,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						return
					},
				},
			},
			true,
		},

		"Computed-only with AtLeastOneOf": {
			map[string]*Schema{
				"string_one": {
					Type:     TypeString,
					Computed: true,
					AtLeastOneOf: []string{
						"string_one",
						"string_two",
					},
				},
				"string_two": {
					Type:     TypeString,
					Computed: true,
					AtLeastOneOf: []string{
						"string_one",
						"string_two",
					},
				},
			},
			true,
		},

		"Computed-only with ConflictsWith": {
			map[string]*Schema{
				"string_one": {
					Type:     TypeString,
					Computed: true,
					ConflictsWith: []string{
						"string_two",
					},
				},
				"string_two": {
					Type:     TypeString,
					Computed: true,
					ConflictsWith: []string{
						"string_one",
					},
				},
			},
			true,
		},

		"Computed-only with Default": {
			map[string]*Schema{
				"string": {
					Type:     TypeString,
					Computed: true,
					Default:  "test",
				},
			},
			true,
		},

		"Computed-only with DefaultFunc": {
			map[string]*Schema{
				"string": {
					Type:        TypeString,
					Computed:    true,
					DefaultFunc: func() (interface{}, error) { return nil, nil },
				},
			},
			true,
		},

		"Computed-only with DiffSuppressFunc": {
			map[string]*Schema{
				"string": {
					Type:             TypeString,
					Computed:         true,
					DiffSuppressFunc: func(k, oldValue, newValue string, d *ResourceData) bool { return false },
				},
			},
			true,
		},

		"DiffSuppressOnRefresh without DiffSuppressFunc": {
			map[string]*Schema{
				"string": {
					Type:                  TypeString,
					Optional:              true,
					DiffSuppressOnRefresh: true,
				},
			},
			true,
		},

		"DiffSuppressOnRefresh with DiffSuppressFunc": {
			map[string]*Schema{
				"string": {
					Type:                  TypeString,
					Optional:              true,
					DiffSuppressFunc:      func(k, oldValue, newValue string, d *ResourceData) bool { return false },
					DiffSuppressOnRefresh: true,
				},
			},
			false,
		},

		"Computed-only with ExactlyOneOf": {
			map[string]*Schema{
				"string_one": {
					Type:     TypeString,
					Computed: true,
					ExactlyOneOf: []string{
						"string_one",
						"string_two",
					},
				},
				"string_two": {
					Type:     TypeString,
					Computed: true,
					ExactlyOneOf: []string{
						"string_one",
						"string_two",
					},
				},
			},
			true,
		},

		"Computed-only with InputDefault": {
			map[string]*Schema{
				"string": {
					Type:         TypeString,
					Computed:     true,
					InputDefault: "test",
				},
			},
			true,
		},

		"Computed-only with MaxItems": {
			map[string]*Schema{
				"string": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Computed: true,
					MaxItems: 1,
				},
			},
			true,
		},

		"Computed-only with MinItems": {
			map[string]*Schema{
				"string": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Computed: true,
					MinItems: 1,
				},
			},
			true,
		},

		"Computed-only with StateFunc": {
			map[string]*Schema{
				"string": {
					Type:      TypeString,
					Computed:  true,
					StateFunc: func(v interface{}) string { return "" },
				},
			},
			true,
		},

		"Computed-only with ValidateFunc": {
			map[string]*Schema{
				"string": {
					Type:         TypeString,
					Computed:     true,
					ValidateFunc: func(v interface{}, k string) ([]string, []error) { return nil, nil },
				},
			},
			true,
		},

		"invalid field name format #1": {
			map[string]*Schema{
				"with space": {
					Type:     TypeString,
					Optional: true,
				},
			},
			true,
		},

		"invalid field name format #2": {
			map[string]*Schema{
				"WithCapitals": {
					Type:     TypeString,
					Optional: true,
				},
			},
			true,
		},

		"invalid field name format of a Deprecated field": {
			map[string]*Schema{
				"WithCapitals": {
					Type:       TypeString,
					Optional:   true,
					Deprecated: "Use with_underscores instead",
				},
			},
			false,
		},

		"ConfigModeBlock with Elem *Resource": {
			map[string]*Schema{
				"block": {
					Type:       TypeList,
					ConfigMode: SchemaConfigModeBlock,
					Optional:   true,
					Elem:       &Resource{},
				},
			},
			false,
		},

		"ConfigModeBlock Computed with Elem *Resource": {
			map[string]*Schema{
				"block": {
					Type:       TypeList,
					ConfigMode: SchemaConfigModeBlock,
					Computed:   true,
					Elem:       &Resource{},
				},
			},
			true, // ConfigMode of block cannot be used for computed schema
		},

		"ConfigModeBlock with Elem *Schema": {
			map[string]*Schema{
				"block": {
					Type:       TypeList,
					ConfigMode: SchemaConfigModeBlock,
					Optional:   true,
					Elem: &Schema{
						Type: TypeString,
					},
				},
			},
			true,
		},

		"ConfigModeBlock with no Elem": {
			map[string]*Schema{
				"block": {
					Type:       TypeString,
					ConfigMode: SchemaConfigModeBlock,
					Optional:   true,
				},
			},
			true,
		},

		"ConfigModeBlock inside ConfigModeAttr": {
			map[string]*Schema{
				"block": {
					Type:       TypeList,
					ConfigMode: SchemaConfigModeAttr,
					Optional:   true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"sub": {
								Type:       TypeList,
								ConfigMode: SchemaConfigModeBlock,
								Elem:       &Resource{},
							},
						},
					},
				},
			},
			true, // ConfigMode of block cannot be used in child of schema with ConfigMode of attribute
		},

		"ConfigModeAuto with *Resource inside ConfigModeAttr": {
			map[string]*Schema{
				"block": {
					Type:       TypeList,
					ConfigMode: SchemaConfigModeAttr,
					Optional:   true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"sub": {
								Type: TypeList,
								Elem: &Resource{},
							},
						},
					},
				},
			},
			true, // in *schema.Resource with ConfigMode of attribute, so must also have ConfigMode of attribute
		},

		"TypeMap with Elem *Resource": {
			map[string]*Schema{
				"map": {
					Type:     TypeMap,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"keynothandled": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			true,
		},

		"ValidateFunc and ValidateDiagFunc cannot both be set": {
			map[string]*Schema{
				"foo": {
					Type:     TypeInt,
					Required: true,
					ValidateFunc: func(interface{}, string) ([]string, []error) {
						return nil, nil
					},
					ValidateDiagFunc: func(interface{}, cty.Path) diag.Diagnostics {
						return nil
					},
				},
			},
			true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			err := schemaMap(tc.In).InternalValidate(nil)
			if err != nil != tc.Err {
				if tc.Err {
					t.Fatalf("%q: Expected error did not occur:\n\n%#v", tn, tc.In)
				}
				t.Fatalf("%q: Unexpected error occurred: %s\n\n%#v", tn, err, tc.In)
			}
		})
	}

}

func TestSchemaMap_DiffSuppress(t *testing.T) {
	cases := map[string]struct {
		Schema       map[string]*Schema
		State        *terraform.InstanceState
		Config       map[string]interface{}
		ExpectedDiff *terraform.InstanceDiff
		Err          bool
	}{
		"#0 - Suppress otherwise valid diff by returning true": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					DiffSuppressFunc: func(k, oldValue, newValue string, d *ResourceData) bool {
						// Always suppress any diff
						return true
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			ExpectedDiff: nil,

			Err: false,
		},

		"#1 - Don't suppress diff by returning false": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					DiffSuppressFunc: func(k, oldValue, newValue string, d *ResourceData) bool {
						// Always suppress any diff
						return false
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},

			ExpectedDiff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "",
						New: "foo",
					},
				},
			},

			Err: false,
		},

		"Default with suppress makes no diff": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Default:  "foo",
					DiffSuppressFunc: func(k, oldValue, newValue string, d *ResourceData) bool {
						return true
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			ExpectedDiff: nil,

			Err: false,
		},

		"Default with false suppress makes diff": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Default:  "foo",
					DiffSuppressFunc: func(k, oldValue, newValue string, d *ResourceData) bool {
						return false
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{},

			ExpectedDiff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"availability_zone": {
						Old: "",
						New: "foo",
					},
				},
			},

			Err: false,
		},

		"Complex structure with set of computed string should mark root set as computed": {
			Schema: map[string]*Schema{
				"outer": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"outer_str": {
								Type:     TypeString,
								Optional: true,
							},
							"inner": {
								Type:     TypeSet,
								Optional: true,
								Elem: &Resource{
									Schema: map[string]*Schema{
										"inner_str": {
											Type:     TypeString,
											Optional: true,
										},
									},
								},
								Set: func(v interface{}) int {
									return 2
								},
							},
						},
					},
					Set: func(v interface{}) int {
						return 1
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"outer": []interface{}{
					map[string]interface{}{
						"outer_str": "foo",
						"inner": []interface{}{
							map[string]interface{}{
								"inner_str": hcl2shim.UnknownVariableValue,
							},
						},
					},
				},
			},

			ExpectedDiff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"outer.#": {
						Old: "0",
						New: "1",
					},
					"outer.~1.outer_str": {
						Old: "",
						New: "foo",
					},
					"outer.~1.inner.#": {
						Old: "0",
						New: "1",
					},
					"outer.~1.inner.~2.inner_str": {
						Old:         "",
						New:         hcl2shim.UnknownVariableValue,
						NewComputed: true,
					},
				},
			},

			Err: false,
		},

		"Complex structure with complex list of computed string should mark root set as computed": {
			Schema: map[string]*Schema{
				"outer": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"outer_str": {
								Type:     TypeString,
								Optional: true,
							},
							"inner": {
								Type:     TypeList,
								Optional: true,
								Elem: &Resource{
									Schema: map[string]*Schema{
										"inner_str": {
											Type:     TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
					Set: func(v interface{}) int {
						return 1
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"outer": []interface{}{
					map[string]interface{}{
						"outer_str": "foo",
						"inner": []interface{}{
							map[string]interface{}{
								"inner_str": hcl2shim.UnknownVariableValue,
							},
						},
					},
				},
			},

			ExpectedDiff: &terraform.InstanceDiff{
				Attributes: map[string]*terraform.ResourceAttrDiff{
					"outer.#": {
						Old: "0",
						New: "1",
					},
					"outer.~1.outer_str": {
						Old: "",
						New: "foo",
					},
					"outer.~1.inner.#": {
						Old: "0",
						New: "1",
					},
					"outer.~1.inner.0.inner_str": {
						Old:         "",
						New:         hcl2shim.UnknownVariableValue,
						NewComputed: true,
					},
				},
			},

			Err: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)

			d, err := schemaMap(tc.Schema).Diff(context.Background(), tc.State, c, nil, nil, true)
			if err != nil != tc.Err {
				t.Fatalf("#%q err: %s", tn, err)
			}

			if !reflect.DeepEqual(tc.ExpectedDiff, d) {
				t.Fatalf("#%q:\n\nexpected:\n%#v\n\ngot:\n%#v", tn, tc.ExpectedDiff, d)
			}
		})
	}
}

func TestSchema_DiffSuppressOnRefresh(t *testing.T) {
	cases := map[string]struct {
		Schema     schemaMap
		PriorState map[string]string
		SetKey     string
		SetVal     interface{}
		WantState  map[string]string
	}{
		"no suppress func string": {
			Schema: schemaMap{
				"v": {
					Type:     TypeString,
					Optional: true,
				},
			},
			PriorState: map[string]string{
				"v": "hello",
			},
			SetKey: "v",
			SetVal: "howdy",
			WantState: map[string]string{
				"v": "howdy", // set was honored
			},
		},
		"suppress func string but not always": {
			Schema: schemaMap{
				"v": {
					Type:     TypeString,
					Optional: true,
					DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
						return true
					},
				},
			},
			PriorState: map[string]string{
				"v": "hello",
			},
			SetKey: "v",
			SetVal: "howdy",
			WantState: map[string]string{
				"v": "howdy", // set was honored
			},
		},
		"suppress func string always": {
			Schema: schemaMap{
				"v": {
					Type:     TypeString,
					Optional: true,
					DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
						return true
					},
					DiffSuppressOnRefresh: true,
				},
			},
			PriorState: map[string]string{
				"v": "hello",
			},
			SetKey: "v",
			SetVal: "howdy",
			WantState: map[string]string{
				"v": "hello", // set was ignored
			},
		},
		"suppress func string always no prior": {
			Schema: schemaMap{
				"v": {
					Type:     TypeString,
					Optional: true,
					DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
						return true
					},
					DiffSuppressOnRefresh: true,
				},
			},
			PriorState: map[string]string{},
			SetKey:     "v",
			SetVal:     "howdy",
			WantState: map[string]string{
				"v": "howdy", // set was honored
			},
		},
		"suppress func nested string but not always": {
			Schema: schemaMap{
				"v": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"w": {
								Type: TypeString,
								DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
									return true
								},
							},
						},
					},
				},
			},
			PriorState: map[string]string{
				"v.#":   "1",
				"v.0.w": "hello",
			},
			SetKey: "v",
			SetVal: []map[string]interface{}{{"w": "howdy"}},
			WantState: map[string]string{
				"v.#":   "1",
				"v.0.w": "howdy", // set was honored
			},
		},
		"suppress func nested string always": {
			Schema: schemaMap{
				"v": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"w": {
								Type: TypeString,
								DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
									return true
								},
								DiffSuppressOnRefresh: true,
							},
						},
					},
				},
				"unrelated": {
					Type:     TypeString,
					Optional: true,
				},
			},
			PriorState: map[string]string{
				"v.#":       "1",
				"v.0.w":     "hello",
				"unrelated": "hi",
			},
			SetKey: "v",
			SetVal: []map[string]interface{}{{"w": "howdy"}},
			WantState: map[string]string{
				"v.#":       "1",
				"v.0.w":     "hello", // set was ignored
				"unrelated": "hi",
			},
		},
		"suppress func nested string always no prior": {
			Schema: schemaMap{
				"v": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"w": {
								Type: TypeString,
								DiffSuppressFunc: func(key, oldV, newV string, d *ResourceData) bool {
									return true
								},
								DiffSuppressOnRefresh: true,
							},
						},
					},
				},
			},
			PriorState: map[string]string{
				"v.#": "0",
			},
			SetKey: "v",
			SetVal: []map[string]interface{}{{"w": "howdy"}},
			WantState: map[string]string{
				"v.#":   "1",
				"v.0.w": "howdy", // set was honored
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			schema := tc.Schema
			priorState := &terraform.InstanceState{
				Attributes: tc.PriorState,
			}

			d, err := schema.Data(priorState, nil)
			if err != nil {
				t.Fatalf("failed to create ResourceData: %s", err)
			}

			d.SetId("-") // just to make d.State think this object exists

			err = d.Set(tc.SetKey, tc.SetVal)
			if err != nil {
				t.Fatalf("failed to Set: %s", err)
			}

			newState := d.State()
			schema.handleDiffSuppressOnRefresh(context.Background(), priorState, newState)
			var newStateAttrs map[string]string
			if newState != nil {
				newStateAttrs = newState.Attributes
				delete(newStateAttrs, "id")
			}

			if diff := cmp.Diff(tc.WantState, newStateAttrs); diff != "" {
				t.Errorf("wrong result state\n%s", diff)
			}
		})
	}
}

func TestSchemaMap_Validate(t *testing.T) {
	cases := map[string]struct {
		Schema   map[string]*Schema
		Config   map[string]interface{}
		Err      bool
		Errors   []error
		Warnings []string
	}{
		"Good": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			Config: map[string]interface{}{
				"availability_zone": "foo",
			},
		},

		"Good, because the var is not set and that error will come elsewhere": {
			Schema: map[string]*Schema{
				"size": {
					Type:     TypeInt,
					Required: true,
				},
			},

			Config: map[string]interface{}{
				"size": hcl2shim.UnknownVariableValue,
			},
		},

		"Required field not set": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Required: true,
				},
			},

			Config: map[string]interface{}{},

			Err: true,
		},

		"Invalid basic type": {
			Schema: map[string]*Schema{
				"port": {
					Type:     TypeInt,
					Required: true,
				},
			},

			Config: map[string]interface{}{
				"port": "I am invalid",
			},

			Err: true,
		},

		"Invalid complex type": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeString,
					Optional: true,
				},
			},

			Config: map[string]interface{}{
				"user_data": []interface{}{
					map[string]interface{}{
						"foo": "bar",
					},
				},
			},

			Err: true,
		},

		"Bad type": {
			Schema: map[string]*Schema{
				"size": {
					Type:     TypeInt,
					Required: true,
				},
			},

			Config: map[string]interface{}{
				"size": "nope",
			},

			Err: true,
		},

		"Required but has DefaultFunc": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Required: true,
					DefaultFunc: func() (interface{}, error) {
						return "foo", nil
					},
				},
			},

			Config: nil,
		},

		"Required but has DefaultFunc return nil": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Required: true,
					DefaultFunc: func() (interface{}, error) {
						return nil, nil
					},
				},
			},

			Config: nil,

			Err: true,
		},

		"Optional sub-resource": {
			Schema: map[string]*Schema{
				"ingress": {
					Type: TypeList,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{},

			Err: false,
		},

		"Sub-resource is the wrong type": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Required: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{"foo"},
			},

			Err: true,
		},

		"Not a list nested block": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": "foo",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf("Error: Attribute must be a list"),
			},
		},

		"Not a list primitive": {
			Schema: map[string]*Schema{
				"strings": {
					Type:     TypeList,
					Optional: true,
					Elem: &Schema{
						Type: TypeString,
					},
				},
			},

			Config: map[string]interface{}{
				"strings": "foo",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf("Error: Attribute must be a list"),
			},
		},

		"Unknown list": {
			Schema: map[string]*Schema{
				"strings": {
					Type:     TypeList,
					Optional: true,
					Elem: &Schema{
						Type: TypeString,
					},
				},
			},

			Config: map[string]interface{}{
				"strings": hcl2shim.UnknownVariableValue,
			},

			Err: false,
		},

		"Unknown + Deprecation": {
			Schema: map[string]*Schema{
				"old_news": {
					Type:       TypeString,
					Optional:   true,
					Deprecated: "please use 'new_news' instead",
				},
			},

			Config: map[string]interface{}{
				"old_news": hcl2shim.UnknownVariableValue,
			},
			Err: false,
		},

		"Required sub-resource field": {
			Schema: map[string]*Schema{
				"ingress": {
					Type: TypeList,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{},
				},
			},

			Err: true,
		},

		"Good sub-resource": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"from": 80,
					},
				},
			},

			Err: false,
		},

		"Good sub-resource, computed value": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"from": {
								Type:     TypeInt,
								Optional: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"from": hcl2shim.UnknownVariableValue,
					},
				},
			},

			Err: false,
		},

		"Invalid/unknown field": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			Config: map[string]interface{}{
				"foo": "bar",
			},

			Err: true,
		},

		"Invalid/unknown field with computed value": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			Config: map[string]interface{}{
				"foo": hcl2shim.UnknownVariableValue,
			},

			Err: true,
		},

		"Computed field set": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Computed: true,
				},
			},

			Config: map[string]interface{}{
				"availability_zone": "bar",
			},

			Err: true,
		},

		"Not a set": {
			Schema: map[string]*Schema{
				"ports": {
					Type:     TypeSet,
					Required: true,
					Elem:     &Schema{Type: TypeInt},
					Set: func(a interface{}) int {
						return a.(int)
					},
				},
			},

			Config: map[string]interface{}{
				"ports": "foo",
			},

			Err: true,
		},

		"Maps": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
				},
			},

			Config: map[string]interface{}{
				"user_data": "foo",
			},

			Err: true,
		},

		"Good map: data surrounded by extra slice": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
				},
			},

			Config: map[string]interface{}{
				"user_data": []interface{}{
					map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},

		"Good map": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
				},
			},

			Config: map[string]interface{}{
				"user_data": map[string]interface{}{
					"foo": "bar",
				},
			},
		},

		"Map with type specified as value type": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
					Elem:     TypeBool,
				},
			},

			Config: map[string]interface{}{
				"user_data": map[string]interface{}{
					"foo": "not_a_bool",
				},
			},

			Err: true,
		},

		"Map with type specified as nested Schema": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
					Elem:     &Schema{Type: TypeBool},
				},
			},

			Config: map[string]interface{}{
				"user_data": map[string]interface{}{
					"foo": "not_a_bool",
				},
			},

			Err: true,
		},

		"Bad map: just a slice": {
			Schema: map[string]*Schema{
				"user_data": {
					Type:     TypeMap,
					Optional: true,
				},
			},

			Config: map[string]interface{}{
				"user_data": []interface{}{
					"foo",
				},
			},

			Err: true,
		},

		"Good set: config has slice with single interpolated value": {
			Schema: map[string]*Schema{
				"security_groups": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeString},
					Set: func(v interface{}) int {
						return len(v.(string))
					},
				},
			},

			Config: map[string]interface{}{
				"security_groups": []interface{}{"${var.foo}"},
			},

			Err: false,
		},

		"Bad set: config has single interpolated value": {
			Schema: map[string]*Schema{
				"security_groups": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					ForceNew: true,
					Elem:     &Schema{Type: TypeString},
				},
			},

			Config: map[string]interface{}{
				"security_groups": "${var.foo}",
			},

			Err: true,
		},

		"Bad, subresource should not allow unknown elements": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"port": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"port":  80,
						"other": "yes",
					},
				},
			},

			Err: true,
		},

		"Bad, subresource should not allow invalid types": {
			Schema: map[string]*Schema{
				"ingress": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"port": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"port": "bad",
					},
				},
			},

			Err: true,
		},

		"Bad, should not allow lists to be assigned to string attributes": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Required: true,
				},
			},

			Config: map[string]interface{}{
				"availability_zone": []interface{}{"foo", "bar", "baz"},
			},

			Err: true,
		},

		"Bad, should not allow maps to be assigned to string attributes": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Required: true,
				},
			},

			Config: map[string]interface{}{
				"availability_zone": map[string]interface{}{"foo": "bar", "baz": "thing"},
			},

			Err: true,
		},

		"Deprecated attribute usage generates warning, but not error": {
			Schema: map[string]*Schema{
				"old_news": {
					Type:       TypeString,
					Optional:   true,
					Deprecated: "please use 'new_news' instead",
				},
			},

			Config: map[string]interface{}{
				"old_news": "extra extra!",
			},

			Err: false,

			Warnings: []string{
				"Warning: Argument is deprecated: please use 'new_news' instead",
			},
		},

		"Deprecated generates no warnings if attr not used": {
			Schema: map[string]*Schema{
				"old_news": {
					Type:       TypeString,
					Optional:   true,
					Deprecated: "please use 'new_news' instead",
				},
			},

			Err: false,

			Warnings: nil,
		},

		"Conflicting attributes generate error": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:     TypeString,
					Optional: true,
				},
				"blacklist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": "white-val",
				"blacklist": "black-val",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf(`Error: Conflicting configuration arguments: "blacklist": conflicts with whitelist`),
			},
		},

		"Conflicting attributes okay when unknown 1": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:     TypeString,
					Optional: true,
				},
				"blacklist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": "white-val",
				"blacklist": hcl2shim.UnknownVariableValue,
			},

			Err: false,
		},

		"Conflicting list attributes okay when unknown 1": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:     TypeList,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
				"blacklist": {
					Type:          TypeList,
					Optional:      true,
					Elem:          &Schema{Type: TypeString},
					ConflictsWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": []interface{}{"white-val"},
				"blacklist": []interface{}{hcl2shim.UnknownVariableValue},
			},

			Err: false,
		},

		"Conflicting attributes okay when unknown 2": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:     TypeString,
					Optional: true,
				},
				"blacklist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
				"blacklist": "black-val",
			},

			Err: false,
		},

		"Conflicting attributes generate error even if one is unknown": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"blacklist", "greenlist"},
				},
				"blacklist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"whitelist", "greenlist"},
				},
				"greenlist": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
				"blacklist": "black-val",
				"greenlist": "green-val",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf(`Error: Conflicting configuration arguments: "blacklist": conflicts with greenlist`),
				fmt.Errorf(`Error: Conflicting configuration arguments: "greenlist": conflicts with blacklist`),
			},
		},

		"Required attribute & undefined conflicting optional are good": {
			Schema: map[string]*Schema{
				"required_att": {
					Type:     TypeString,
					Required: true,
				},
				"optional_att": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"required_att"},
				},
			},

			Config: map[string]interface{}{
				"required_att": "required-val",
			},

			Err: false,
		},

		"Required conflicting attribute & defined optional generate error": {
			Schema: map[string]*Schema{
				"required_att": {
					Type:     TypeString,
					Required: true,
				},
				"optional_att": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"required_att"},
				},
			},

			Config: map[string]interface{}{
				"required_att": "required-val",
				"optional_att": "optional-val",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf(`Error: Conflicting configuration arguments: "optional_att": conflicts with required_att`),
			},
		},

		"Computed + Optional fields conflicting with each other": {
			Schema: map[string]*Schema{
				"foo_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"bar_att"},
				},
				"bar_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"foo_att"},
				},
			},

			Config: map[string]interface{}{
				"foo_att": "foo-val",
				"bar_att": "bar-val",
			},

			Err: true,
			Errors: []error{
				fmt.Errorf(`Error: Conflicting configuration arguments: "bar_att": conflicts with foo_att`),
				fmt.Errorf(`Error: Conflicting configuration arguments: "foo_att": conflicts with bar_att`),
			},
		},

		"Computed + Optional fields NOT conflicting with each other": {
			Schema: map[string]*Schema{
				"foo_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"bar_att"},
				},
				"bar_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"foo_att"},
				},
			},

			Config: map[string]interface{}{
				"foo_att": "foo-val",
			},

			Err: false,
		},

		"Computed + Optional fields that conflict with none set": {
			Schema: map[string]*Schema{
				"foo_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"bar_att"},
				},
				"bar_att": {
					Type:          TypeString,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"foo_att"},
				},
			},

			Config: map[string]interface{}{},

			Err: false,
		},

		"Good with ValidateFunc": {
			Schema: map[string]*Schema{
				"validate_me": {
					Type:     TypeString,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						return
					},
				},
			},
			Config: map[string]interface{}{
				"validate_me": "valid",
			},
			Err: false,
		},

		"Bad with ValidateFunc": {
			Schema: map[string]*Schema{
				"validate_me": {
					Type:     TypeString,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						es = append(es, fmt.Errorf("something is not right here"))
						return
					},
				},
			},
			Config: map[string]interface{}{
				"validate_me": "invalid",
			},
			Err: true,
			Errors: []error{
				fmt.Errorf(`Error: something is not right here`),
			},
		},

		"ValidateFunc not called when type does not match": {
			Schema: map[string]*Schema{
				"number": {
					Type:     TypeInt,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						t.Fatalf("Should not have gotten validate call")
						return
					},
				},
			},
			Config: map[string]interface{}{
				"number": "NaN",
			},
			Err: true,
		},

		"ValidateFunc gets decoded type": {
			Schema: map[string]*Schema{
				"maybe": {
					Type:     TypeBool,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						if _, ok := v.(bool); !ok {
							t.Fatalf("Expected bool, got: %#v", v)
						}
						return
					},
				},
			},
			Config: map[string]interface{}{
				"maybe": "true",
			},
		},

		"ValidateFunc is not called with a computed value": {
			Schema: map[string]*Schema{
				"validate_me": {
					Type:     TypeString,
					Required: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						es = append(es, fmt.Errorf("something is not right here"))
						return
					},
				},
			},
			Config: map[string]interface{}{
				"validate_me": hcl2shim.UnknownVariableValue,
			},

			Err: false,
		},

		"special timeouts field": {
			Schema: map[string]*Schema{
				"availability_zone": {
					Type:     TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},

			Config: map[string]interface{}{
				TimeoutsConfigKey: "bar",
			},

			Err: false,
		},

		"invalid bool field": {
			Schema: map[string]*Schema{
				"bool_field": {
					Type:     TypeBool,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"bool_field": "abcdef",
			},
			Err: true,
		},
		"invalid integer field": {
			Schema: map[string]*Schema{
				"integer_field": {
					Type:     TypeInt,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"integer_field": "abcdef",
			},
			Err: true,
		},
		"invalid float field": {
			Schema: map[string]*Schema{
				"float_field": {
					Type:     TypeFloat,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"float_field": "abcdef",
			},
			Err: true,
		},

		// Invalid map values
		"invalid bool map value": {
			Schema: map[string]*Schema{
				"boolMap": {
					Type:     TypeMap,
					Elem:     TypeBool,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"boolMap": map[string]interface{}{
					"boolField": "notbool",
				},
			},
			Err: true,
		},
		"invalid int map value": {
			Schema: map[string]*Schema{
				"intMap": {
					Type:     TypeMap,
					Elem:     TypeInt,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"intMap": map[string]interface{}{
					"intField": "notInt",
				},
			},
			Err: true,
		},
		"invalid float map value": {
			Schema: map[string]*Schema{
				"floatMap": {
					Type:     TypeMap,
					Elem:     TypeFloat,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"floatMap": map[string]interface{}{
					"floatField": "notFloat",
				},
			},
			Err: true,
		},

		"map with positive validate function": {
			Schema: map[string]*Schema{
				"floatInt": {
					Type:     TypeMap,
					Elem:     TypeInt,
					Optional: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						return
					},
				},
			},
			Config: map[string]interface{}{
				"floatInt": map[string]interface{}{
					"rightAnswer": "42",
					"tooMuch":     "43",
				},
			},
			Err: false,
		},
		"map with negative validate function": {
			Schema: map[string]*Schema{
				"floatInt": {
					Type:     TypeMap,
					Elem:     TypeInt,
					Optional: true,
					ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
						es = append(es, fmt.Errorf("this is not fine"))
						return
					},
				},
			},
			Config: map[string]interface{}{
				"floatInt": map[string]interface{}{
					"rightAnswer": "42",
					"tooMuch":     "43",
				},
			},
			Err: true,
		},

		// The Validation function should not see interpolation strings from
		// non-computed values.
		"set with partially computed list and map": {
			Schema: map[string]*Schema{
				"outer": {
					Type:     TypeSet,
					Optional: true,
					Computed: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"list": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
									ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
										if strings.HasPrefix(v.(string), "${") {
											es = append(es, fmt.Errorf("should not have interpolations"))
										}
										return
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"outer": []interface{}{
					map[string]interface{}{
						"list": []interface{}{"A", hcl2shim.UnknownVariableValue, "c"},
					},
				},
			},
			Err: false,
		},
		"unexpected nils values": {
			Schema: map[string]*Schema{
				"strings": {
					Type:     TypeList,
					Optional: true,
					Elem: &Schema{
						Type: TypeString,
					},
				},
				"block": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"int": {
								Type:     TypeInt,
								Required: true,
							},
						},
					},
				},
			},

			Config: map[string]interface{}{
				"strings": []interface{}{"1", nil},
				"block": []interface{}{map[string]interface{}{
					"int": nil,
				},
					nil,
				},
			},
			Err: true,
		},
		"bool with DefaultFunc that returns str": {
			Schema: map[string]*Schema{
				"bool_field": {
					Type:        TypeBool,
					Optional:    true,
					DefaultFunc: func() (interface{}, error) { return "true", nil },
				},
			},
		},
		"float with DefaultFunc that returns str": {
			Schema: map[string]*Schema{
				"float_field": {
					Type:        TypeFloat,
					Optional:    true,
					DefaultFunc: func() (interface{}, error) { return "1.23", nil },
				},
			},
		},
		"int with DefaultFunc that returns str": {
			Schema: map[string]*Schema{
				"int_field": {
					Type:        TypeInt,
					Optional:    true,
					DefaultFunc: func() (interface{}, error) { return "1", nil },
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)

			diags := schemaMap(tc.Schema).Validate(c)
			if diags.HasError() != tc.Err {
				if !diags.HasError() {
					t.Errorf("%q: no errors", tn)
				}

				for _, e := range diagutils.ErrorDiags(diags).Errors() {
					t.Errorf("%q: err: %s", tn, e)
				}

				t.FailNow()
			}

			ws := diagutils.WarningDiags(diags).Warnings()
			if !reflect.DeepEqual(ws, tc.Warnings) {
				t.Fatalf("%q: warnings:\n\ngot:  %#v\nwant: %#v", tn, ws, tc.Warnings)
			}

			es := diagutils.ErrorDiags(diags).Errors()
			if tc.Errors != nil {
				sort.Sort(errorSort(es))
				sort.Sort(errorSort(tc.Errors))

				if !errorEquals(es, tc.Errors) {
					t.Fatalf("%q: errors:\n\ngot:  %q\nwant: %q", tn, es, tc.Errors)
				}
			}
		})

	}
}

func errorEquals(a []error, b []error) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].Error() != b[i].Error() {
			return false
		}
	}
	return true
}

func TestSchemaSet_ValidateMaxItems(t *testing.T) {
	cases := map[string]struct {
		Schema          map[string]*Schema
		State           *terraform.InstanceState
		Config          map[string]interface{}
		ConfigVariables map[string]string
		Diff            *terraform.InstanceDiff
		Err             bool
		Errors          []error
	}{
		"#0": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					MaxItems: 1,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo", "bar"},
			},
			Diff: nil,
			Err:  true,
			Errors: []error{
				fmt.Errorf("Error: Too many list items: Attribute aliases supports 1 item maximum, but config has 2 declared."),
			},
		},
		"#1": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo", "bar"},
			},
			Diff:   nil,
			Err:    false,
			Errors: nil,
		},
		"#2": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					MaxItems: 1,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo"},
			},
			Diff:   nil,
			Err:    false,
			Errors: nil,
		},
		"#3": {
			Schema: map[string]*Schema{
				"service_account": {
					Type:     TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"aliases": {
								Type:     TypeSet,
								Optional: true,
								MinItems: 2,
								Elem:     &Schema{Type: TypeString},
							},
						},
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"service_account": []interface{}{
					map[string]interface{}{
						"aliases": []interface{}{"foo"},
					},
				},
			},
			Diff: nil,
			Err:  true,
			Errors: []error{
				fmt.Errorf("Error: Not enough list items: Attribute service_account.0.aliases requires 2 item minimum, but config has only 1 declared."),
			},
		},
	}

	for tn, tc := range cases {
		c := terraform.NewResourceConfigRaw(tc.Config)
		diags := schemaMap(tc.Schema).Validate(c)

		if diags.HasError() != tc.Err {
			if !diags.HasError() {
				t.Errorf("%q: no errors", tn)
			}

			for _, e := range diagutils.ErrorDiags(diags).Errors() {
				t.Errorf("%q: err: %s", tn, e)
			}

			t.FailNow()
		}

		es := diagutils.ErrorDiags(diags).Errors()
		if tc.Errors != nil {
			if !errorEquals(es, tc.Errors) {
				t.Fatalf("%q: expected: %q\ngot: %q", tn, tc.Errors, es)
			}
		}
	}
}

func TestSchemaSet_ValidateMinItems(t *testing.T) {
	cases := map[string]struct {
		Schema          map[string]*Schema
		State           *terraform.InstanceState
		Config          map[string]interface{}
		ConfigVariables map[string]string
		Diff            *terraform.InstanceDiff
		Err             bool
		Errors          []error
	}{
		"#0": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					MinItems: 2,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo", "bar"},
			},
			Diff:   nil,
			Err:    false,
			Errors: nil,
		},
		"#1": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo", "bar"},
			},
			Diff:   nil,
			Err:    false,
			Errors: nil,
		},
		"#2": {
			Schema: map[string]*Schema{
				"aliases": {
					Type:     TypeSet,
					Optional: true,
					MinItems: 2,
					Elem:     &Schema{Type: TypeString},
				},
			},
			State: nil,
			Config: map[string]interface{}{
				"aliases": []interface{}{"foo"},
			},
			Diff: nil,
			Err:  true,
			Errors: []error{
				fmt.Errorf("Error: Not enough list items: Attribute aliases requires 2 item minimum, but config has only 1 declared."),
			},
		},
		"#3": {
			Schema: map[string]*Schema{
				"service_account": {
					Type:     TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"aliases": {
								Type:     TypeSet,
								Optional: true,
								MinItems: 2,
								Elem:     &Schema{Type: TypeString},
							},
						},
					},
				},
			},

			State: nil,

			Config: map[string]interface{}{
				"service_account": []interface{}{
					map[string]interface{}{
						"aliases": []interface{}{"foo"},
					},
				},
			},
			Diff: nil,
			Err:  true,
			Errors: []error{
				fmt.Errorf("Error: Not enough list items: Attribute service_account.0.aliases requires 2 item minimum, but config has only 1 declared."),
			},
		},
	}

	for tn, tc := range cases {
		c := terraform.NewResourceConfigRaw(tc.Config)
		diags := schemaMap(tc.Schema).Validate(c)

		if diags.HasError() != tc.Err {
			if !diags.HasError() {
				t.Errorf("%q: no errors", tn)
			}

			for _, e := range diagutils.ErrorDiags(diags).Errors() {
				t.Errorf("%q: err: %s", tn, e)
			}

			t.FailNow()
		}

		es := diagutils.ErrorDiags(diags).Errors()
		if tc.Errors != nil {
			if !errorEquals(es, tc.Errors) {
				t.Fatalf("%q: wrong errors\ngot:  %q\nwant: %q", tn, es, tc.Errors)
			}
		}
	}
}

// errorSort implements sort.Interface to sort errors by their error message
type errorSort []error

func (e errorSort) Len() int      { return len(e) }
func (e errorSort) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e errorSort) Less(i, j int) bool {
	return e[i].Error() < e[j].Error()
}

func TestSchemaMapDeepCopy(t *testing.T) {
	schema := map[string]*Schema{
		"foo": {
			Type: TypeString,
		},
	}
	source := schemaMap(schema)
	dest := source.DeepCopy()
	dest["foo"].ForceNew = true
	if reflect.DeepEqual(source, dest) {
		t.Fatalf("source and dest should not match")
	}
}

func TestValidateConflictingAttributes(t *testing.T) {
	cases := map[string]struct {
		Key    string
		Schema *Schema
		Config map[string]interface{}
		Err    bool
	}{
		"root attribute self conflicting": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"self"},
			},
			Config: map[string]interface{}{
				"self": true,
			},
			Err: true,
		},

		"root attribute conflicting unconfigured self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{},
			Err:    false,
		},

		"root attribute conflicting unconfigured self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"self": hcl2shim.UnknownVariableValue,
			},
			Err: false,
		},

		"root attribute conflicting unconfigured self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"self": true,
			},
			Err: false,
		},

		"root attribute conflicting unknown self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": hcl2shim.UnknownVariableValue,
			},
			Err: false,
		},

		"root attribute conflicting unknown self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": hcl2shim.UnknownVariableValue,
				"self":      hcl2shim.UnknownVariableValue,
			},
			Err: false,
		},

		"root attribute conflicting unknown self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": hcl2shim.UnknownVariableValue,
				"self":      true,
			},
			Err: false,
		},

		"root attribute conflicting known self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": true,
			},
			Err: true,
		},

		"root attribute conflicting known self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": true,
				"self":      hcl2shim.UnknownVariableValue,
			},
			Err: true,
		},

		"root attribute conflicting known self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"root_attr"},
			},
			Config: map[string]interface{}{
				"root_attr": true,
				"self":      true,
			},
			Err: true,
		},

		"configuration block attribute list index syntax self conflicting": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.self"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"self": true,
					},
				},
			},
			Err: true,
		},

		"configuration block attribute list index syntax conflicting unconfigured self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{},
			Err:    false,
		},

		"configuration block attribute list index syntax conflicting unconfigured self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"self": hcl2shim.UnknownVariableValue,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute list index syntax conflicting unconfigured self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"self": true,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute list index syntax conflicting unknown self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": hcl2shim.UnknownVariableValue,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute list index syntax conflicting unknown self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": hcl2shim.UnknownVariableValue,
						"self":        hcl2shim.UnknownVariableValue,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute list index syntax conflicting unknown self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": hcl2shim.UnknownVariableValue,
						"self":        true,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute list index syntax conflicting known self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
					},
				},
			},
			Err: true,
		},

		"configuration block attribute list index syntax conflicting known self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
						"self":        hcl2shim.UnknownVariableValue,
					},
				},
			},
			Err: true,
		},

		"configuration block attribute list index syntax conflicting known self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.0.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
						"self":        true,
					},
				},
			},
			Err: true,
		},

		"configuration block attribute map key syntax self conflicting": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.self"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"self": true,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute map key syntax conflicting known self unconfigured": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute map key syntax conflicting known self unknown": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
						"self":        hcl2shim.UnknownVariableValue,
					},
				},
			},
			Err: false,
		},

		"configuration block attribute map key syntax conflicting known self known": {
			Key: "self",
			Schema: &Schema{
				Type:          TypeBool,
				Optional:      true,
				ConflictsWith: []string{"config_block_attr.nested_attr"},
			},
			Config: map[string]interface{}{
				"config_block_attr": []interface{}{
					map[string]interface{}{
						"nested_attr": true,
						"self":        true,
					},
				},
			},
			Err: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)

			err := validateConflictingAttributes(tc.Key, tc.Schema, c)
			if err == nil && tc.Err {
				t.Fatalf("expected error")
			}

			if err != nil && !tc.Err {
				t.Fatalf("didn't expect error, got error: %+v", err)
			}
		})
	}

}

func TestValidateExactlyOneOfAttributes(t *testing.T) {
	cases := map[string]struct {
		Key    string
		Schema map[string]*Schema
		Config map[string]interface{}
		Err    bool
	}{

		"two attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
				"blacklist": true,
			},
			Err: true,
		},

		"one attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
			},
			Err: false,
		},

		"two attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist":  true,
				"purplelist": true,
			},
			Err: true,
		},

		"one attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": true,
			},
			Err: false,
		},

		"no attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{},
			Err:    true,
		},

		"Only Unknown Variable Value": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": hcl2shim.UnknownVariableValue,
			},
			Err: false,
		},

		"Unknown Variable Value and Known Value": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": hcl2shim.UnknownVariableValue,
				"whitelist":  true,
			},
			Err: false,
		},

		"Unknown Variable Value and 2 Known Value": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					ExactlyOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": hcl2shim.UnknownVariableValue,
				"whitelist":  true,
				"blacklist":  true,
			},
			Err: true,
		},

		"unknown list values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": hcl2shim.UnknownVariableValue,
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": hcl2shim.UnknownVariableValue,
				}},
				"purplelist": "blah",
			},
			Err: false,
		},

		// This should probably fail, but we let it pass and rely on 2nd
		// validation phase when unknowns become known, which will then fail.
		"partially known list values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": "TCP",
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": "TCP",
				}},
				"purplelist": "blah",
			},
			Err: false,
		},

		"known list values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    []interface{}{"80"},
					"protocol": "TCP",
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    []interface{}{"80"},
					"protocol": "TCP",
				}},
			},
			Err: true,
		},

		"wholly unknown set values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": hcl2shim.UnknownVariableValue,
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": hcl2shim.UnknownVariableValue,
				}},
				"purplelist": "blah",
			},
			Err: false,
		},

		// This should probably fail, but we let it pass and rely on 2nd
		// validation phase when unknowns become known, which will then fail.
		"partially known set values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": "TCP",
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    hcl2shim.UnknownVariableValue,
					"protocol": "UDP",
				}},
			},
			Err: false,
		},

		"known set values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"protocol": {
								Type:     TypeString,
								Required: true,
							},
							"ports": {
								Type:     TypeList,
								Optional: true,
								Elem: &Schema{
									Type: TypeString,
								},
							},
						},
					},
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{map[string]interface{}{
					"ports":    []interface{}{"80"},
					"protocol": "TCP",
				}},
				"deny": []interface{}{map[string]interface{}{
					"ports":    []interface{}{"80"},
					"protocol": "TCP",
				}},
			},
			Err: true,
		},

		"wholly unknown simple lists": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{
					hcl2shim.UnknownVariableValue,
					hcl2shim.UnknownVariableValue,
				},
				"deny": []interface{}{
					hcl2shim.UnknownVariableValue,
					hcl2shim.UnknownVariableValue,
				},
				"purplelist": "blah",
			},
			Err: false,
		},

		// This should probably fail, but we let it pass and rely on 2nd
		// validation phase when unknowns become known, which will then fail.
		"partially known simple lists": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{
					hcl2shim.UnknownVariableValue,
					"known",
				},
				"deny": []interface{}{
					hcl2shim.UnknownVariableValue,
					"known",
				},
			},
			Err: false,
		},

		"known simple lists": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": []interface{}{
					"blah",
					"known",
				},
				"deny": []interface{}{
					"known",
				},
			},
			Err: true,
		},

		"wholly unknown map keys and values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeMap,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": map[string]interface{}{
					hcl2shim.UnknownVariableValue: hcl2shim.UnknownVariableValue,
				},
				"deny": map[string]interface{}{
					hcl2shim.UnknownVariableValue: hcl2shim.UnknownVariableValue,
				},
				"purplelist": "blah",
			},
			Err: false,
		},

		// This should probably fail, but we let it pass and rely on 2nd
		// validation phase when unknowns become known, which will then fail.
		"wholly unknown map values": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeMap,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": map[string]interface{}{
					"key": hcl2shim.UnknownVariableValue,
				},
				"deny": map[string]interface{}{
					"key": hcl2shim.UnknownVariableValue,
				},
				"purplelist": "blah",
			},
			Err: false,
		},

		// This should probably fail, but we let it pass and rely on 2nd
		// validation phase when unknowns become known, which will then fail.
		"partially known maps": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeMap,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"purplelist": {
					Type:     TypeString,
					Optional: true,
				},
			},
			Config: map[string]interface{}{
				"allow": map[string]interface{}{
					"first":  "value",
					"second": hcl2shim.UnknownVariableValue,
				},
				"deny": map[string]interface{}{
					"first":  "value",
					"second": hcl2shim.UnknownVariableValue,
				},
				"purplelist": "blah",
			},
			Err: false,
		},

		"known maps": {
			Key: "allow",
			Schema: map[string]*Schema{
				"allow": {
					Type:         TypeMap,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
				"deny": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"allow", "deny"},
				},
			},
			Config: map[string]interface{}{
				"allow": map[string]interface{}{
					"first":  "value",
					"second": "blah",
				},
				"deny": map[string]interface{}{
					"first":  "value",
					"second": "boo",
				},
			},
			Err: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)

			err := validateExactlyOneAttribute(tc.Key, tc.Schema[tc.Key], c)
			if err == nil && tc.Err {
				t.Fatalf("expected error")
			}

			if err != nil && !tc.Err {
				t.Fatalf("didn't expect error, got error: %+v", err)
			}
		})
	}

}

func TestValidateAtLeastOneOfAttributes(t *testing.T) {
	cases := map[string]struct {
		Key    string
		Schema map[string]*Schema
		Config map[string]interface{}
		Err    bool
	}{

		"two attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
				"blacklist": true,
			},
			Err: false,
		},

		"one attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
			},
			Err: false,
		},

		"two attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist":  true,
				"purplelist": true,
			},
			Err: false,
		},

		"three attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist":  true,
				"purplelist": true,
				"blacklist":  true,
			},
			Err: false,
		},

		"one attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": true,
			},
			Err: false,
		},

		"no attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{},
			Err:    true,
		},

		"Only Unknown Variable Value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
			},

			Err: false,
		},

		"only unknown list value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					AtLeastOneOf: []string{"whitelist", "blacklist"},
				},
				"blacklist": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					AtLeastOneOf: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": []interface{}{hcl2shim.UnknownVariableValue},
			},

			Err: false,
		},

		"Unknown Variable Value and Known Value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					AtLeastOneOf: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
				"blacklist": true,
			},

			Err: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)
			diags := schemaMap(tc.Schema).Validate(c)
			if diags.HasError() != tc.Err {
				if !diags.HasError() {
					t.Fatalf("expected error")
				}

				for _, e := range diagutils.ErrorDiags(diags).Errors() {
					t.Fatalf("didn't expect error, got error: %+v", e)
				}

				t.FailNow()
			}
		})
	}
}

func TestPanicOnErrorDefaultsFalse(t *testing.T) {
	t.Setenv("TF_ACC", "")

	if schemaMap(nil).panicOnError() {
		t.Fatalf("panicOnError should be false when TF_ACC is empty")
	}
}

func TestPanicOnErrorTF_ACCSet(t *testing.T) {
	t.Setenv("TF_ACC", "1")

	if !schemaMap(nil).panicOnError() {
		t.Fatalf("panicOnError should be true when TF_ACC is not empty")
	}
}

func TestValidateRequiredWithAttributes(t *testing.T) {
	cases := map[string]struct {
		Key    string
		Schema map[string]*Schema
		Config map[string]interface{}
		Err    bool
	}{

		"two attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
				"blacklist": true,
			},
			Err: false,
		},

		"one attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": true,
			},
			Err: true,
		},

		"no attributes specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"blacklist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{},
			Err:    false,
		},

		"two attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist":  true,
				"purplelist": true,
			},
			Err: false,
		},

		"three attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist":  true,
				"purplelist": true,
				"blacklist":  true,
			},
			Err: false,
		},

		"one attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"purplelist": true,
			},
			Err: true,
		},

		"no attributes of three specified": {
			Key: "whitelist",
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{},
			Err:    false,
		},

		"Only Unknown Variable Value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
			},

			Err: true,
		},

		"only unknown list value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					RequiredWith: []string{"whitelist", "blacklist"},
				},
				"blacklist": {
					Type:         TypeList,
					Optional:     true,
					Elem:         &Schema{Type: TypeString},
					RequiredWith: []string{"whitelist", "blacklist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": []interface{}{hcl2shim.UnknownVariableValue},
			},

			Err: true,
		},

		"Unknown Variable Value and Known Value": {
			Schema: map[string]*Schema{
				"whitelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"blacklist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
				"purplelist": {
					Type:         TypeBool,
					Optional:     true,
					RequiredWith: []string{"whitelist", "blacklist", "purplelist"},
				},
			},

			Config: map[string]interface{}{
				"whitelist": hcl2shim.UnknownVariableValue,
				"blacklist": true,
			},

			Err: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			c := terraform.NewResourceConfigRaw(tc.Config)
			diags := schemaMap(tc.Schema).Validate(c)
			es := diagutils.ErrorDiags(diags).Errors()
			if len(es) > 0 != tc.Err {
				if len(es) == 0 {
					t.Fatalf("expected error")
				}

				for _, e := range es {
					t.Fatalf("didn't expect error, got error: %+v", e)
				}

				t.FailNow()
			}
		})
	}
}
