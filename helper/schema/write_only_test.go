package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
)

func Test_setWriteOnlyNullValues(t *testing.T) {
	for n, tc := range map[string]struct {
		Schema   *configschema.Block
		Val      cty.Value
		Expected cty.Value
	}{
		"Empty returns no empty object": {
			&configschema.Block{},
			cty.EmptyObjectVal,
			cty.EmptyObjectVal,
		},
		"Top level attributes and block: write only attributes with values": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"required_attribute": {
						Type:     cty.String,
						Required: true,
					},
					"write_only_attribute": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"required_block_attribute": {
									Type:     cty.String,
									Required: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"required_attribute":   cty.StringVal("boop"),
				"write_only_attribute": cty.StringVal("blep"),
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.StringVal("blep"),
					"required_block_attribute":   cty.StringVal("boop"),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"required_attribute":   cty.StringVal("boop"),
				"write_only_attribute": cty.NullVal(cty.String),
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.NullVal(cty.String),
					"required_block_attribute":   cty.StringVal("boop"),
				}),
			}),
		},
		"Top level attributes and block: all null values": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"write_only_attribute1": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"write_only_attribute2": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute1": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"write_only_block_attribute2": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.Object(map[string]cty.Type{
				"write_only_attribute1": cty.String,
				"write_only_attribute2": cty.String,
				"nested_block": cty.Object(map[string]cty.Type{
					"write_only_block_attribute1": cty.String,
					"write_only_block_attribute2": cty.String,
				}),
			})),
			cty.NullVal(cty.Object(map[string]cty.Type{
				"write_only_attribute1": cty.String,
				"write_only_attribute2": cty.String,
				"nested_block": cty.Object(map[string]cty.Type{
					"write_only_block_attribute1": cty.String,
					"write_only_block_attribute2": cty.String,
				}),
			})),
		},
		"Set nested block: write only Nested Attribute": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"required_attribute": {
						Type:     cty.String,
						Required: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"required_attribute": cty.StringVal("boop"),
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("beep"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"required_attribute": cty.StringVal("boop"),
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"Nested single block: write only nested attribute": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"optional_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.StringVal("boop"),
					"optional_attribute":         cty.StringVal("boop"),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.NullVal(cty.String),
					"optional_attribute":         cty.StringVal("boop"),
				}),
			}),
		},
		"Map nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"map_block": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.NullVal(cty.String),
						"write_only_block_attribute": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.StringVal("blep"),
						"write_only_block_attribute": cty.StringVal("boop2"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.NullVal(cty.String),
						"write_only_block_attribute": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.StringVal("blep"),
						"write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"List nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list_block": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("beep"),
						"optional_block_attribute":   cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("boop"),
						"optional_block_attribute":   cty.StringVal("blep"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.NullVal(cty.String),
						"optional_block_attribute":   cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.NullVal(cty.String),
						"optional_block_attribute":   cty.StringVal("blep"),
					}),
				}),
			}),
		},
		"Set nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("blep"),
						"optional_block_attribute":   cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("boop"),
						"optional_block_attribute":   cty.StringVal("boop"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.NullVal(cty.String),
						"optional_block_attribute":   cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.NullVal(cty.String),
						"optional_block_attribute":   cty.StringVal("boop"),
					}),
				}),
			}),
		},
	} {
		t.Run(n, func(t *testing.T) {
			got := setWriteOnlyNullValues(tc.Val, tc.Schema)

			if !got.RawEquals(tc.Expected) {
				t.Errorf("\nexpected: %#v\ngot:      %#v\n", tc.Expected, got)
			}
		})
	}
}

