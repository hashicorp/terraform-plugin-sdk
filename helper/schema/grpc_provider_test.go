package schema

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/msgpack"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugin/convert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The GRPCProviderServer will directly implement the go protobuf server
var _ tfprotov5.ProviderServer = (*GRPCProviderServer)(nil)

func TestUpgradeState_jsonState(t *testing.T) {
	r := &Resource{
		SchemaVersion: 2,
		Schema: map[string]*Schema{
			"two": {
				Type:     TypeInt,
				Optional: true,
			},
		},
	}

	r.StateUpgraders = []StateUpgrader{
		{
			Version: 0,
			Type: cty.Object(map[string]cty.Type{
				"id":   cty.String,
				"zero": cty.Number,
			}),
			Upgrade: func(ctx context.Context, m map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
				_, ok := m["zero"].(float64)
				if !ok {
					return nil, fmt.Errorf("zero not found in %#v", m)
				}
				m["one"] = float64(1)
				delete(m, "zero")
				return m, nil
			},
		},
		{
			Version: 1,
			Type: cty.Object(map[string]cty.Type{
				"id":  cty.String,
				"one": cty.Number,
			}),
			Upgrade: func(ctx context.Context, m map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
				_, ok := m["one"].(float64)
				if !ok {
					return nil, fmt.Errorf("one not found in %#v", m)
				}
				m["two"] = float64(2)
				delete(m, "one")
				return m, nil
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	req := &tfprotov5.UpgradeResourceStateRequest{
		TypeName: "test",
		Version:  0,
		RawState: &tfprotov5.RawState{
			JSON: []byte(`{"id":"bar","zero":0}`),
		},
	}

	resp, err := server.UpgradeResourceState(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Diagnostics) > 0 {
		for _, d := range resp.Diagnostics {
			t.Errorf("%#v", d)
		}
		t.Fatal("error")
	}

	val, err := msgpack.Unmarshal(resp.UpgradedState.MsgPack, r.CoreConfigSchema().ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	expected := cty.ObjectVal(map[string]cty.Value{
		"id":  cty.StringVal("bar"),
		"two": cty.NumberIntVal(2),
	})

	if !cmp.Equal(expected, val, valueComparer, equateEmpty) {
		t.Fatal(cmp.Diff(expected, val, valueComparer, equateEmpty))
	}
}

func TestUpgradeState_jsonStateBigInt(t *testing.T) {
	r := &Resource{
		UseJSONNumber: true,
		SchemaVersion: 2,
		Schema: map[string]*Schema{
			"int": {
				Type:     TypeInt,
				Required: true,
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	req := &tfprotov5.UpgradeResourceStateRequest{
		TypeName: "test",
		Version:  0,
		RawState: &tfprotov5.RawState{
			JSON: []byte(`{"id":"bar","int":7227701560655103598}`),
		},
	}

	resp, err := server.UpgradeResourceState(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Diagnostics) > 0 {
		for _, d := range resp.Diagnostics {
			t.Errorf("%#v", d)
		}
		t.Fatal("error")
	}

	val, err := msgpack.Unmarshal(resp.UpgradedState.MsgPack, r.CoreConfigSchema().ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	expected := cty.ObjectVal(map[string]cty.Value{
		"id":  cty.StringVal("bar"),
		"int": cty.NumberIntVal(7227701560655103598),
	})

	if !cmp.Equal(expected, val, valueComparer, equateEmpty) {
		t.Fatal(cmp.Diff(expected, val, valueComparer, equateEmpty))
	}
}

func TestUpgradeState_removedAttr(t *testing.T) {
	r1 := &Resource{
		Schema: map[string]*Schema{
			"two": {
				Type:     TypeString,
				Optional: true,
			},
		},
	}

	r2 := &Resource{
		Schema: map[string]*Schema{
			"multi": {
				Type:     TypeSet,
				Optional: true,
				Elem: &Resource{
					Schema: map[string]*Schema{
						"set": {
							Type:     TypeSet,
							Optional: true,
							Elem: &Resource{
								Schema: map[string]*Schema{
									"required": {
										Type:     TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	r3 := &Resource{
		Schema: map[string]*Schema{
			"config_mode_attr": {
				Type:       TypeList,
				ConfigMode: SchemaConfigModeAttr,
				Optional:   true,
				Elem: &Resource{
					Schema: map[string]*Schema{
						"foo": {
							Type:     TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}

	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"r1": r1,
			"r2": r2,
			"r3": r3,
		},
	}

	server := NewGRPCProviderServer(p)

	for _, tc := range []struct {
		name     string
		raw      string
		expected cty.Value
	}{
		{
			name: "r1",
			raw:  `{"id":"bar","removed":"removed","two":"2"}`,
			expected: cty.ObjectVal(map[string]cty.Value{
				"id":  cty.StringVal("bar"),
				"two": cty.StringVal("2"),
			}),
		},
		{
			name: "r2",
			raw:  `{"id":"bar","multi":[{"set":[{"required":"ok","removed":"removed"}]}]}`,
			expected: cty.ObjectVal(map[string]cty.Value{
				"id": cty.StringVal("bar"),
				"multi": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"set": cty.SetVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"required": cty.StringVal("ok"),
							}),
						}),
					}),
				}),
			}),
		},
		{
			name: "r3",
			raw:  `{"id":"bar","config_mode_attr":[{"foo":"ok","removed":"removed"}]}`,
			expected: cty.ObjectVal(map[string]cty.Value{
				"id": cty.StringVal("bar"),
				"config_mode_attr": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"foo": cty.StringVal("ok"),
					}),
				}),
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := &tfprotov5.UpgradeResourceStateRequest{
				TypeName: tc.name,
				Version:  0,
				RawState: &tfprotov5.RawState{
					JSON: []byte(tc.raw),
				},
			}
			resp, err := server.UpgradeResourceState(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}

			if len(resp.Diagnostics) > 0 {
				for _, d := range resp.Diagnostics {
					t.Errorf("%#v", d)
				}
				t.Fatal("error")
			}
			val, err := msgpack.Unmarshal(resp.UpgradedState.MsgPack, p.ResourcesMap[tc.name].CoreConfigSchema().ImpliedType())
			if err != nil {
				t.Fatal(err)
			}
			if !tc.expected.RawEquals(val) {
				t.Fatalf("\nexpected: %#v\ngot:      %#v\n", tc.expected, val)
			}
		})
	}

}

func TestUpgradeState_flatmapState(t *testing.T) {
	r := &Resource{
		SchemaVersion: 4,
		Schema: map[string]*Schema{
			"four": {
				Type:     TypeInt,
				Required: true,
			},
			"block": {
				Type:     TypeList,
				Optional: true,
				Elem: &Resource{
					Schema: map[string]*Schema{
						"attr": {
							Type:     TypeString,
							Optional: true,
						},
					},
				},
			},
		},
		// this MigrateState will take the state to version 2
		MigrateState: func(v int, is *terraform.InstanceState, _ interface{}) (*terraform.InstanceState, error) {
			switch v {
			case 0:
				_, ok := is.Attributes["zero"]
				if !ok {
					return nil, fmt.Errorf("zero not found in %#v", is.Attributes)
				}
				is.Attributes["one"] = "1"
				delete(is.Attributes, "zero")
				fallthrough
			case 1:
				_, ok := is.Attributes["one"]
				if !ok {
					return nil, fmt.Errorf("one not found in %#v", is.Attributes)
				}
				is.Attributes["two"] = "2"
				delete(is.Attributes, "one")
			default:
				return nil, fmt.Errorf("invalid schema version %d", v)
			}
			return is, nil
		},
	}

	r.StateUpgraders = []StateUpgrader{
		{
			Version: 2,
			Type: cty.Object(map[string]cty.Type{
				"id":  cty.String,
				"two": cty.Number,
			}),
			Upgrade: func(ctx context.Context, m map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
				_, ok := m["two"].(float64)
				if !ok {
					return nil, fmt.Errorf("two not found in %#v", m)
				}
				m["three"] = float64(3)
				delete(m, "two")
				return m, nil
			},
		},
		{
			Version: 3,
			Type: cty.Object(map[string]cty.Type{
				"id":    cty.String,
				"three": cty.Number,
			}),
			Upgrade: func(ctx context.Context, m map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
				_, ok := m["three"].(float64)
				if !ok {
					return nil, fmt.Errorf("three not found in %#v", m)
				}
				m["four"] = float64(4)
				delete(m, "three")
				return m, nil
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	testReqs := []*tfprotov5.UpgradeResourceStateRequest{
		{
			TypeName: "test",
			Version:  0,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":   "bar",
					"zero": "0",
				},
			},
		},
		{
			TypeName: "test",
			Version:  1,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":  "bar",
					"one": "1",
				},
			},
		},
		// two and  up could be stored in flatmap or json states
		{
			TypeName: "test",
			Version:  2,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":  "bar",
					"two": "2",
				},
			},
		},
		{
			TypeName: "test",
			Version:  2,
			RawState: &tfprotov5.RawState{
				JSON: []byte(`{"id":"bar","two":2}`),
			},
		},
		{
			TypeName: "test",
			Version:  3,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":    "bar",
					"three": "3",
				},
			},
		},
		{
			TypeName: "test",
			Version:  3,
			RawState: &tfprotov5.RawState{
				JSON: []byte(`{"id":"bar","three":3}`),
			},
		},
		{
			TypeName: "test",
			Version:  4,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":   "bar",
					"four": "4",
				},
			},
		},
		{
			TypeName: "test",
			Version:  4,
			RawState: &tfprotov5.RawState{
				JSON: []byte(`{"id":"bar","four":4}`),
			},
		},
	}

	for i, req := range testReqs {
		t.Run(fmt.Sprintf("%d-%d", i, req.Version), func(t *testing.T) {
			resp, err := server.UpgradeResourceState(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}

			if len(resp.Diagnostics) > 0 {
				for _, d := range resp.Diagnostics {
					t.Errorf("%#v", d)
				}
				t.Fatal("error")
			}

			val, err := msgpack.Unmarshal(resp.UpgradedState.MsgPack, r.CoreConfigSchema().ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			expected := cty.ObjectVal(map[string]cty.Value{
				"block": cty.ListValEmpty(cty.Object(map[string]cty.Type{"attr": cty.String})),
				"id":    cty.StringVal("bar"),
				"four":  cty.NumberIntVal(4),
			})

			if !cmp.Equal(expected, val, valueComparer, equateEmpty) {
				t.Fatal(cmp.Diff(expected, val, valueComparer, equateEmpty))
			}
		})
	}
}

func TestUpgradeState_flatmapStateMissingMigrateState(t *testing.T) {
	r := &Resource{
		SchemaVersion: 1,
		Schema: map[string]*Schema{
			"one": {
				Type:     TypeInt,
				Required: true,
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	testReqs := []*tfprotov5.UpgradeResourceStateRequest{
		{
			TypeName: "test",
			Version:  0,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":  "bar",
					"one": "1",
				},
			},
		},
		{
			TypeName: "test",
			Version:  1,
			RawState: &tfprotov5.RawState{
				Flatmap: map[string]string{
					"id":  "bar",
					"one": "1",
				},
			},
		},
		{
			TypeName: "test",
			Version:  1,
			RawState: &tfprotov5.RawState{
				JSON: []byte(`{"id":"bar","one":1}`),
			},
		},
	}

	for i, req := range testReqs {
		t.Run(fmt.Sprintf("%d-%d", i, req.Version), func(t *testing.T) {
			resp, err := server.UpgradeResourceState(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}

			if len(resp.Diagnostics) > 0 {
				for _, d := range resp.Diagnostics {
					t.Errorf("%#v", d)
				}
				t.Fatal("error")
			}

			val, err := msgpack.Unmarshal(resp.UpgradedState.MsgPack, r.CoreConfigSchema().ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			expected := cty.ObjectVal(map[string]cty.Value{
				"id":  cty.StringVal("bar"),
				"one": cty.NumberIntVal(1),
			})

			if !cmp.Equal(expected, val, valueComparer, equateEmpty) {
				t.Fatal(cmp.Diff(expected, val, valueComparer, equateEmpty))
			}
		})
	}
}

func TestPlanResourceChange(t *testing.T) {
	r := &Resource{
		SchemaVersion: 4,
		Schema: map[string]*Schema{
			"foo": {
				Type:     TypeInt,
				Optional: true,
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	schema := r.CoreConfigSchema()
	priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	// A propsed state with only the ID unknown will produce a nil diff, and
	// should return the propsed state value.
	proposedVal, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"id": cty.UnknownVal(cty.String),
	}))
	if err != nil {
		t.Fatal(err)
	}
	proposedState, err := msgpack.Marshal(proposedVal, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"id": cty.NullVal(cty.String),
	}))
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	testReq := &tfprotov5.PlanResourceChangeRequest{
		TypeName: "test",
		PriorState: &tfprotov5.DynamicValue{
			MsgPack: priorState,
		},
		ProposedNewState: &tfprotov5.DynamicValue{
			MsgPack: proposedState,
		},
		Config: &tfprotov5.DynamicValue{
			MsgPack: configBytes,
		},
	}

	resp, err := server.PlanResourceChange(context.Background(), testReq)
	if err != nil {
		t.Fatal(err)
	}

	plannedStateVal, err := msgpack.Unmarshal(resp.PlannedState.MsgPack, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(proposedVal, plannedStateVal, valueComparer) {
		t.Fatal(cmp.Diff(proposedVal, plannedStateVal, valueComparer))
	}
}

func TestPlanResourceChange_bigint(t *testing.T) {
	r := &Resource{
		UseJSONNumber: true,
		Schema: map[string]*Schema{
			"foo": {
				Type:     TypeInt,
				Required: true,
			},
		},
	}

	server := NewGRPCProviderServer(&Provider{
		ResourcesMap: map[string]*Resource{
			"test": r,
		},
	})

	schema := r.CoreConfigSchema()
	priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	proposedVal := cty.ObjectVal(map[string]cty.Value{
		"id":  cty.UnknownVal(cty.String),
		"foo": cty.MustParseNumberVal("7227701560655103598"),
	})
	proposedState, err := msgpack.Marshal(proposedVal, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"id":  cty.NullVal(cty.String),
		"foo": cty.MustParseNumberVal("7227701560655103598"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	testReq := &tfprotov5.PlanResourceChangeRequest{
		TypeName: "test",
		PriorState: &tfprotov5.DynamicValue{
			MsgPack: priorState,
		},
		ProposedNewState: &tfprotov5.DynamicValue{
			MsgPack: proposedState,
		},
		Config: &tfprotov5.DynamicValue{
			MsgPack: configBytes,
		},
	}

	resp, err := server.PlanResourceChange(context.Background(), testReq)
	if err != nil {
		t.Fatal(err)
	}

	plannedStateVal, err := msgpack.Unmarshal(resp.PlannedState.MsgPack, schema.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(proposedVal, plannedStateVal, valueComparer) {
		t.Fatal(cmp.Diff(proposedVal, plannedStateVal, valueComparer))
	}

	plannedStateFoo, acc := plannedStateVal.GetAttr("foo").AsBigFloat().Int64()
	if acc != big.Exact {
		t.Fatalf("Expected exact accuracy, got %s", acc)
	}
	if plannedStateFoo != 7227701560655103598 {
		t.Fatalf("Expected %d, got %d, this represents a loss of precision in planning large numbers", 7227701560655103598, plannedStateFoo)
	}
}

func TestApplyResourceChange(t *testing.T) {
	testCases := []struct {
		Description  string
		TestResource *Resource
	}{
		{
			Description: "Create",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				Create: func(rd *ResourceData, _ interface{}) error {
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateContext",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateContext: func(_ context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateWithoutTimeout",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateWithoutTimeout: func(_ context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "Create_cty",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateWithoutTimeout: func(_ context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					if rd.GetRawConfig().IsNull() {
						return diag.FromErr(errors.New("null raw config"))
					}
					if !rd.GetRawState().IsNull() {
						return diag.FromErr(fmt.Errorf("non-null raw state: %s", rd.GetRawState().GoString()))
					}
					if rd.GetRawPlan().IsNull() {
						return diag.FromErr(errors.New("null raw plan"))
					}
					rd.SetId("bar")
					return nil
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				ResourcesMap: map[string]*Resource{
					"test": testCase.TestResource,
				},
			})

			schema := testCase.TestResource.CoreConfigSchema()
			priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			// A proposed state with only the ID unknown will produce a nil diff, and
			// should return the proposed state value.
			plannedVal, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.UnknownVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			plannedState, err := msgpack.Marshal(plannedVal, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.NullVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test",
				PriorState: &tfprotov5.DynamicValue{
					MsgPack: priorState,
				},
				PlannedState: &tfprotov5.DynamicValue{
					MsgPack: plannedState,
				},
				Config: &tfprotov5.DynamicValue{
					MsgPack: configBytes,
				},
			}

			resp, err := server.ApplyResourceChange(context.Background(), testReq)
			if err != nil {
				t.Fatal(err)
			}

			newStateVal, err := msgpack.Unmarshal(resp.NewState.MsgPack, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			id := newStateVal.GetAttr("id").AsString()
			if id != "bar" {
				t.Fatalf("incorrect final state: %#v\n", newStateVal)
			}
		})
	}
}

func TestApplyResourceChange_bigint(t *testing.T) {
	testCases := []struct {
		Description  string
		TestResource *Resource
	}{
		{
			Description: "Create",
			TestResource: &Resource{
				UseJSONNumber: true,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Required: true,
					},
				},
				Create: func(rd *ResourceData, _ interface{}) error {
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateContext",
			TestResource: &Resource{
				UseJSONNumber: true,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Required: true,
					},
				},
				CreateContext: func(_ context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateWithoutTimeout",
			TestResource: &Resource{
				UseJSONNumber: true,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Required: true,
					},
				},
				CreateWithoutTimeout: func(_ context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					rd.SetId("bar")
					return nil
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				ResourcesMap: map[string]*Resource{
					"test": testCase.TestResource,
				},
			})

			schema := testCase.TestResource.CoreConfigSchema()
			priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			plannedVal := cty.ObjectVal(map[string]cty.Value{
				"id":  cty.UnknownVal(cty.String),
				"foo": cty.MustParseNumberVal("7227701560655103598"),
			})
			plannedState, err := msgpack.Marshal(plannedVal, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id":  cty.NullVal(cty.String),
				"foo": cty.MustParseNumberVal("7227701560655103598"),
			}))
			if err != nil {
				t.Fatal(err)
			}
			configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test",
				PriorState: &tfprotov5.DynamicValue{
					MsgPack: priorState,
				},
				PlannedState: &tfprotov5.DynamicValue{
					MsgPack: plannedState,
				},
				Config: &tfprotov5.DynamicValue{
					MsgPack: configBytes,
				},
			}

			resp, err := server.ApplyResourceChange(context.Background(), testReq)
			if err != nil {
				t.Fatal(err)
			}

			newStateVal, err := msgpack.Unmarshal(resp.NewState.MsgPack, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			id := newStateVal.GetAttr("id").AsString()
			if id != "bar" {
				t.Fatalf("incorrect final state: %#v\n", newStateVal)
			}

			foo, acc := newStateVal.GetAttr("foo").AsBigFloat().Int64()
			if acc != big.Exact {
				t.Fatalf("Expected exact accuracy, got %s", acc)
			}
			if foo != 7227701560655103598 {
				t.Fatalf("Expected %d, got %d, this represents a loss of precision in applying large numbers", 7227701560655103598, foo)
			}
		})
	}
}

func TestReadDataSource(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server   *GRPCProviderServer
		req      *tfprotov5.ReadDataSourceRequest
		expected *tfprotov5.ReadDataSourceResponse
	}{
		"missing-set-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Computed: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.NullVal(cty.Object(map[string]cty.Type{
							"id": cty.String,
						})),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.NullVal(cty.String),
						}),
					),
				},
			},
		},
		"empty": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Computed: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.EmptyObject,
						cty.NullVal(cty.EmptyObject),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
		},
		"null-object": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Computed: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.NullVal(cty.Object(map[string]cty.Type{
							"id": cty.String,
						})),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
		},
		"computed-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Computed: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.NullVal(cty.String),
						}),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
		},
		"optional-computed-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Optional: true,
								Computed: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.NullVal(cty.String),
						}),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
		},
		"optional-no-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"test": {
								Type:     TypeString,
								Optional: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id":   cty.String,
							"test": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id":   cty.NullVal(cty.String),
							"test": cty.NullVal(cty.String),
						}),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id":   cty.String,
							"test": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id":   cty.StringVal("test-id"),
							"test": cty.NullVal(cty.String),
						}),
					),
				},
			},
		},
		"required-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"id": {
								Type:     TypeString,
								Required: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal("test-id"),
						}),
					),
				},
			},
		},
		"required-no-id": {
			server: NewGRPCProviderServer(&Provider{
				DataSourcesMap: map[string]*Resource{
					"test": {
						SchemaVersion: 1,
						Schema: map[string]*Schema{
							"test": {
								Type:     TypeString,
								Required: true,
							},
						},
						ReadContext: func(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
							d.SetId("test-id")
							return nil
						},
					},
				},
			}),
			req: &tfprotov5.ReadDataSourceRequest{
				TypeName: "test",
				Config: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id":   cty.String,
							"test": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id":   cty.NullVal(cty.String),
							"test": cty.StringVal("test-string"),
						}),
					),
				},
			},
			expected: &tfprotov5.ReadDataSourceResponse{
				State: &tfprotov5.DynamicValue{
					MsgPack: mustMsgpackMarshal(
						cty.Object(map[string]cty.Type{
							"id":   cty.String,
							"test": cty.String,
						}),
						cty.ObjectVal(map[string]cty.Value{
							"id":   cty.StringVal("test-id"),
							"test": cty.StringVal("test-string"),
						}),
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp, err := testCase.server.ReadDataSource(context.Background(), testCase.req)

			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(resp, testCase.expected, valueComparer); diff != "" {
				ty := testCase.server.getDatasourceSchemaBlock("test").ImpliedType()

				if resp != nil && resp.State != nil {
					t.Logf("resp.State.MsgPack: %s", mustMsgpackUnmarshal(ty, resp.State.MsgPack))
				}

				if testCase.expected != nil && testCase.expected.State != nil {
					t.Logf("expected: %s", mustMsgpackUnmarshal(ty, testCase.expected.State.MsgPack))
				}

				t.Error(diff)
			}
		})
	}
}

