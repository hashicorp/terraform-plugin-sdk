package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	proto "github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
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
		Cons func([]*proto.Diagnostic) []*proto.Diagnostic
		Want []diagFlat
	}{
		"nil": {
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				return diags
			},
			nil,
		},
		"error": {
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				return append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				return append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				return append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityWarning,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				return append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityWarning,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				diags = append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
					Summary:  "first error",
				}, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				diags = append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityWarning,
					Summary:  "warning",
				}, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				diags = append(diags, &proto.Diagnostic{
					Severity: proto.DiagnosticSeverityError,
					Summary:  "error",
					Detail:   "error detail",
					Attribute: &tftypes.AttributePath{
						Steps: []tftypes.AttributePathStep{

							tftypes.AttributeName("attribute_name"),
						},
					},
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
			func(diags []*proto.Diagnostic) []*proto.Diagnostic {
				diags = append(diags,
					&proto.Diagnostic{
						Severity: proto.DiagnosticSeverityError,
						Summary:  "error 1",
						Detail:   "error 1 detail",
						Attribute: &tftypes.AttributePath{
							Steps: []tftypes.AttributePathStep{
								tftypes.AttributeName("attr"),
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.DiagnosticSeverityError,
						Summary:  "error 2",
						Detail:   "error 2 detail",
						Attribute: &tftypes.AttributePath{
							Steps: []tftypes.AttributePathStep{
								tftypes.AttributeName("attr"),
								tftypes.AttributeName("sub"),
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.DiagnosticSeverityWarning,
						Summary:  "warning",
						Detail:   "warning detail",
						Attribute: &tftypes.AttributePath{
							Steps: []tftypes.AttributePathStep{
								tftypes.AttributeName("attr"),
								tftypes.ElementKeyInt(1),
								tftypes.AttributeName("sub"),
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.DiagnosticSeverityError,
						Summary:  "error 3",
						Detail:   "error 3 detail",
						Attribute: &tftypes.AttributePath{
							Steps: []tftypes.AttributePathStep{
								tftypes.AttributeName("attr"),
								tftypes.ElementKeyString("idx"),
								tftypes.AttributeName("sub"),
							},
						},
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
