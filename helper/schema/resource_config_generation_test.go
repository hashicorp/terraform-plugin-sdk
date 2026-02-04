package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
)

func TestGenerateConflictsWith(t *testing.T) {
	cases := map[string]struct {
		Schema              map[string]*Schema
		expectConflictsWith map[string][]string
	}{
		"Empty": {
			Schema:              map[string]*Schema{},
			expectConflictsWith: map[string][]string{},
		},
		"Top-level attribute no ConflictsWith returns empty map": {
			Schema: map[string]*Schema{
				"string_attr": {
					Type:     TypeString,
					Optional: true,
				},
			},
			expectConflictsWith: map[string][]string{},
		},
		"Multiple top-level attributes with ConflictsWith": {
			Schema: map[string]*Schema{
				"list_attr": {
					Type:          TypeList,
					Optional:      true,
					Elem:          &Schema{Type: TypeString},
					ConflictsWith: []string{"bool_attr"},
				},
				"bool_attr": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"list_attr"},
				},
			},
			expectConflictsWith: map[string][]string{
				"870212088": {"bool_attr", "list_attr"},
			},
		},
		"Multiple top-level attributes multiple ConflictsWith sets": {
			Schema: map[string]*Schema{
				"list_attr": {
					Type:          TypeList,
					Optional:      true,
					Elem:          &Schema{Type: TypeString},
					ConflictsWith: []string{"bool_attr"},
				},
				"bool_attr": {
					Type:          TypeBool,
					Optional:      true,
					ConflictsWith: []string{"list_attr"},
				},
				"set_attr": {
					Type:          TypeSet,
					Optional:      true,
					Elem:          &Schema{Type: TypeString},
					ConflictsWith: []string{"string_attr"},
				},
				"string_attr": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"set_attr"},
				},
			},
			expectConflictsWith: map[string][]string{
				"1259552064": {"set_attr", "string_attr"},
				"870212088":  {"bool_attr", "list_attr"},
			},
		},
		"List configuration block": {
			Schema: map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"bool_attr"},
							},
						},
					},
				},
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
			},
			expectConflictsWith: map[string][]string{
				"3804550687": {"bool_attr", "config_block_attr.0.nested_attr"},
			},
		},
		"List configuration block double nested": {
			Schema: map[string]*Schema{
				"config_block_attr": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
								Elem: &Resource{
									Schema: map[string]*Schema{
										"nested_nested_attr": {
											Type:          TypeString,
											Optional:      true,
											ConflictsWith: []string{"bool_attr"},
										},
									},
								},
							},
						},
					},
				},
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
			},
			expectConflictsWith: map[string][]string{
				"3810369191": {"bool_attr", "config_block_attr.0.nested_attr.0.nested_nested_attr"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualConflictsWith := generateConflictsWith(tc.Schema, "")
			if diff := cmp.Diff(actualConflictsWith, tc.expectConflictsWith); diff != "" {

				t.Error(diff)
			}
		})
	}
}

