// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configschema

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestBlockEmptyValue(t *testing.T) {
	tests := []struct {
		Schema *Block
		Want   cty.Value
	}{
		{
			&Block{},
			cty.EmptyObjectVal,
		},
		{
			&Block{
				Attributes: map[string]*Attribute{
					"str": {Type: cty.String, Required: true},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"str": cty.NullVal(cty.String),
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"single": {
						Nesting: NestingSingle,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.String, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"single": cty.NullVal(cty.Object(map[string]cty.Type{
					"str": cty.String,
				})),
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"group": {
						Nesting: NestingGroup,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.String, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"group": cty.ObjectVal(map[string]cty.Value{
					"str": cty.NullVal(cty.String),
				}),
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"list": {
						Nesting: NestingList,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.String, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.ListValEmpty(cty.Object(map[string]cty.Type{
					"str": cty.String,
				})),
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"list_dynamic": {
						Nesting: NestingList,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.DynamicPseudoType, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list_dynamic": cty.EmptyTupleVal,
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"map": {
						Nesting: NestingMap,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.String, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.Object(map[string]cty.Type{
					"str": cty.String,
				})),
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"map_dynamic": {
						Nesting: NestingMap,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.DynamicPseudoType, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_dynamic": cty.EmptyObjectVal,
			}),
		},
		{
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"set": {
						Nesting: NestingSet,
						Block: Block{
							Attributes: map[string]*Attribute{
								"str": {Type: cty.String, Required: true},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"set": cty.SetValEmpty(cty.Object(map[string]cty.Type{
					"str": cty.String,
				})),
			}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Schema), func(t *testing.T) {
			got := test.Schema.EmptyValue()
			if !test.Want.RawEquals(got) {
				t.Errorf("wrong result\nschema: %#v\ngot: %#v\nwant: %#v", test.Schema, got, test.Want)
			}

			// The empty value must always conform to the implied type of
			// the schema.
			wantTy := test.Schema.ImpliedType()
			gotTy := got.Type()
			if errs := gotTy.TestConformance(wantTy); len(errs) > 0 {
				t.Errorf("empty value has incorrect type\ngot: %#v\nwant: %#v\nerrors: %#v", gotTy, wantTy, errs)
			}
		})
	}
}