func Test_validateWriteOnlyNullValues(t *testing.T) {
	for n, tc := range map[string]struct {
		Schema   *configschema.Block
		Val      cty.Value
		Expected diag.Diagnostics
	}{
		"Empty returns no diags": {
			&configschema.Block{},
			cty.EmptyObjectVal,
			diag.Diagnostics{},
		},
		"All null values return no diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"write_only_attribute1": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"write_only_attribute2": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"single_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute1": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"write_only_block_attribute2": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.Object(map[string]cty.Type{
				"write_only_attribute1": cty.String,
				"write_only_attribute2": cty.String,
				"single_block": cty.Object(map[string]cty.Type{
					"write_only_block_attribute1": cty.String,
					"write_only_block_attribute2": cty.String,
				}),
			})),
			diag.Diagnostics{},
		},
		"Set nested block WriteOnly attribute with value returns diag": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"write_only_attribute": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"write_only_attribute": cty.StringVal("val"),
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("block_val"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "write_only_attribute"},
					},
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"write_only_block_attribute": cty.StringVal("block_val"),
						})},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"Nested single block, WriteOnly attribute with value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.StringVal("beep"),
					"optional_block_attribute1":  cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "nested_block"},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"Map nested block, WriteOnly attribute with value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"map_block": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_attribute":         cty.NullVal(cty.String),
						"write_only_block_attribute": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_attribute":         cty.StringVal("blep"),
						"write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_block"},
						cty.IndexStep{Key: cty.StringVal("a")},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"List nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list_block": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute": cty.StringVal("blep"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_block"},
						cty.IndexStep{Key: cty.NumberIntVal(0)},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_block"},
						cty.IndexStep{Key: cty.NumberIntVal(1)},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"Set nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute1": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"write_only_block_attribute2": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute1": cty.StringVal("blep"),
						"write_only_block_attribute2": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"write_only_block_attribute1": cty.StringVal("boop"),
						"write_only_block_attribute2": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute1\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"write_only_block_attribute1": cty.StringVal("blep"),
							"write_only_block_attribute2": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "write_only_block_attribute1"},
					},
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute1\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"write_only_block_attribute1": cty.StringVal("boop"),
							"write_only_block_attribute2": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "write_only_block_attribute1"},
					},
				},
			},
		},
		"Map nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"map_block": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.NullVal(cty.String),
						"write_only_block_attribute": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.StringVal("blep"),
						"write_only_block_attribute": cty.StringVal("boop2"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_block"},
						cty.IndexStep{Key: cty.StringVal("a")},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_block"},
						cty.IndexStep{Key: cty.StringVal("b")},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"List nested block, WriteOnly attribute with dynamic value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list_block": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"write_only_block_attribute": {
									Type:      cty.DynamicPseudoType,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":   cty.NullVal(cty.String),
						"write_only_block_attribute": cty.NumberIntVal(8),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_block"},
						cty.IndexStep{Key: cty.NumberIntVal(0)},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
		"multiple nested blocks, multiple WriteOnly attributes with value returns diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block1": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					"nested_block2": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"nested_block1": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.StringVal("beep"),
					"optional_block_attribute1":  cty.StringVal("boop"),
				}),
				"nested_block2": cty.ObjectVal(map[string]cty.Value{
					"write_only_block_attribute": cty.StringVal("beep"),
					"optional_block_attribute1":  cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "nested_block1"},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail: "The resource contains a non-null value for WriteOnly attribute \"write_only_block_attribute\" " +
						"Write-only attributes are only supported in Terraform 1.11 and later.",

					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "nested_block2"},
						cty.GetAttrStep{Name: "write_only_block_attribute"},
					},
				},
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			got := validateWriteOnlyNullValues("test_resource", tc.Val, tc.Schema, cty.Path{})

			if diff := cmp.Diff(got, tc.Expected,
				cmp.AllowUnexported(cty.GetAttrStep{}, cty.IndexStep{}),
				cmp.Comparer(indexStepComparer)); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func Test_validateWriteOnlyRequiredValues(t *testing.T) {
	for n, tc := range map[string]struct {
		Schema   *configschema.Block
		Val      cty.Value
		Expected diag.Diagnostics
	}{
		"Empty returns no diags": {
			&configschema.Block{},
			cty.EmptyObjectVal,
			diag.Diagnostics{},
		},
		"All Required + WriteOnly with values return no diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"required_write_only_attribute1": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
					"required_write_only_attribute2": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"required_write_only_block_attribute1": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"required_write_only_block_attribute2": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"required_write_only_attribute1": cty.StringVal("boop"),
				"required_write_only_attribute2": cty.StringVal("blep"),
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"required_write_only_block_attribute1": cty.StringVal("blep"),
					"required_write_only_block_attribute2": cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{},
		},
		"All Optional + WriteOnly with null values return no diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"optional_write_only_attribute1": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"optional_write_only_attribute2": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_write_only_block_attribute1": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"optional_write_only_block_attribute2": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.Object(map[string]cty.Type{
				"optional_write_only_attribute1": cty.String,
				"optional_write_only_attribute2": cty.String,
				"nested_block": cty.Object(map[string]cty.Type{
					"optional_write_only_block_attribute1": cty.String,
					"optional_write_only_block_attribute2": cty.String,
				}),
			})),
			diag.Diagnostics{},
		},
		"Set nested block Required + WriteOnly attribute with null return diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"required_write_only_attribute": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"required_write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"required_write_only_attribute": cty.NullVal(cty.String),
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_attribute\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
			},
		},
		"Nested single block, Required + WriteOnly attribute with null returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"nested_block": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"optional_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"nested_block": cty.ObjectVal(map[string]cty.Value{
					"optional_attribute": cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"write_only_block_attribute\"",
				},
			},
		},
		"Map nested block, Required + WriteOnly attribute with null value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"map_block": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"required_write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":            cty.NullVal(cty.String),
						"required_write_only_block_attribute": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":            cty.StringVal("blep"),
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
			},
		},
		"List nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"list_block": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"required_write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"list_block": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute": cty.StringVal("blep"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
			},
		},
		"Set nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"set_block": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_write_only_block_attribute": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"required_write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"set_block": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"optional_write_only_block_attribute": cty.StringVal("blep"),
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"optional_write_only_block_attribute": cty.StringVal("boop"),
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
			},
		},
		"Map nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"map_block": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional_block_attribute": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"required_write_only_block_attribute": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"map_block": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":            cty.NullVal(cty.String),
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"optional_block_attribute":            cty.StringVal("blep"),
						"required_write_only_block_attribute": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The resource contains a null value for Required WriteOnly attribute \"required_write_only_block_attribute\"",
				},
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			got := validateWriteOnlyRequiredValues("test_resource", tc.Val, tc.Schema)

			if diff := cmp.Diff(got, tc.Expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func indexStepComparer(step cty.IndexStep, other cty.IndexStep) bool {
	return true
}
