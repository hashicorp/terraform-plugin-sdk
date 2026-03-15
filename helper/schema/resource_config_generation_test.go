package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
)

func TestProcessConflictsWith(t *testing.T) {
	cases := map[string]struct {
		Schema                         map[string]*Schema
		ctyVal                         cty.Value
		expectedMarkedForNullification cty.PathSet
	}{
		"primitive attribute with ConflictsWith: self marked for nullification": {
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
			expectedMarkedForNullification: cty.NewPathSet(cty.GetAttrPath("attr_b")),
		},
		"primitive attribute with ConflictsWith: other attribute marked for nullification": {
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
			expectedMarkedForNullification: cty.NewPathSet(cty.GetAttrPath("attr_b")),
		},
		"list attribute with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
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
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list attribute with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeList,
					Elem:          &Schema{Type: TypeString},
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
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set attribute with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
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
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set attribute with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeSet,
					Elem:          &Schema{Type: TypeString},
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
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"map attribute with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
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
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("false"),
					"key_b": cty.StringVal("true")}),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"map attribute with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:          TypeMap,
					Elem:          &Schema{Type: TypeString},
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
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("false"),
					"key_b": cty.StringVal("true")}),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list block with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list block with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set block with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set block with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
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
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"single nested block attribute with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ConflictsWith: []string{"attr_a.0.nested_attr"},
								Type:          TypeString,
								Optional:      true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
		"single nested block attribute with ConflictsWith: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ConflictsWith: []string{"attr_b.0.nested_attr"},
								Type:          TypeString,
								Optional:      true,
							},
						},
					},
				},
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
		"set nested block attribute with ConflictsWith: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ConflictsWith: []string{"attr_a"},
								Type:          TypeString,
								Optional:      true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("hello"),
				})).GetAttr("nested_attr"),
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("goodbye"),
				})).GetAttr("nested_attr"),
			),
		},
		"full flatmap path does not exist": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ConflictsWith: []string{"attr_a.0.nested_attr"},
								Type:          TypeString,
								Optional:      true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			actualMarkedForNullification := cty.NewPathSet()
			_, err := cty.Transform(tc.ctyVal, func(path cty.Path, val cty.Value) (cty.Value, error) {
				if val.IsNull() {
					return val, nil
				}

				if len(path) == 0 {
					return val, nil
				}

				// find the attribute or block schema representing the value
				attr := schemaMap(tc.Schema).AttributeByPath(path)
				block := schemaMap(tc.Schema).BlockByPath(path)

				if attr == nil && block == nil {
					return val, nil
				}
				var schema *Schema
				if attr != nil {
					schema = attr
				} else {
					schema = block
				}

				markedfordeletion, err := processConflictsWith(schema.ConflictsWith, tc.ctyVal, path)
				if err != nil {
					return cty.Value{}, err
				}
				for _, p := range markedfordeletion.List() {
					actualMarkedForNullification.Add(p)
				}

				return val, nil

			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actualMarkedForNullification, tc.expectedMarkedForNullification, pathSetComparer); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestProcessExactlyOneOf(t *testing.T) {
	cases := map[string]struct {
		Schema                         map[string]*Schema
		ctyVal                         cty.Value
		expectedMarkedForNullification cty.PathSet
	}{
		"primitive attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:         TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.StringVal("val"),
				"attr_b": cty.StringVal("val"),
			}),
			expectedMarkedForNullification: cty.NewPathSet(cty.GetAttrPath("attr_b")),
		},
		"primitive attribute with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
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
			expectedMarkedForNullification: cty.NewPathSet(cty.GetAttrPath("attr_b")),
		},
		"list attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeList,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list attribute with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeList,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeSet,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set attribute with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeSet,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal([]cty.Value{cty.StringVal("false"), cty.StringVal("true")}),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"map attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeMap,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("false"),
					"key_b": cty.StringVal("true")}),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"map attribute with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeMap,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
				},
				"attr_b": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("false"),
					"key_b": cty.StringVal("true")}),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list block with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list block with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeList,
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set block with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:         TypeSet,
					Optional:     true,
					ExactlyOneOf: []string{"attr_a", "attr_b"},
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
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set block with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:         TypeSet,
					Optional:     true,
					ExactlyOneOf: []string{"attr_b"},
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								Type:     TypeString,
								Optional: true,
							},
						},
					},
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
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"single nested block attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ExactlyOneOf: []string{"attr_a.0.nested_attr"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
		"single nested block attribute with ExactlyOneOf: other attribute marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ExactlyOneOf: []string{"attr_b.0.nested_attr"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
					}),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
		"set nested block attribute with ExactlyOneOf: self marked for nullification": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ExactlyOneOf: []string{"attr_a"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.SetVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("goodbye"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("hello"),
				})).GetAttr("nested_attr"),
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("goodbye"),
				})).GetAttr("nested_attr"),
			),
		},
		"full flatmap path does not exist": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								ExactlyOneOf: []string{"attr_a.0.nested_attr"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			actualMarkedForNullification := cty.NewPathSet()
			_, err := cty.Transform(tc.ctyVal, func(path cty.Path, val cty.Value) (cty.Value, error) {
				if val.IsNull() {
					return val, nil
				}

				if len(path) == 0 {
					return val, nil
				}

				// find the attribute or block schema representing the value
				attr := schemaMap(tc.Schema).AttributeByPath(path)
				block := schemaMap(tc.Schema).BlockByPath(path)

				if attr == nil && block == nil {
					return val, nil
				}
				var schema *Schema
				if attr != nil {
					schema = attr
				} else {
					schema = block
				}

				markedfordeletion, err := processExactlyOneOf(schema.ExactlyOneOf, tc.ctyVal, path)
				if err != nil {
					return cty.Value{}, err
				}
				for _, p := range markedfordeletion.List() {
					actualMarkedForNullification.Add(p)
				}

				return val, nil

			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actualMarkedForNullification, tc.expectedMarkedForNullification, pathSetComparer); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestProcessRequiredWith(t *testing.T) {
	cases := map[string]struct {
		Schema                         map[string]*Schema
		ctyVal                         cty.Value
		expectedMarkedForNullification cty.PathSet
	}{
		"primitive attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeString,
					Optional: true,
				},
				"attr_b": {
					Type:         TypeString,
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.String),
				"attr_b": cty.StringVal("val"),
			}),
			expectedMarkedForNullification: cty.NewPathSet(cty.GetAttrPath("attr_b")),
		},
		"list attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeList,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.List(cty.String)),
				"attr_b": cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeSet,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeSet,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.Set(cty.String)),
				"attr_b": cty.SetVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"map attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeMap,
					Elem:     &Schema{Type: TypeString},
					Optional: true,
				},
				"attr_b": {
					Type:         TypeMap,
					Elem:         &Schema{Type: TypeString},
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.Map(cty.String)),
				"attr_b": cty.MapVal(map[string]cty.Value{
					"key_a": cty.StringVal("true"),
					"key_b": cty.StringVal("false")}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"list block with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:         TypeList,
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
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
				"attr_a": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"set block with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:         TypeSet,
					Optional:     true,
					RequiredWith: []string{"attr_a", "attr_b"},
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
				"attr_a": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b"),
			),
		},
		"single nested block attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								RequiredWith: []string{"attr_a.0.nested_attr"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.NullVal(cty.String),
						}),
					}),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
		"set nested block attribute with RequiredWith": {
			Schema: map[string]*Schema{
				"attr_a": {
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
				"attr_b": {
					Type:     TypeSet,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								RequiredWith: []string{"attr_a"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
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
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("hello"),
				})).GetAttr("nested_attr"),
				cty.GetAttrPath("attr_b").Index(cty.ObjectVal(map[string]cty.Value{
					"nested_attr": cty.StringVal("goodbye"),
				})).GetAttr("nested_attr"),
			),
		},
		"two of three attributes set": {
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
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"whitelist":  cty.BoolVal(true),
				"purplelist": cty.BoolVal(true),
			}),
			expectedMarkedForNullification: cty.NewPathSet(),
		},
		"full flatmap path does not exist": {
			Schema: map[string]*Schema{
				"attr_a": {
					Type:     TypeList,
					MaxItems: 1,
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
				"attr_b": {
					Type:     TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &Resource{
						Schema: map[string]*Schema{
							"nested_attr": {
								RequiredWith: []string{"attr_a.0.nested_attr"},
								Type:         TypeString,
								Optional:     true,
							},
						},
					},
				},
			},
			ctyVal: cty.ObjectVal(map[string]cty.Value{
				"attr_a": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
					"nested_attr": cty.String,
				}))),
				"attr_b": cty.ListVal(
					[]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"nested_attr": cty.StringVal("hello"),
						}),
					}),
			}),
			expectedMarkedForNullification: cty.NewPathSet(
				cty.GetAttrPath("attr_b").IndexInt(0).GetAttr("nested_attr"),
			),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			actualMarkedForNullification := cty.NewPathSet()
			_, err := cty.Transform(tc.ctyVal, func(path cty.Path, val cty.Value) (cty.Value, error) {
				if val.IsNull() {
					return val, nil
				}

				if len(path) == 0 {
					return val, nil
				}

				// find the attribute or block schema representing the value
				attr := schemaMap(tc.Schema).AttributeByPath(path)
				block := schemaMap(tc.Schema).BlockByPath(path)

				if attr == nil && block == nil {
					return val, nil
				}
				var schema *Schema
				if attr != nil {
					schema = attr
				} else {
					schema = block
				}

				markedfordeletion, err := processRequiredWith(schema.RequiredWith, tc.ctyVal, path)
				if err != nil {
					return cty.Value{}, err
				}
				for _, p := range markedfordeletion.List() {
					actualMarkedForNullification.Add(p)
				}

				return val, nil

			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actualMarkedForNullification, tc.expectedMarkedForNullification, pathSetComparer); diff != "" {
				t.Error(diff)
			}
		})
	}
}
