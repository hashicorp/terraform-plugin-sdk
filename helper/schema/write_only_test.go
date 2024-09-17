package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
)

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
					"foo": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"bar": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"biz": {
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
				"foo": cty.String,
				"bar": cty.String,
				"baz": cty.Object(map[string]cty.Type{
					"boz": cty.String,
					"biz": cty.String,
				}),
			})),
			diag.Diagnostics{},
		},
		"Set nested block WriteOnly attribute with value returns diag": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
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
				"foo": cty.StringVal("foo_val"),
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.StringVal("blep"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"foo\"",
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"boz\"",
				},
			},
		},
		"Nested single block, WriteOnly attribute with value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.StringVal("beep"),
					"baz": cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"bar\"",
				},
			},
		},
		"Map nested block, WriteOnly attribute with value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"baz\"",
				},
			},
		},
		"List nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"bar\"",
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"bar\"",
				},
			},
		},
		"Set nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("boop"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"bar\"",
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"bar\"",
				},
			},
		},
		"Map nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.StringVal("boop2"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"baz\"",
				},
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"baz\"",
				},
			},
		},
		"List nested block, WriteOnly attribute with dynamic value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NumberIntVal(8),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "WriteOnly Attribute Not Allowed",
					Detail:   "The \"test_resource\" resource contains a non-null value for WriteOnly attribute \"baz\"",
				},
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			got := validateWriteOnlyNullValues("test_resource", tc.Val, tc.Schema)

			if diff := cmp.Diff(got, tc.Expected); diff != "" {
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
					"foo": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
					"bar": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"biz": {
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
				"foo": cty.StringVal("boop"),
				"bar": cty.StringVal("blep"),
				"baz": cty.ObjectVal(map[string]cty.Value{
					"boz": cty.StringVal("blep"),
					"biz": cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{},
		},
		"All Optional + WriteOnly with null values return no diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"bar": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"biz": {
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
				"foo": cty.String,
				"bar": cty.String,
				"baz": cty.Object(map[string]cty.Type{
					"boz": cty.String,
					"biz": cty.String,
				}),
			})),
			diag.Diagnostics{},
		},
		"Set nested block Required + WriteOnly attribute with null return diags": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
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
				"foo": cty.NullVal(cty.String),
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"foo\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"boz\"",
				},
			},
		},
		"Nested single block, Required + WriteOnly attribute with null returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ObjectVal(map[string]cty.Value{
					"baz": cty.StringVal("boop"),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"bar\"",
				},
			},
		},
		"Map nested block, Required + WriteOnly attribute with null value returns diag": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"baz\"",
				},
			},
		},
		"List nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("blep"),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"bar\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"bar\"",
				},
			},
		},
		"Set nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("boop"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"baz\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"baz\"",
				},
			},
		},
		"Map nested block, WriteOnly attribute with multiple values returns multiple diags": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"baz\"",
				},
				{
					Severity: diag.Error,
					Summary:  "Required WriteOnly Attribute",
					Detail:   "The \"test_resource\" resource contains a null value for Required WriteOnly attribute \"baz\"",
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
					"foo": {
						Type:     cty.String,
						Required: true,
					},
					"bar": {
						Type:      cty.String,
						Required:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"biz": {
									Type:     cty.String,
									Required: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("boop"),
				"bar": cty.StringVal("blep"),
				"baz": cty.ObjectVal(map[string]cty.Value{
					"boz": cty.StringVal("blep"),
					"biz": cty.StringVal("boop"),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("boop"),
				"bar": cty.NullVal(cty.String),
				"baz": cty.ObjectVal(map[string]cty.Value{
					"boz": cty.NullVal(cty.String),
					"biz": cty.StringVal("boop"),
				}),
			}),
		},
		"Top level attributes and block: all null values": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
					"bar": {
						Type:      cty.String,
						Optional:  true,
						WriteOnly: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"biz": {
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
				"foo": cty.String,
				"bar": cty.String,
				"baz": cty.Object(map[string]cty.Type{
					"boz": cty.String,
					"biz": cty.String,
				}),
			})),
			cty.NullVal(cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.String,
				"baz": cty.Object(map[string]cty.Type{
					"boz": cty.String,
					"biz": cty.String,
				}),
			})),
		},
		"Set nested block: write only Nested Attribute": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:     cty.String,
						Required: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
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
				"foo": cty.StringVal("boop"),
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.StringVal("beep"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("boop"),
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"Nested single block: write only nested attribute": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.StringVal("boop"),
					"baz": cty.StringVal("boop"),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.NullVal(cty.String),
					"baz": cty.StringVal("boop"),
				}),
			}),
		},
		"Map nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.StringVal("boop2"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"List nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Required:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("boop"),
						"baz": cty.StringVal("blep"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("bap"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.StringVal("blep"),
					}),
				}),
			}),
		},
		"Set nested block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:      cty.String,
									Optional:  true,
									WriteOnly: true,
								},
								"baz": {
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
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("boop"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"Set nested Map block: multiple write only nested attributes": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingMap,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"baz": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.NullVal(cty.String),
						"baz": cty.NullVal(cty.String),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
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
