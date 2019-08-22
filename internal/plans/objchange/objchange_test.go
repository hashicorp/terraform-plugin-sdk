package objchange

import (
	"testing"

	"github.com/apparentlymart/go-dump/dump"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/configschema"
)

func TestProposedNewObject(t *testing.T) {
	tests := map[string]struct {
		Schema *configschema.Block
		Prior  cty.Value
		Config cty.Value
		Want   cty.Value
	}{
		"empty": {
			&configschema.Block{},
			cty.EmptyObjectVal,
			cty.EmptyObjectVal,
			cty.EmptyObjectVal,
		},
		"no prior": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:     cty.String,
						Optional: true,
					},
					"bar": {
						Type:     cty.String,
						Computed: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
								"biz": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.DynamicPseudoType),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello"),
				"bar": cty.NullVal(cty.String),
				"baz": cty.ObjectVal(map[string]cty.Value{
					"boz": cty.StringVal("world"),

					// An unknown in the config represents a situation where
					// an argument is explicitly set to an expression result
					// that is derived from an unknown value. This is distinct
					// from leaving it null, which allows the provider itself
					// to decide the value during PlanResourceChange.
					"biz": cty.UnknownVal(cty.String),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello"),

				// unset computed attributes are null in the proposal; provider
				// usually changes them to "unknown" during PlanResourceChange,
				// to indicate that the value will be decided during apply.
				"bar": cty.NullVal(cty.String),

				"baz": cty.ObjectVal(map[string]cty.Value{
					"boz": cty.StringVal("world"),
					"biz": cty.UnknownVal(cty.String), // explicit unknown preserved from config
				}),
			}),
		},
		"null block remains null": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:     cty.String,
						Optional: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.DynamicPseudoType),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"baz": cty.NullVal(cty.Object(map[string]cty.Type{
					"boz": cty.String,
				})),
			}),
			// The baz block does not exist in the config, and therefore
			// shouldn't be planned.
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"baz": cty.NullVal(cty.Object(map[string]cty.Type{
					"boz": cty.String,
				})),
			}),
		},
		"no prior with set": {
			// This one is here because our handling of sets is more complex
			// than others (due to the fuzzy correlation heuristic) and
			// historically that caused us some panic-related grief.
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"baz": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"boz": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.DynamicPseudoType),
			cty.ObjectVal(map[string]cty.Value{
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.StringVal("world"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"baz": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"boz": cty.StringVal("world"),
					}),
				}),
			}),
		},
		"prior attributes": {
			&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"foo": {
						Type:     cty.String,
						Optional: true,
					},
					"bar": {
						Type:     cty.String,
						Computed: true,
					},
					"baz": {
						Type:     cty.String,
						Optional: true,
						Computed: true,
					},
					"boz": {
						Type:     cty.String,
						Optional: true,
						Computed: true,
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bonjour"),
				"bar": cty.StringVal("petit dejeuner"),
				"baz": cty.StringVal("grande dejeuner"),
				"boz": cty.StringVal("a la monde"),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello"),
				"bar": cty.NullVal(cty.String),
				"baz": cty.NullVal(cty.String),
				"boz": cty.StringVal("world"),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("hello"),
				"bar": cty.StringVal("petit dejeuner"),
				"baz": cty.StringVal("grande dejeuner"),
				"boz": cty.StringVal("world"),
			}),
		},
		"prior nested single": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSingle,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
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
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.StringVal("bap"),
					"baz": cty.NullVal(cty.String),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.StringVal("bap"),
					"baz": cty.StringVal("boop"),
				}),
			}),
		},
		"prior nested list": {
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
						"baz": cty.StringVal("boop"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.StringVal("boop"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"prior nested list with dynamic": {
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
									Type:     cty.DynamicPseudoType,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("boop"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.StringVal("boop"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"prior nested map": {
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
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.StringVal("boot"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.NullVal(cty.String),
					}),
					"c": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.StringVal("boop"),
					}),
					"c": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"prior nested map with dynamic": {
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
									Type:     cty.DynamicPseudoType,
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
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("boop"),
					}),
					"b": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.StringVal("boot"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.NullVal(cty.String),
					}),
					"c": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bap"),
						"baz": cty.StringVal("boop"),
					}),
					"c": cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"prior nested set": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"foo": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"bar": {
									// This non-computed attribute will serve
									// as our matching key for propagating
									// "baz" from elements in the prior value.
									Type:     cty.String,
									Optional: true,
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
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("boop"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("blep"),
						"baz": cty.StringVal("boot"),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.NullVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"foo": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("beep"),
						"baz": cty.StringVal("boop"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"bar": cty.StringVal("bosh"),
						"baz": cty.NullVal(cty.String),
					}),
				}),
			}),
		},
		"sets differing only by unknown": {
			&configschema.Block{
				BlockTypes: map[string]*configschema.NestedBlock{
					"multi": {
						Nesting: configschema.NestingSet,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"optional": {
									Type:     cty.String,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			cty.NullVal(cty.DynamicPseudoType),
			cty.ObjectVal(map[string]cty.Value{
				"multi": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"optional": cty.UnknownVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"optional": cty.UnknownVal(cty.String),
					}),
				}),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"multi": cty.SetVal([]cty.Value{
					// These remain distinct because unknown values never
					// compare equal. They may be consolidated together once
					// the values become known, though.
					cty.ObjectVal(map[string]cty.Value{
						"optional": cty.UnknownVal(cty.String),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"optional": cty.UnknownVal(cty.String),
					}),
				}),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := ProposedNewObject(test.Schema, test.Prior, test.Config)
			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %swant: %s", dump.Value(got), dump.Value(test.Want))
			}
		})
	}
}
