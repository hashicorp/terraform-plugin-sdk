// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProtoSchema(t *testing.T) {
	tests := map[string]struct {
		input    *schema.Resource
		expected *tfprotov5.Schema
	}{
		"empty": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
		"primitives": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"int": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "foo bar baz",
					},
					"float": {
						Type:     schema.TypeFloat,
						Optional: true,
					},
					"bool": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"string": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						{
							Name:     "bool",
							Type:     tftypes.Bool,
							Computed: true,
						},
						{
							Name:     "float",
							Type:     tftypes.Number,
							Optional: true,
						},
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:        "int",
							Type:        tftypes.Number,
							Description: "foo bar baz",
							Required:    true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
		"simple collections": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"list": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeInt,
						},
					},
					"set": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"map": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeBool,
						},
					},
					"map_default_type": {
						Type:     schema.TypeMap,
						Optional: true,
						// Maps historically don't have elements because we
						// assumed they would be strings, so this needs to work
						// for pre-existing schemas.
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.Number},
							Required: true,
						},
						{
							Name:     "map",
							Type:     tftypes.Map{ElementType: tftypes.Bool},
							Optional: true,
						},
						{
							Name:     "map_default_type",
							Type:     tftypes.Map{ElementType: tftypes.String},
							Optional: true,
						},
						{
							Name:     "set",
							Type:     tftypes.Set{ElementType: tftypes.String},
							Optional: true,
						},
					},
				},
			},
		},
		"incorrectly-specified collections": {
			// Historically we tolerated setting a type directly as the Elem
			// attribute, rather than a Schema object. This is common enough
			// in existing provider code that we must support it as an alias
			// for a schema object with the given type.
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"list": {
						Type:     schema.TypeList,
						Required: true,
						Elem:     schema.TypeInt,
					},
					"set": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     schema.TypeString,
					},
					"map": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem:     schema.TypeBool,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.Number},
							Required: true,
						},
						{
							Name:     "map",
							Type:     tftypes.Map{ElementType: tftypes.Bool},
							Optional: true,
						},
						{
							Name:     "set",
							Type:     tftypes.Set{ElementType: tftypes.String},
							Optional: true,
						},
					},
				},
			},
		},
		"sub-resource collections": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"list": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
						MinItems: 1,
						MaxItems: 2,
					},
					"set": {
						Type:     schema.TypeSet,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
					},
					"map": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						// This one becomes a string attribute because helper/schema
						// doesn't actually support maps of resource. The given
						// "Elem" is just ignored entirely here, which is important
						// because that is also true of the helper/schema logic and
						// existing providers rely on this being ignored for
						// correct operation.
						{
							Name:     "map",
							Type:     tftypes.Map{ElementType: tftypes.String},
							Optional: true,
						},
					},
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "list",
							Block:    &tfprotov5.SchemaBlock{},
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
							MinItems: 1,
							MaxItems: 2,
						},
						{
							TypeName: "set",
							Block:    &tfprotov5.SchemaBlock{},
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
							MinItems: 1, // because schema is Required
						},
					},
				},
			},
		},
		"sub-resource collections minitems+optional": {
			// This particular case is an odd one where the provider gives
			// conflicting information about whether a sub-resource is required,
			// by marking it as optional but also requiring one item.
			// Historically the optional-ness "won" here, and so we must
			// honor that for compatibility with providers that relied on this
			// undocumented interaction.
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"list": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
						MinItems: 1,
						MaxItems: 1,
					},
					"set": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
						MinItems: 1,
						MaxItems: 1,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "list",
							Block:    &tfprotov5.SchemaBlock{},
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
							MinItems: 0,
							MaxItems: 1,
						},
						{
							TypeName: "set",
							Block:    &tfprotov5.SchemaBlock{},
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
							MinItems: 0,
							MaxItems: 1,
						},
					},
				},
			},
		},
		"sub-resource collections minitems+computed": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"list": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
						MinItems: 1,
						MaxItems: 1,
					},
					"set": {
						Type:     schema.TypeSet,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
						MinItems: 1,
						MaxItems: 1,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}},
							Computed: true,
						},
						{
							Name:     "set",
							Type:     tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}},
							Computed: true,
						},
					},
				},
			},
		},
		"nested attributes and blocks": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"foo": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"bar": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
								"baz": {
									Type:     schema.TypeSet,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "foo",
							Block: &tfprotov5.SchemaBlock{
								Attributes: []*tfprotov5.SchemaAttribute{
									{
										Name:     "bar",
										Type:     tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}},
										Required: true,
									},
								},
								BlockTypes: []*tfprotov5.SchemaNestedBlock{
									{
										TypeName: "baz",
										Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
										Block:    &tfprotov5.SchemaBlock{},
									},
								},
							},
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
							MinItems: 1, // because schema is Required
						},
					},
				},
			},
		},
		"sensitive": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:      "string",
							Type:      tftypes.String,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		"conditionally required on": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:     schema.TypeString,
						Required: true,
						DefaultFunc: func() (interface{}, error) {
							return nil, nil
						},
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
				},
			},
		},
		"conditionally required off": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:     schema.TypeString,
						Required: true,
						DefaultFunc: func() (interface{}, error) {
							// If we return a non-nil default then this overrides
							// the "Required: true" for the purpose of building
							// the core schema, so that core will ignore it not
							// being set and let the provider handle it.
							return "boop", nil
						},
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Optional: true,
						},
					},
				},
			},
		},
		"conditionally required error": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:     schema.TypeString,
						Required: true,
						DefaultFunc: func() (interface{}, error) {
							return nil, fmt.Errorf("placeholder error")
						},
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Optional: true, // Just so we can progress to provider-driven validation and return the error there
						},
					},
				},
			},
		},
		"write-only": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:      schema.TypeString,
						Optional:  true,
						WriteOnly: true,
					},
				},
			},
			expected: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						// ID is automatically added by SDKv2
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
						{
							Name:      "string",
							Type:      tftypes.String,
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.ProtoSchema(context.Background())()
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestProtoIdentitySchema(t *testing.T) {
	tests := map[string]struct {
		input    *schema.Resource
		expected *tfprotov5.ResourceIdentitySchema
	}{
		"empty": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{},
			},
			expected: nil,
		},
		"no-identity": {
			input: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"string": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
			expected: nil,
		},
		"primitives": {
			input: &schema.Resource{
				Identity: &schema.ResourceIdentity{
					SchemaFunc: func() map[string]*schema.Schema {
						return map[string]*schema.Schema{
							"float": {
								Type:              schema.TypeFloat,
								OptionalForImport: true,
							},
							"bool": {
								Type:              schema.TypeBool,
								OptionalForImport: true,
							},
							"string": {
								Type:              schema.TypeString,
								OptionalForImport: true,
							},
							"int": {
								Type:              schema.TypeInt,
								RequiredForImport: true,
								Description:       "foo bar baz",
							},
						}
					},
				},
				Schema: map[string]*schema.Schema{},
			},
			expected: &tfprotov5.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
					{
						Name:              "bool",
						Type:              tftypes.Bool,
						OptionalForImport: true,
					},
					{
						Name:              "float",
						Type:              tftypes.Number,
						OptionalForImport: true,
					},
					{
						Name:              "int",
						Type:              tftypes.Number,
						Description:       "foo bar baz",
						RequiredForImport: true,
					},
					{
						Name:              "string",
						Type:              tftypes.String,
						OptionalForImport: true,
					},
				},
			},
		},
		"list": {
			input: &schema.Resource{
				Identity: &schema.ResourceIdentity{
					SchemaFunc: func() map[string]*schema.Schema {
						return map[string]*schema.Schema{
							"list": {
								Type:              schema.TypeList,
								RequiredForImport: true,
								Elem: &schema.Schema{
									Type: schema.TypeInt,
								},
							},
						}
					},
				},
				Schema: map[string]*schema.Schema{},
			},
			expected: &tfprotov5.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
					{
						Name:              "list",
						Type:              tftypes.List{ElementType: tftypes.Number},
						RequiredForImport: true,
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.ProtoIdentitySchema(context.Background())
			// Nil identity function is valid, we can return
			if got == nil && tc.expected == nil {
				return
			}

			if diff := cmp.Diff(got(), tc.expected); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
