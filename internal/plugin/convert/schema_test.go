// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
)

var (
	equateEmpty   = cmpopts.EquateEmpty()
	typeComparer  = cmp.Comparer(cty.Type.Equals)
	valueComparer = cmp.Comparer(cty.Value.RawEquals)
)

// Test that we can convert configschema to protobuf types and back again.
func TestConvertSchemaBlocks(t *testing.T) {
	tests := map[string]struct {
		Block *tfprotov5.SchemaBlock
		Want  *configschema.Block
	}{
		"attributes": {
			&tfprotov5.SchemaBlock{
				Attributes: []*tfprotov5.SchemaAttribute{
					{
						Name: "computed",
						Type: tftypes.List{
							ElementType: tftypes.Bool,
						},
						Computed: true,
					},
					{
						Name:     "optional",
						Type:     tftypes.String,
						Optional: true,
					},
					{
						Name: "optional_computed",
						Type: tftypes.Map{
							ElementType: tftypes.Bool,
						},
						Optional: true,
						Computed: true,
					},
					{
						Name:     "required",
						Type:     tftypes.Number,
						Required: true,
					},
				},
			},
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"computed": {
						Type:     cty.List(cty.Bool),
						Computed: true,
					},
					"optional": {
						Type:     cty.String,
						Optional: true,
					},
					"optional_computed": {
						Type:     cty.Map(cty.Bool),
						Optional: true,
						Computed: true,
					},
					"required": {
						Type:     cty.Number,
						Required: true,
					},
				},
			},
		},
		"blocks": {
			&tfprotov5.SchemaBlock{
				BlockTypes: []*tfprotov5.SchemaNestedBlock{
					{
						TypeName: "list",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "map",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeMap,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "set",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "single",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "foo",
									Type:     tftypes.DynamicPseudoType,
									Required: true,
								},
							},
						},
					},
				},
			},
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list": {
						Nesting: configschema.NestingList,
					},
					"map": {
						Nesting: configschema.NestingMap,
					},
					"set": {
						Nesting: configschema.NestingSet,
					},
					"single": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"foo": {
									Type:     cty.DynamicPseudoType,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"deep block nesting": {
			&tfprotov5.SchemaBlock{
				BlockTypes: []*tfprotov5.SchemaNestedBlock{
					{
						TypeName: "single",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
						Block: &tfprotov5.SchemaBlock{
							BlockTypes: []*tfprotov5.SchemaNestedBlock{
								{
									TypeName: "list",
									Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
									Block: &tfprotov5.SchemaBlock{
										BlockTypes: []*tfprotov5.SchemaNestedBlock{
											{
												TypeName: "set",
												Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
												Block:    &tfprotov5.SchemaBlock{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"single": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							BlockTypes: map[string]*configschema.NestedBlock{
								"list": {
									Nesting: configschema.NestingList,
									Block: configschema.Block{
										BlockTypes: map[string]*configschema.NestedBlock{
											"set": {
												Nesting: configschema.NestingSet,
											},
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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			converted := ProtoToConfigSchema(context.Background(), tc.Block)
			if !cmp.Equal(converted, tc.Want, typeComparer, valueComparer, equateEmpty) {
				t.Fatal(cmp.Diff(converted, tc.Want, typeComparer, valueComparer, equateEmpty))
			}
		})
	}
}

// Test that we can convert configschema to protobuf types and back again.
func TestConvertProtoSchemaBlocks(t *testing.T) {
	tests := map[string]struct {
		Want  *tfprotov5.SchemaBlock
		Block *configschema.Block
	}{
		"attributes": {
			&tfprotov5.SchemaBlock{
				Attributes: []*tfprotov5.SchemaAttribute{
					{
						Name: "computed",
						Type: tftypes.List{
							ElementType: tftypes.Bool,
						},
						Computed: true,
					},
					{
						Name:     "optional",
						Type:     tftypes.String,
						Optional: true,
					},
					{
						Name: "optional_computed",
						Type: tftypes.Map{
							ElementType: tftypes.Bool,
						},
						Optional: true,
						Computed: true,
					},
					{
						Name:     "required",
						Type:     tftypes.Number,
						Required: true,
					},
				},
			},
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"computed": {
						Type:     cty.List(cty.Bool),
						Computed: true,
					},
					"optional": {
						Type:     cty.String,
						Optional: true,
					},
					"optional_computed": {
						Type:     cty.Map(cty.Bool),
						Optional: true,
						Computed: true,
					},
					"required": {
						Type:     cty.Number,
						Required: true,
					},
				},
			},
		},
		"blocks": {
			&tfprotov5.SchemaBlock{
				BlockTypes: []*tfprotov5.SchemaNestedBlock{
					{
						TypeName: "list",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "map",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeMap,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "set",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
						Block:    &tfprotov5.SchemaBlock{},
					},
					{
						TypeName: "single",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "foo",
									Type:     tftypes.DynamicPseudoType,
									Required: true,
								},
							},
						},
					},
				},
			},
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list": {
						Nesting: configschema.NestingList,
					},
					"map": {
						Nesting: configschema.NestingMap,
					},
					"set": {
						Nesting: configschema.NestingSet,
					},
					"single": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"foo": {
									Type:     cty.DynamicPseudoType,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"deep block nesting": {
			&tfprotov5.SchemaBlock{
				BlockTypes: []*tfprotov5.SchemaNestedBlock{
					{
						TypeName: "single",
						Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
						Block: &tfprotov5.SchemaBlock{
							BlockTypes: []*tfprotov5.SchemaNestedBlock{
								{
									TypeName: "list",
									Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
									Block: &tfprotov5.SchemaBlock{
										BlockTypes: []*tfprotov5.SchemaNestedBlock{
											{
												TypeName: "set",
												Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
												Block:    &tfprotov5.SchemaBlock{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"single": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							BlockTypes: map[string]*configschema.NestedBlock{
								"list": {
									Nesting: configschema.NestingList,
									Block: configschema.Block{
										BlockTypes: map[string]*configschema.NestedBlock{
											"set": {
												Nesting: configschema.NestingSet,
											},
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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			converted := ConfigSchemaToProto(context.Background(), tc.Block)
			if !cmp.Equal(converted, tc.Want, typeComparer, equateEmpty) {
				t.Fatal(cmp.Diff(converted, tc.Want, typeComparer, equateEmpty))
			}
		})
	}
}
