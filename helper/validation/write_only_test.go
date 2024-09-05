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
		oldAttributePath       cty.Path
		writeOnlyAttributePath cty.Path
		validateConfigReq      schema.ValidateResourceConfigRequest
		expectedDiags          diag.Diagnostics
	}{
		"writeOnlyAttributeAllowed unset returns no diags": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: false,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NullVal(cty.Number),
				}),
			},
		},
		"oldAttribute and writeOnlyAttribute set returns no diags": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
		},
		"writeOnlyAttribute set returns no diags": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NullVal(cty.Number),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
		},
		"oldAttributePath pointing to missing attribute returns error diag": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid oldAttributePath",
					Detail:        "Encountered an error when applying the specified oldAttribute path, original error: object has no attribute \"oldAttribute\"",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "oldAttribute"}},
				},
			},
		},
		"writeOnlyAttributePath pointing to missing attribute returns error diag": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid writeOnlyAttributePath",
					Detail:        "Encountered an error when applying the specified writeOnlyAttribute path, original error: object has no attribute \"writeOnlyAttribute\"",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "writeOnlyAttribute"}},
				},
			},
		},
		"oldAttributePath with empty path returns error diag": {
			oldAttributePath:       cty.Path{},
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid oldAttributePath",
					Detail:        "The specified oldAttribute path must point to an attribute",
					AttributePath: cty.Path{},
				},
			},
		},
		"writeOnlyAttributePath with empty path returns error diag": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.Path{},
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"oldAttribute":       cty.NumberIntVal(42),
					"writeOnlyAttribute": cty.NumberIntVal(42),
				}),
			},
			expectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Invalid writeOnlyAttributePath",
					Detail:        "The specified writeOnlyAttribute path must point to an attribute",
					AttributePath: cty.Path{},
				},
			},
		},
		"only oldAttribute set returns warning diag": {
			oldAttributePath:       cty.GetAttrPath("oldAttribute"),
			writeOnlyAttributePath: cty.GetAttrPath("writeOnlyAttribute"),
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
		"block: oldAttribute and writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "writeOnlyAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
				WriteOnlyAttributesAllowed: true,
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					"id": cty.NullVal(cty.String),
					"config_block_attr": cty.ObjectVal(map[string]cty.Value{
						"oldAttribute":       cty.StringVal("value"),
						"writeOnlyAttribute": cty.StringVal("value"),
					}),
				}),
			},
		},
		"block: writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "writeOnlyAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "writeOnlyAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
		"set nested block: oldAttribute and writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"oldAttribute":       cty.StringVal("value"),
						"writeOnlyAttribute": cty.StringVal("value"),
					}),
				})},
				cty.IndexStep{Key: cty.StringVal("oldAttribute")},
			},
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "set_nested_block"},
				cty.IndexStep{Key: cty.NumberIntVal(0)},
				cty.IndexStep{Key: cty.StringVal("writeOnlyAttribute")},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
		},
		"set nested block: writeOnlyAttribute set returns no diags": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "writeOnlyAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
		"set nested block: only oldAttribute set returns warning diag": {
			oldAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "oldAttribute"},
			},
			writeOnlyAttributePath: cty.Path{
				cty.GetAttrStep{Name: "config_block_attr"},
				cty.GetAttrStep{Name: "writeOnlyAttribute"},
			},
			validateConfigReq: schema.ValidateResourceConfigRequest{
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := PreferWriteOnlyAttribute(tc.oldAttributePath, tc.writeOnlyAttributePath)

			actual := &schema.ValidateResourceConfigResponse{}
			f(context.Background(), tc.validateConfigReq, actual)

			if len(actual.Diagnostics) == 0 && tc.expectedDiags == nil {
				return
			}

			if len(actual.Diagnostics) != 0 && tc.expectedDiags == nil {
				t.Fatalf("expected no diagnostics but got %v", actual.Diagnostics)
			}

			if diff := cmp.Diff(tc.expectedDiags, actual.Diagnostics, cmp.AllowUnexported(cty.GetAttrStep{})); diff != "" {
				t.Errorf("Unexpected diagnostics (-wanted +got): %s", diff)
			}
		})
	}
}