func TestPrepareProviderConfig(t *testing.T) {
	for _, tc := range []struct {
		Name         string
		Schema       map[string]*Schema
		ConfigVal    cty.Value
		ExpectError  string
		ExpectConfig cty.Value
	}{
		{
			Name: "test prepare",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}),
		},
		{
			Name: "test default",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Optional: true,
					Default:  "default",
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("default"),
			}),
		},
		{
			Name: "test defaultfunc",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "defaultfunc", nil
					},
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("defaultfunc"),
			}),
		},
		{
			Name: "test default required",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Required: true,
					DefaultFunc: func() (interface{}, error) {
						return "defaultfunc", nil
					},
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("defaultfunc"),
			}),
		},
		{
			Name: "test incorrect type",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Required: true,
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NumberIntVal(3),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("3"),
			}),
		},
		{
			Name: "test incorrect default type",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeString,
					Optional: true,
					Default:  true,
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.String),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("true"),
			}),
		},
		{
			Name: "test incorrect default bool type",
			Schema: map[string]*Schema{
				"foo": {
					Type:     TypeBool,
					Optional: true,
					Default:  "",
				},
			},
			ConfigVal: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NullVal(cty.Bool),
			}),
			ExpectConfig: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.False,
			}),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				Schema: tc.Schema,
			})

			block := InternalMap(tc.Schema).CoreConfigSchema()

			rawConfig, err := msgpack.Marshal(tc.ConfigVal, block.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.PrepareProviderConfigRequest{
				Config: &tfprotov5.DynamicValue{
					MsgPack: rawConfig,
				},
			}

			resp, err := server.PrepareProviderConfig(context.Background(), testReq)
			if err != nil {
				t.Fatal(err)
			}

			if tc.ExpectError != "" && len(resp.Diagnostics) > 0 {
				for _, d := range resp.Diagnostics {
					if !strings.Contains(d.Summary, tc.ExpectError) {
						t.Fatalf("Unexpected error: %s/%s", d.Summary, d.Detail)
					}
				}
				return
			}

			// we should have no errors past this point
			for _, d := range resp.Diagnostics {
				if d.Severity == tfprotov5.DiagnosticSeverityError {
					t.Fatal(resp.Diagnostics)
				}
			}

			val, err := msgpack.Unmarshal(resp.PreparedConfig.MsgPack, block.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			if tc.ExpectConfig.GoString() != val.GoString() {
				t.Fatalf("\nexpected: %#v\ngot: %#v", tc.ExpectConfig, val)
			}
		})
	}
}