// Todo: add tests for all configGenSchema fields
func TestConfigGenerationSchemaMap(t *testing.T) {
	cases := map[string]struct {
		Schema          map[string]*Schema
		expectSchemaMap map[string]*configGenSchema
	}{
		"Empty": {
			Schema:          map[string]*Schema{},
			expectSchemaMap: map[string]*configGenSchema{},
		},
		"Top-level attribute": {
			Schema: map[string]*Schema{
				"string_attr": {
					Type:     TypeString,
					Optional: true,
				},
			},
			expectSchemaMap: map[string]*configGenSchema{
				"string_attr": {
					Optional: true,
				},
			},
		},
		"List configuration block": {
			Schema: map[string]*Schema{
				"config_block_attr": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"bool_attr"},
							},
						},
					},
				},
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
			},
			expectSchemaMap: map[string]*configGenSchema{
				"config_block_attr": {
					Optional: true,
				},
				"config_block_attr.0.nested_attr": {
					Optional:      true,
					ConflictsWith: []string{"bool_attr"},
				},
				"bool_attr": {
					Optional: true,
				},
			},
		},
		"Double Nested Block": {
			Schema: map[string]*Schema{
				"config_block": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_block": {
								Type:     TypeString,
								Optional: true,
								Elem: &Resource{
									Schema: map[string]*Schema{
										"nested_block_attr": {
											Type:          TypeString,
											Optional:      true,
											ConflictsWith: []string{"bool_attr"},
										},
									},
								},
							},
						},
					},
				},
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
			},
			expectSchemaMap: map[string]*configGenSchema{
				"config_block": {
					Optional: true,
				},
				"config_block.0.nested_block": {
					Optional: true,
				},
				"config_block.0.nested_block.0.nested_block_attr": {
					Optional:      true,
					ConflictsWith: []string{"bool_attr"},
				},
				"bool_attr": {
					Optional: true,
				},
			},
		},
		//"Multiple top-level attributes with ConflictsWith": {
		//	Schema: map[string]*Schema{
		//		"list_attr": {
		//			Type:          TypeList,
		//			Optional:      true,
		//			Elem:          &Schema{Type: TypeString},
		//			ConflictsWith: []string{"bool_attr"},
		//		},
		//		"bool_attr": {
		//			Type:          TypeBool,
		//			Optional:      true,
		//			ConflictsWith: []string{"list_attr"},
		//		},
		//	},
		//	expectConflictsWith: map[string][]string{
		//		"870212088": {"bool_attr", "list_attr"},
		//	},
		//},
		//"Multiple top-level attributes multiple ConflictsWith sets": {
		//	Schema: map[string]*Schema{
		//		"list_attr": {
		//			Type:          TypeList,
		//			Optional:      true,
		//			Elem:          &Schema{Type: TypeString},
		//			ConflictsWith: []string{"bool_attr"},
		//		},
		//		"bool_attr": {
		//			Type:          TypeBool,
		//			Optional:      true,
		//			ConflictsWith: []string{"list_attr"},
		//		},
		//		"set_attr": {
		//			Type:          TypeSet,
		//			Optional:      true,
		//			Elem:          &Schema{Type: TypeString},
		//			ConflictsWith: []string{"string_attr"},
		//		},
		//		"string_attr": {
		//			Type:          TypeString,
		//			Optional:      true,
		//			ConflictsWith: []string{"set_attr"},
		//		},
		//	},
		//	expectConflictsWith: map[string][]string{
		//		"1259552064": {"set_attr", "string_attr"},
		//		"870212088":  {"bool_attr", "list_attr"},
		//	},
		//},
		//"List configuration block": {
		//	Schema: map[string]*Schema{
		//		"config_block_attr": {
		//			Type:     TypeList,
		//			Optional: true,
		//			MaxItems: 1,
		//			Elem: &Resource{
		//				Schema: map[string]*Schema{
		//					"nested_attr": {
		//						Type:          TypeString,
		//						Optional:      true,
		//						ConflictsWith: []string{"bool_attr"},
		//					},
		//				},
		//			},
		//		},
		//		"bool_attr": {
		//			Type:     TypeBool,
		//			Optional: true,
		//		},
		//	},
		//	expectConflictsWith: map[string][]string{
		//		"3804550687": {"bool_attr", "config_block_attr.0.nested_attr"},
		//	},
		//},

	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualConflictsWith := make(map[string]*configGenSchema)
			configGenerationSchemaMap(tc.Schema, actualConflictsWith, "")
			if diff := cmp.Diff(actualConflictsWith, tc.expectSchemaMap); diff != "" {

				t.Error(diff)
			}
		})
	}
}

