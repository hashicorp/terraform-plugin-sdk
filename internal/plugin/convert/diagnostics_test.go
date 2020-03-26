package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
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
					Severity: proto.Diagnostic_ERROR,
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
					Severity: proto.Diagnostic_ERROR,
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
					Severity: proto.Diagnostic_WARNING,
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
					Severity: proto.Diagnostic_WARNING,
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
					Severity: proto.Diagnostic_ERROR,
					Summary:  "first error",
				}, &proto.Diagnostic{
					Severity: proto.Diagnostic_ERROR,
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
					Severity: proto.Diagnostic_WARNING,
					Summary:  "warning",
				}, &proto.Diagnostic{
					Severity: proto.Diagnostic_ERROR,
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
					Severity: proto.Diagnostic_ERROR,
					Summary:  "error",
					Detail:   "error detail",
					Attribute: &proto.AttributePath{
						Steps: []*proto.AttributePath_Step{
							{
								Selector: &proto.AttributePath_Step_AttributeName{
									AttributeName: "attribute_name",
								},
							},
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
						Severity: proto.Diagnostic_ERROR,
						Summary:  "error 1",
						Detail:   "error 1 detail",
						Attribute: &proto.AttributePath{
							Steps: []*proto.AttributePath_Step{
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "attr",
									},
								},
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.Diagnostic_ERROR,
						Summary:  "error 2",
						Detail:   "error 2 detail",
						Attribute: &proto.AttributePath{
							Steps: []*proto.AttributePath_Step{
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "attr",
									},
								},
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "sub",
									},
								},
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.Diagnostic_WARNING,
						Summary:  "warning",
						Detail:   "warning detail",
						Attribute: &proto.AttributePath{
							Steps: []*proto.AttributePath_Step{
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "attr",
									},
								},
								{
									Selector: &proto.AttributePath_Step_ElementKeyInt{
										ElementKeyInt: 1,
									},
								},
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "sub",
									},
								},
							},
						},
					},
					&proto.Diagnostic{
						Severity: proto.Diagnostic_ERROR,
						Summary:  "error 3",
						Detail:   "error 3 detail",
						Attribute: &proto.AttributePath{
							Steps: []*proto.AttributePath_Step{
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "attr",
									},
								},
								{
									Selector: &proto.AttributePath_Step_ElementKeyString{
										ElementKeyString: "idx",
									},
								},
								{
									Selector: &proto.AttributePath_Step_AttributeName{
										AttributeName: "sub",
									},
								},
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
			// we take the
			diags := ProtoToDiags(tc.Cons(nil))

			flat := flattenDiags(diags)

			if !cmp.Equal(flat, tc.Want, typeComparer, valueComparer, equateEmpty) {
				t.Fatal(cmp.Diff(flat, tc.Want, typeComparer, valueComparer, equateEmpty))
			}
		})
	}
}