func TestGetSchemaTimeouts(t *testing.T) {
	r := &Resource{
		SchemaVersion: 4,
		Timeouts: &ResourceTimeout{
			Create:  DefaultTimeout(time.Second),
			Read:    DefaultTimeout(2 * time.Second),
			Update:  DefaultTimeout(3 * time.Second),
			Default: DefaultTimeout(10 * time.Second),
		},
		Schema: map[string]*Schema{
			"foo": {
				Type:     TypeInt,
				Optional: true,
			},
		},
	}

	// verify that the timeouts appear in the schema as defined
	block := r.CoreConfigSchema()
	timeoutsBlock := block.BlockTypes["timeouts"]
	if timeoutsBlock == nil {
		t.Fatal("missing timeouts in schema")
	}

	if timeoutsBlock.Attributes["create"] == nil {
		t.Fatal("missing create timeout in schema")
	}
	if timeoutsBlock.Attributes["read"] == nil {
		t.Fatal("missing read timeout in schema")
	}
	if timeoutsBlock.Attributes["update"] == nil {
		t.Fatal("missing update timeout in schema")
	}
	if d := timeoutsBlock.Attributes["delete"]; d != nil {
		t.Fatalf("unexpected delete timeout in schema: %#v", d)
	}
	if timeoutsBlock.Attributes["default"] == nil {
		t.Fatal("missing default timeout in schema")
	}
}

