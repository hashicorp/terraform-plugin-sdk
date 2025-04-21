// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestDiagnostics(t *testing.T) {
	type diagFlat struct {
		Severity diag.Severity
		Attr     []interface{}
		Summary  string
		Detail   string
	}

	tests := map[string]struct {
		Cons func([]*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic
		Want []diagFlat
	}{
		"nil": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				return diags
			},
			nil,
		},
		"error": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				return append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "simple error",
				})
			},
			[]diagFlat{
				{
					Severity: diag.Error,
					Summary:  "simple error",
				},
			},
		},
		"detailed error": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				return append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "simple error",
					Detail:   "detailed error",
				})
			},
			[]diagFlat{
				{
					Severity: diag.Error,
					Summary:  "simple error",
					Detail:   "detailed error",
				},
			},
		},
		"warning": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				return append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityWarning,
					Summary:  "simple warning",
				})
			},
			[]diagFlat{
				{
					Severity: diag.Warning,
					Summary:  "simple warning",
				},
			},
		},
		"detailed warning": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				return append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityWarning,
					Summary:  "simple warning",
					Detail:   "detailed warning",
				})
			},
			[]diagFlat{
				{
					Severity: diag.Warning,
					Summary:  "simple warning",
					Detail:   "detailed warning",
				},
			},
		},
		"multi error": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				diags = append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "first error",
				}, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "second error",
				})
				return diags
			},
			[]diagFlat{
				{
					Severity: diag.Error,
					Summary:  "first error",
				},
				{
					Severity: diag.Error,
					Summary:  "second error",
				},
			},
		},
		"warning and error": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				diags = append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityWarning,
					Summary:  "warning",
				}, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "error",
				})
				return diags
			},
			[]diagFlat{
				{
					Severity: diag.Warning,
					Summary:  "warning",
				},
				{
					Severity: diag.Error,
					Summary:  "error",
				},
			},
		},
		"attr error": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				diags = append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "error",
					Detail:   "error detail",
					Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
						tftypes.AttributeName("attribute_name"),
					}),
				})
				return diags
			},
			[]diagFlat{
				{
					Severity: diag.Error,
					Summary:  "error",
					Detail:   "error detail",
					Attr:     []interface{}{"attribute_name"},
				},
			},
		},
		"multi attr": {
			func(diags []*tfprotov5.Diagnostic) []*tfprotov5.Diagnostic {
				diags = append(diags,
					&tfprotov5.Diagnostic{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error 1",
						Detail:   "error 1 detail",
						Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
							tftypes.AttributeName("attr"),
						}),
					},
					&tfprotov5.Diagnostic{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error 2",
						Detail:   "error 2 detail",
						Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
							tftypes.AttributeName("attr"),
							tftypes.AttributeName("sub"),
						}),
					},
					&tfprotov5.Diagnostic{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning",
						Detail:   "warning detail",
						Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
							tftypes.AttributeName("attr"),
							tftypes.ElementKeyInt(1),
							tftypes.AttributeName("sub"),
						}),
					},
					&tfprotov5.Diagnostic{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error 3",
						Detail:   "error 3 detail",
						Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
							tftypes.AttributeName("attr"),
							tftypes.ElementKeyString("idx"),
							tftypes.AttributeName("sub"),
						}),
					},
				)

				return diags
			},
			[]diagFlat{
				{
					Severity: diag.Error,
					Summary:  "error 1",
					Detail:   "error 1 detail",
					Attr:     []interface{}{"attr"},
				},
				{
					Severity: diag.Error,
					Summary:  "error 2",
					Detail:   "error 2 detail",
					Attr:     []interface{}{"attr", "sub"},
				},
				{
					Severity: diag.Warning,
					Summary:  "warning",
					Detail:   "warning detail",
					Attr:     []interface{}{"attr", 1, "sub"},
				},
				{
					Severity: diag.Error,
					Summary:  "error 3",
					Detail:   "error 3 detail",
					Attr:     []interface{}{"attr", "idx", "sub"},
				},
			},
		},
	}

	flattenDiags := func(ds diag.Diagnostics) []diagFlat {
		var flat []diagFlat
		for _, item := range ds {

			var attr []interface{}

			for _, a := range item.AttributePath {
				switch step := a.(type) {
				case cty.GetAttrStep:
					attr = append(attr, step.Name)
				case cty.IndexStep:
					switch step.Key.Type() {
					case cty.Number:
						i, _ := step.Key.AsBigFloat().Int64()
						attr = append(attr, int(i))
					case cty.String:
						attr = append(attr, step.Key.AsString())
					}
				}
			}

			flat = append(flat, diagFlat{
				Severity: item.Severity,
				Attr:     attr,
				Summary:  item.Summary,
				Detail:   item.Detail,
			})
		}
		return flat
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			diags := ProtoToDiags(tc.Cons(nil))

			flat := flattenDiags(diags)

			if !cmp.Equal(flat, tc.Want, typeComparer, valueComparer, equateEmpty) {
				t.Fatal(cmp.Diff(flat, tc.Want, typeComparer, valueComparer, equateEmpty))
			}
		})
	}
}

func TestPathToAttributePath(t *testing.T) {
	tests := map[string]struct {
		path cty.Path
		want *tftypes.AttributePath
	}{
		"no steps": {
			path: cty.Path{},
			want: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := PathToAttributePath(tc.path)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Unexpected diff: %s", diff)
			}
		})
	}
}
