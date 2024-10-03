// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestPreferWriteOnlyAttribute(t *testing.T) {
	cases := map[string]struct {
		oldAttributePath  cty.Path
		validateConfigReq schema.ValidateResourceConfigFuncRequest
		expectedDiags     diag.Diagnostics
	}{
		"writeOnlyAttributeAllowed set to false with oldAttribute set returns no diags": {
			oldAttributePath: cty.GetAttrPath("oldAttribute"),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: false,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NullVal(cty.Number),
				}),
			},
		},
		"invalid oldAttributePath returns error diag": {
			oldAttributePath: cty.GetAttrPath("oldAttribute").Index(cty.UnknownVal(cty.Number)),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute": cty.ListVal([]cty.Value{
						cty.StringVal("val1"),
						cty.StringVal("val2"),
					}),
					"writeOnlyAttribute": cty.NullVal(cty.Number),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Invalid oldAttribute path",
					Detail: "The Terraform Provider unexpectedly provided a path that does not match the current schema. " +
						"This can happen if the path does not correctly follow the schema in structure or types. " +
						"Please report this to the provider developers. \n\n" +
						"The oldAttribute path provided is invalid. The last step in the path must be a cty.GetAttrStep{}",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "oldAttribute"},
						cty.IndexStep{
							Key: cty.NumberIntVal(0),
						},
					},
				},
			},
		},
		"oldAttribute and writeOnlyAttribute set returns warning diags": {
			oldAttributePath: cty.GetAttrPath("oldAttribute"),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "oldAttribute"}},
				},
			},
		},
		"writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.GetAttrPath("oldAttribute"),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NullVal(cty.Number),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
		},
		"oldAttributePath pointing to missing attribute returns no diags": {
			oldAttributePath: cty.GetAttrPath("oldAttribute"),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: nil,
		},
		"oldAttributePath with empty path returns no diags": {
			oldAttributePath: cty.Path{},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: nil,
		},
		"only oldAttribute set returns warning diag": {
			oldAttributePath: cty.GetAttrPath("oldAttribute"),
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NullVal(cty.Number),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "oldAttribute"}},
				},
			},
		},
		"block: oldAttribute and writeOnlyAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"config_block_attr": cty.ObjectVal(map[string]cty.Value{
						"oldAttribute":       cty.StringVal("value"),
						"writeOnlyAttribute": cty.StringVal("value"),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "config_block_attr"},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"block: writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"config_block_attr": cty.ObjectVal(map[string]cty.Value{
						"oldAttribute":       cty.NullVal(cty.String),
						"writeOnlyAttribute": cty.StringVal("value"),
					}),
				}),
			},
		},
		"block: only oldAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"config_block_attr": cty.ObjectVal(map[string]cty.Value{
						"oldAttribute":       cty.StringVal("value"),
						"writeOnlyAttribute": cty.NullVal(cty.String),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "config_block_attr"},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"list nested block: oldAttribute and writeOnlyAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "list_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Number)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"list_nested_block": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value"),
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_nested_block"},
						cty.IndexStep{Key: cty.NumberIntVal(0)},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"list nested block: writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "list_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Number)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"list_nested_block": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: nil,
		},
		"list nested block: only oldAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "list_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Number)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"list_nested_block": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_nested_block"},
						cty.IndexStep{Key: cty.NumberIntVal(0)},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"list nested block: multiple oldAttribute set returns multiple warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "list_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Number)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"list_nested_block": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.NullVal(cty.String),
							"writeOnlyAttribute": cty.StringVal("value2"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_nested_block"},
						cty.IndexStep{Key: cty.NumberIntVal(0)},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "list_nested_block"},
						cty.IndexStep{Key: cty.NumberIntVal(2)},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"set nested block: oldAttribute and writeOnlyAttribute set returns warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Object(
					map[string]cty.Type{
						"oldAttribute":       cty.String,
						"writeOnlyAttribute": cty.String,
					},
				))},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"set_nested_block": cty.SetVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value"),
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value"),
							"writeOnlyAttribute": cty.StringVal("value"),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"set nested block: writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Object(
					map[string]cty.Type{
						"oldAttribute": cty.String,
					},
				))},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"set_nested_block": cty.SetVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: nil,
		},
		"set nested block: only oldAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Object(
					map[string]cty.Type{
						"oldAttribute": cty.String,
					},
				))},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"set_nested_block": cty.SetVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute": cty.StringVal("value"),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"set nested block: multiple oldAttribute set returns multiple warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Object(nil))},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"set_nested_block": cty.SetVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.NullVal(cty.String),
							"writeOnlyAttribute": cty.StringVal("value2"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"map nested block: oldAttribute and writeOnlyAttribute map returns warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "map_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"map_nested_block": cty.MapVal(map[string]cty.Value{
						"key1": cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value"),
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key1")},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"map nested block: writeOnlyAttribute map returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "map_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"map_nested_block": cty.MapVal(map[string]cty.Value{
						"key1": cty.ObjectVal(map[string]cty.Value{
							"writeOnlyAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: nil,
		},
		"map nested block: only oldAttribute map returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "map_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"map_nested_block": cty.MapVal(map[string]cty.Value{
						"key1": cty.ObjectVal(map[string]cty.Value{
							"oldAttribute": cty.StringVal("value"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key1")},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"map nested block: multiple oldAttribute map returns multiple warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "map_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"map_nested_block": cty.MapVal(map[string]cty.Value{
						"key1": cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
						"key2": cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.NullVal(cty.String),
							"writeOnlyAttribute": cty.StringVal("value2"),
						}),
						"key3": cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key1")},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key3")},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
		"map nested set nested block: multiple oldAttribute map returns multiple warning diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "map_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.String)},
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.UnknownVal(cty.Object(nil))},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigFuncRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"map_nested_block": cty.MapVal(map[string]cty.Value{
						"key1": cty.ObjectVal(map[string]cty.Value{
							"set_nested_block": cty.SetVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.StringVal("value1"),
									"writeOnlyAttribute": cty.NullVal(cty.String),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.NullVal(cty.String),
									"writeOnlyAttribute": cty.StringVal("value2"),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.StringVal("value3"),
									"writeOnlyAttribute": cty.NullVal(cty.String),
								}),
							}),
							"string_nested_attribute": cty.NullVal(cty.String),
						}),
						"key2": cty.ObjectVal(map[string]cty.Value{
							"set_nested_block": cty.SetVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.NullVal(cty.String),
									"writeOnlyAttribute": cty.StringVal("value2"),
								}),
							}),
							"string_nested_attribute": cty.StringVal("value1"),
						}),
						"key3": cty.ObjectVal(map[string]cty.Value{
							"set_nested_block": cty.SetVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.StringVal("value1"),
									"writeOnlyAttribute": cty.NullVal(cty.String),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.NullVal(cty.String),
									"writeOnlyAttribute": cty.StringVal("value2"),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"oldAttribute":       cty.StringVal("value3"),
									"writeOnlyAttribute": cty.NullVal(cty.String),
								}),
							}),
							"string_nested_attribute": cty.StringVal("value1"),
						}),
					}),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key1")},
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
						},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key1")},
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key3")},
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value1"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						})},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
				{
					Severity: diag.Warning,
					Summary:  "Available Write-Only Attribute Alternative",
					Detail: "The attribute oldAttribute has a WriteOnly version writeOnlyAttribute available. " +
						"Use the WriteOnly version of the attribute when possible.",
					AttributePath: cty.Path{
						cty.GetAttrStep{Name: "map_nested_block"},
						cty.IndexStep{Key: cty.StringVal("key3")},
						cty.GetAttrStep{Name: "set_nested_block"},
						cty.IndexStep{Key: cty.ObjectVal(map[string]cty.Value{
							"oldAttribute":       cty.StringVal("value3"),
							"writeOnlyAttribute": cty.NullVal(cty.String),
						}),
						},
						cty.GetAttrStep{Name: "oldAttribute"},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := PreferWriteOnlyAttribute(tc.oldAttributePath, cty.GetAttrPath("writeOnlyAttribute"))

			actual := &schema.ValidateResourceConfigFuncResponse{}
			f(context.Background(), tc.validateConfigReq, actual)

			if len(actual.Diagnostics) == 0 && tc.expectedDiags == nil {
				return
			}

			if len(actual.Diagnostics) != 0 && tc.expectedDiags == nil {
				t.Fatalf("expected no diagnostics but got %v", actual.Diagnostics)
			}

			if diff := cmp.Diff(tc.expectedDiags, actual.Diagnostics,
				cmp.AllowUnexported(cty.GetAttrStep{}, cty.IndexStep{}),
				cmp.Comparer(indexStepComparer),
			); diff != "" {
				t.Errorf("Unexpected diagnostics (-wanted +got): %s", diff)
			}
		})
	}
}

func indexStepComparer(step cty.IndexStep, other cty.IndexStep) bool {
	return step.Key.RawEquals(other.Key)
}