func TestNormalizeNullValues(t *testing.T) {
	for i, tc := range []struct {
		Src, Dst, Expect cty.Value
		Apply            bool
	}{
		{
			// The known set value is copied over the null set value
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"foo": cty.NullVal(cty.String),
					}),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"foo": cty.String,
				}))),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"foo": cty.NullVal(cty.String),
					}),
				}),
			}),
			Apply: true,
		},
		{
			// A zero set value is kept
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.String),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.String),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.String),
			}),
		},
		{
			// The known set value is copied over the null set value
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"foo": cty.NullVal(cty.String),
					}),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"foo": cty.String,
				}))),
			}),
			// If we're only in a plan, we can't compare sets at all
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"foo": cty.String,
				}))),
			}),
		},
		{
			// The empty map is copied over the null map
			Src: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"map": cty.NullVal(cty.Map(cty.String)),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
			}),
			Apply: true,
		},
		{
			// A zero value primitive is copied over a null primitive
			Src: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NullVal(cty.String),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
			Apply: true,
		},
		{
			// Plan primitives are kept
			Src: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NumberIntVal(0),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NullVal(cty.Number),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NullVal(cty.Number),
			}),
		},
		{
			// Neither plan nor apply should remove empty strings
			Src: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NullVal(cty.String),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
		},
		{
			// Neither plan nor apply should remove empty strings
			Src: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"string": cty.NullVal(cty.String),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal(""),
			}),
			Apply: true,
		},
		{
			// The null map is retained, because the src was unknown
			Src: cty.ObjectVal(map[string]cty.Value{
				"map": cty.UnknownVal(cty.Map(cty.String)),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"map": cty.NullVal(cty.Map(cty.String)),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"map": cty.NullVal(cty.Map(cty.String)),
			}),
			Apply: true,
		},
		{
			// the nul set is retained, because the src set contains an unknown value
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"foo": cty.UnknownVal(cty.String),
					}),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"foo": cty.String,
				}))),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"foo": cty.String,
				}))),
			}),
			Apply: true,
		},
		{
			// Retain don't re-add unexpected planned values in a map
			Src: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal(""),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
				}),
			}),
		},
		{
			// Remove extra values after apply
			Src: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal("b"),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
				}),
			}),
			Apply: true,
		},
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("a"),
			}),
			Dst: cty.EmptyObjectVal,
			Expect: cty.ObjectVal(map[string]cty.Value{
				"a": cty.NullVal(cty.String),
			}),
		},

		// a list in an object in a list, going from null to empty
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.UnknownVal(cty.String),
						"access_config": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String}))),
						"address":       cty.NullVal(cty.String),
						"name":          cty.StringVal("nic0"),
					})}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.StringVal("10.128.0.64"),
						"access_config": cty.ListValEmpty(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String})),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.StringVal("10.128.0.64"),
						"access_config": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String}))),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
			Apply: true,
		},

		// a list in an object in a list, going from empty to null
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.UnknownVal(cty.String),
						"access_config": cty.ListValEmpty(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String})),
						"address":       cty.NullVal(cty.String),
						"name":          cty.StringVal("nic0"),
					})}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.StringVal("10.128.0.64"),
						"access_config": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String}))),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.StringVal("10.128.0.64"),
						"access_config": cty.ListValEmpty(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String})),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
			Apply: true,
		},
		// the empty list should be transferred, but the new unknown should not be overridden
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.StringVal("10.128.0.64"),
						"access_config": cty.ListValEmpty(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String})),
						"address":       cty.NullVal(cty.String),
						"name":          cty.StringVal("nic0"),
					})}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.UnknownVal(cty.String),
						"access_config": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String}))),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"network_interface": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"network_ip":    cty.UnknownVal(cty.String),
						"access_config": cty.ListValEmpty(cty.Object(map[string]cty.Type{"public_ptr_domain_name": cty.String, "nat_ip": cty.String})),
						"address":       cty.StringVal("address"),
						"name":          cty.StringVal("nic0"),
					}),
				}),
			}),
		},
		{
			// fix unknowns added to a map
			Src: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal(""),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.UnknownVal(cty.String),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("a"),
					"b": cty.StringVal(""),
				}),
			}),
		},
		{
			// fix unknowns lost from a list
			Src: cty.ObjectVal(map[string]cty.Value{
				"top": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"list": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"values": cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
							}),
						}),
					}),
				}),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"top": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"list": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"values": cty.NullVal(cty.List(cty.String)),
							}),
						}),
					}),
				}),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"top": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"list": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"values": cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
							}),
						}),
					}),
				}),
			}),
		},
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				}))),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				})),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				})),
			}),
		},
		{
			Src: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				}))),
			}),
			Dst: cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				})),
			}),
			Expect: cty.ObjectVal(map[string]cty.Value{
				"set": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
				}))),
			}),
			Apply: true,
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := normalizeNullValues(tc.Dst, tc.Src, tc.Apply)
			if !got.RawEquals(tc.Expect) {
				t.Fatalf("\nexpected: %#v\ngot:      %#v\n", tc.Expect, got)
			}
		})
	}
}