func TestProcessConflictsWith(t *testing.T) {
	cases := map[string]struct {
		Schema           map[string]*Schema
		ctyVal           cty.Value
		expectedCtyVal   cty.Value
		expectedPaths    []string
		expectedCtyPaths cty.PathSet
	}{
		//"Empty": {
		//	Schema:          map[string]*Schema{},
		//	expectSchemaMap: map[string]*configGenSchema{},
		//},
		"Top-level attribute - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("val"),
				"attr_b": cty.StringVal("val"),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("val"),
				"attr_b": cty.StringVal("val"),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"Top-level attribute - other": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("val"),
				"attr_b": cty.StringVal("val"),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("val"),
				"attr_b": cty.NullVal(cty.String),
			}),
			expectedPaths: []string{},
		},
		"list attribute - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeList,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.NullVal(cty.List(cty.String)),
			}),
			expectedPaths: []string{},
		},
		"list attribute - marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"list attribute - other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeList,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
				"attr_b": cty.StringVal("none"),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
				"attr_b": cty.StringVal("none"),
			}),
			// expect 3 paths for each visit to attr_a,
			// one for the parent list itself and 2 for the elements.
			expectedPaths: []string{
				"attr_b",
				"attr_b",
				"attr_b",
			},
		},
		"set attribute - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeSet,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.NullVal(cty.Set(cty.String)),
			}),
			expectedPaths: []string{},
		},
		"set attribute - marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"set attribute - other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeSet,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
				"attr_b": cty.StringVal("none"),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
				"attr_b": cty.StringVal("none"),
			}),
			// expect 3 paths for each visit to attr_a,
			// one for the parent list itself and 2 for the elements.
			expectedPaths: []string{
				"attr_b",
				"attr_b",
				"attr_b",
			},
		},
		"map attribute - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeMap,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.NullVal(cty.Map(cty.String)),
			}),
			expectedPaths: []string{},
		},
		"map attribute - marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"map attribute - other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeMap,
					Elem:          &Schema{Type: TypeString},
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
				"attr_b": cty.StringVal("value"),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
				"attr_b": cty.StringVal("value"),
			}),
			// expect 3 paths for each visit to attr_a,
			// one for the parent list itself and 2 for the elements.
			expectedPaths: []string{
				"attr_b",
				"attr_b",
				"attr_b",
			},
		},
		"list block - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeList,
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.NullVal(cty.List(
					cty.Object(map[string]cty.Type{
						"nested_attr": cty.String,
					},
					))),
			}),
			expectedPaths: []string{},
		},
		"list block - marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
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
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"list block - other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_a": {
					Type:          TypeList,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_b": cty.StringVal("value"),
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_b": cty.StringVal("value"),
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedPaths: []string{
				"attr_b",
				"attr_b",
				"attr_b",
			},
		},
		"set block - nullified": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:          TypeSet,
					Optional:      true,
					ConflictsWith: []string{"attr_a"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.NullVal(cty.Set(
					cty.Object(map[string]cty.Type{
						"nested_attr": cty.String,
					},
					))),
			}),
			expectedPaths: []string{},
		},
		"set block - marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeString,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
				},
				"attr_b": {
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
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("value"),
				"attr_b": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedPaths: []string{
				"attr_b",
			},
		},
		"set block - other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_b": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_a": {
					Type:          TypeSet,
					Optional:      true,
					ConflictsWith: []string{"attr_b"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_b": cty.StringVal("value"),
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedCtyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_b": cty.StringVal("value"),
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedPaths: []string{
				"attr_b",
				"attr_b",
				"attr_b",
			},
		},
		"set block": {
			Schema: map[string]*Schema{
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
				"set_block": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"bool_attr"},
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"bool_attr": cty.BoolVal(true),
			}),
			expectedPaths: []string{
				"bool_attr",
				"set_block.0.nested_attr",
				"set_block",
				"set_block.0.nested_attr",
				"set_block",
				"set_block",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualCty, actualPaths, err := processConflictsWith(tc.ctyVal, tc.Schema)
			if err != nil {
				t.Fatal(err)
			}
			//if !actualCty.RawEquals(tc.expectedCtyVal) {
			//	t.Errorf("Cty val does not match")
			//}
			if diff := cmp.Diff(actualCty, tc.expectedCtyVal, valueComparer); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(actualPaths, tc.expectedPaths); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCtyWalkWithFlatmapPaths(t *testing.T) {
	cases := map[string]struct {
		Schema        map[string]*Schema
		ctyVal        cty.Value
		expectedPaths []string
	}{
		//"Empty": {
		//	Schema:          map[string]*Schema{},
		//	expectSchemaMap: map[string]*configGenSchema{},
		//},
		"Top-level attribute": {
			Schema: map[string]*Schema{
				"string_attr": {
					Type:     TypeString,
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"string_attr": cty.StringVal("val"),
			}),
			expectedPaths: []string{
				"string_attr",
			},
		},
		"list attribute": {
			Schema: map[string]*Schema{
				"list_attr": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"list_attr": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"list_attr",
				"list_attr",
				"list_attr",
			},
		},
		"set attribute": {
			Schema: map[string]*Schema{
				"set_attr": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"set_attr": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"set_attr",
				"set_attr",
				"set_attr",
			},
		},
		"map attribute": {
			Schema: map[string]*Schema{
				"map_attr": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"map_attr": cty.MapVal(map[string]cty.Value{
					"vala": cty.StringVal("true"),
					"valb": cty.StringVal("false")}),
			}),
			expectedPaths: []string{
				"map_attr",
				"map_attr",
				"map_attr",
			},
		},
		"list block": {
			Schema: map[string]*Schema{
				"list_block": {
					Type:     TypeList,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"bool_attr"},
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
			}),
			expectedPaths: []string{
				"list_block",
				"list_block",
				"list_block.0.nested_attr",
				"list_block",
				"list_block.0.nested_attr",
			},
		},
		"set block": {
			Schema: map[string]*Schema{
				"bool_attr": {
					Type:     TypeBool,
					Optional: true,
				},
				"set_block": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:          TypeString,
								Optional:      true,
								ConflictsWith: []string{"bool_attr"},
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"bool_attr": cty.BoolVal(true),
			}),
			expectedPaths: []string{
				"bool_attr",
				"set_block.0.nested_attr",
				"set_block",
				"set_block.0.nested_attr",
				"set_block",
				"set_block",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualPaths := make([]string, 0)
			configGenSchemaMap := make(map[string]*configGenSchema)
			configGenerationSchemaMap(tc.Schema, configGenSchemaMap, "")
			newResourceConfigShimmedComputedKeys(tc.ctyVal, "")
			_, err := cty.Transform(tc.ctyVal, func(path cty.Path, value cty.Value) (cty.Value, error) {
				if len(path) == 0 {
					return value, nil
				}

				flatmapPath := ctyPathToFlatmapPath(path)
				actualPaths = append(actualPaths, flatmapPath)
				_, ok := configGenSchemaMap[flatmapPath]
				if !ok {
					t.Errorf("Cannot retrieve config schema at key %s", flatmapPath)
				}
				return value, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actualPaths, tc.expectedPaths); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestFlatmapPathToCtyPath(t *testing.T) {
	tests := map[string]struct {
		p    string
		want cty.Path
	}{
		"empty paths returns true": {
			p:    "",
			want: cty.Path{},
		},
		"exact same path returns true": {
			p:    "attribute",
			want: cty.GetAttrPath("attribute"),
		},
		"path with unknown number index returns true": {
			p:    "attribute.1.nestedAttribute",
			want: cty.GetAttrPath("attribute").Index(cty.NumberIntVal(1)).GetAttr("nestedAttribute"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := flatmapPathToCtyPath(tc.p); !got.Equals(tc.want) {
				t.Errorf("Converted path: '%s' does not match expected %v", tc.p, tc.want)
			}
		})
	}
}