func TestValidateNulls(t *testing.T) {
	for i, tc := range []struct {
		Cfg cty.Value
		Err bool
	}{
		{
			Cfg: cty.ObjectVal(map[string]cty.Value{
				"list": cty.ListVal([]cty.Value{
					cty.StringVal("string"),
					cty.NullVal(cty.String),
				}),
			}),
			Err: true,
		},
		{
			Cfg: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapVal(map[string]cty.Value{
					"string": cty.StringVal("string"),
					"null":   cty.NullVal(cty.String),
				}),
			}),
			Err: false,
		},
		{
			Cfg: cty.ObjectVal(map[string]cty.Value{
				"object": cty.ObjectVal(map[string]cty.Value{
					"list": cty.ListVal([]cty.Value{
						cty.StringVal("string"),
						cty.NullVal(cty.String),
					}),
				}),
			}),
			Err: true,
		},
		{
			Cfg: cty.ObjectVal(map[string]cty.Value{
				"object": cty.ObjectVal(map[string]cty.Value{
					"list": cty.ListVal([]cty.Value{
						cty.StringVal("string"),
						cty.NullVal(cty.String),
					}),
					"list2": cty.ListVal([]cty.Value{
						cty.StringVal("string"),
						cty.NullVal(cty.String),
					}),
				}),
			}),
			Err: true,
		},
		{
			Cfg: cty.ObjectVal(map[string]cty.Value{
				"object": cty.ObjectVal(map[string]cty.Value{
					"list": cty.SetVal([]cty.Value{
						cty.StringVal("string"),
						cty.NullVal(cty.String),
					}),
				}),
			}),
			Err: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := validateConfigNulls(context.Background(), tc.Cfg, nil)
			diags := convert.ProtoToDiags(d)
			switch {
			case tc.Err:
				if !diags.HasError() {
					t.Fatal("expected error")
				}
			default:
				for _, d := range diags {
					if d.Severity == diag.Error {
						t.Fatalf("unexpected error: %q", d)
					}
				}
			}
		})
	}
}

func TestStopContext_grpc(t *testing.T) {
	testCases := []struct {
		Description  string
		TestResource *Resource
	}{
		{
			Description: "CreateContext",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateContext: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateWithoutTimeout",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateWithoutTimeout: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				ResourcesMap: map[string]*Resource{
					"test": testCase.TestResource,
				},
			})

			schema := testCase.TestResource.CoreConfigSchema()
			priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			plannedVal, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.UnknownVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			plannedState, err := msgpack.Marshal(plannedVal, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.NullVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test",
				PriorState: &tfprotov5.DynamicValue{
					MsgPack: priorState,
				},
				PlannedState: &tfprotov5.DynamicValue{
					MsgPack: plannedState,
				},
				Config: &tfprotov5.DynamicValue{
					MsgPack: configBytes,
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			ctx = server.StopContext(ctx)
			doneCh := make(chan struct{})
			errCh := make(chan error)
			go func() {
				if _, err := server.ApplyResourceChange(ctx, testReq); err != nil {
					errCh <- err
				}
				close(doneCh)
			}()
			// GRPC request cancel
			cancel()
			select {
			case <-doneCh:
			case err := <-errCh:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("context cancel did not propagate")
			}
		})
	}
}

func TestStopContext_stop(t *testing.T) {
	testCases := []struct {
		Description  string
		TestResource *Resource
	}{
		{
			Description: "CreateContext",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateContext: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateWithoutTimeout",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateWithoutTimeout: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				ResourcesMap: map[string]*Resource{
					"test": testCase.TestResource,
				},
			})

			schema := testCase.TestResource.CoreConfigSchema()
			priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			plannedVal, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.UnknownVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			plannedState, err := msgpack.Marshal(plannedVal, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.NullVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test",
				PriorState: &tfprotov5.DynamicValue{
					MsgPack: priorState,
				},
				PlannedState: &tfprotov5.DynamicValue{
					MsgPack: plannedState,
				},
				Config: &tfprotov5.DynamicValue{
					MsgPack: configBytes,
				},
			}

			ctx := server.StopContext(context.Background())
			doneCh := make(chan struct{})
			errCh := make(chan error)
			go func() {
				if _, err := server.ApplyResourceChange(ctx, testReq); err != nil {
					errCh <- err
				}
				close(doneCh)
			}()

			if _, err := server.StopProvider(context.Background(), &tfprotov5.StopProviderRequest{}); err != nil {
				t.Fatalf("unexpected StopProvider error: %s", err)
			}

			select {
			case <-doneCh:
			case err := <-errCh:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Stop message did not cancel request context")
			}
		})
	}
}

func TestStopContext_stopReset(t *testing.T) {
	testCases := []struct {
		Description  string
		TestResource *Resource
	}{
		{
			Description: "CreateContext",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateContext: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
		{
			Description: "CreateWithoutTimeout",
			TestResource: &Resource{
				SchemaVersion: 4,
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				CreateWithoutTimeout: func(ctx context.Context, rd *ResourceData, _ interface{}) diag.Diagnostics {
					<-ctx.Done()
					rd.SetId("bar")
					return nil
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			server := NewGRPCProviderServer(&Provider{
				ResourcesMap: map[string]*Resource{
					"test": testCase.TestResource,
				},
			})

			schema := testCase.TestResource.CoreConfigSchema()
			priorState, err := msgpack.Marshal(cty.NullVal(schema.ImpliedType()), schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			plannedVal, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.UnknownVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			plannedState, err := msgpack.Marshal(plannedVal, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			config, err := schema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"id": cty.NullVal(cty.String),
			}))
			if err != nil {
				t.Fatal(err)
			}
			configBytes, err := msgpack.Marshal(config, schema.ImpliedType())
			if err != nil {
				t.Fatal(err)
			}

			testReq := &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test",
				PriorState: &tfprotov5.DynamicValue{
					MsgPack: priorState,
				},
				PlannedState: &tfprotov5.DynamicValue{
					MsgPack: plannedState,
				},
				Config: &tfprotov5.DynamicValue{
					MsgPack: configBytes,
				},
			}

			// test first stop
			ctx := server.StopContext(context.Background())
			if ctx.Err() != nil {
				t.Fatal("StopContext does not produce a non-closed context")
			}
			doneCh := make(chan struct{})
			errCh := make(chan error)
			go func(d chan struct{}) {
				if _, err := server.ApplyResourceChange(ctx, testReq); err != nil {
					errCh <- err
				}
				close(d)
			}(doneCh)

			if _, err := server.StopProvider(context.Background(), &tfprotov5.StopProviderRequest{}); err != nil {
				t.Fatalf("unexpected StopProvider error: %s", err)
			}

			select {
			case <-doneCh:
			case err := <-errCh:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Stop message did not cancel request context")
			}

			// test internal stop synchronization was reset
			ctx = server.StopContext(context.Background())
			if ctx.Err() != nil {
				t.Fatal("StopContext does not produce a non-closed context")
			}
			doneCh = make(chan struct{})
			errCh = make(chan error)
			go func(d chan struct{}) {
				if _, err := server.ApplyResourceChange(ctx, testReq); err != nil {
					errCh <- err
				}
				close(d)
			}(doneCh)

			if _, err := server.StopProvider(context.Background(), &tfprotov5.StopProviderRequest{}); err != nil {
				t.Fatalf("unexpected StopProvider error: %s", err)
			}

			select {
			case <-doneCh:
			case err := <-errCh:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Stop message did not cancel request context")
			}
		})
	}
}

func Test_pathToAttributePath_noSteps(t *testing.T) {
	res := pathToAttributePath(cty.Path{})
	if res != nil {
		t.Errorf("Expected nil attribute path, got %+v", res)
	}
}

func mustMsgpackMarshal(ty cty.Type, val cty.Value) []byte {
	result, err := msgpack.Marshal(val, ty)

	if err != nil {
		panic(fmt.Sprintf("cannot marshal msgpack: %s\n\ntype: %v\n\nvalue: %v", err, ty, val))
	}

	return result
}

func mustMsgpackUnmarshal(ty cty.Type, b []byte) cty.Value {
	result, err := msgpack.Unmarshal(b, ty)

	if err != nil {
		panic(fmt.Sprintf("cannot unmarshal msgpack: %s\n\ntype: %v\n\nvalue: %v", err, ty, b))
	}

	return result
}
