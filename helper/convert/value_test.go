package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimitiveTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.StringVal("test"),
			Want:  tftypes.NewValue(tftypes.String, "test"),
		},
		{
			Value: cty.BoolVal(true),
			Want:  tftypes.NewValue(tftypes.Bool, true),
		},
		{
			Value: cty.NumberIntVal(42),
			Want:  tftypes.NewValue(tftypes.Number, 42),
		},
		// TODO other number types
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got := PrimitiveTfValue(test.Value)

			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestListTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.ListVal([]cty.Value{
				cty.StringVal("apple"),
				cty.StringVal("cherry"),
				cty.StringVal("kangaroo"),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "apple"),
				tftypes.NewValue(tftypes.String, "cherry"),
				tftypes.NewValue(tftypes.String, "kangaroo"),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.BoolVal(true),
				cty.BoolVal(false),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(100),
				cty.NumberIntVal(200),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 100),
				tftypes.NewValue(tftypes.Number, 200),
			}),
		},
		// TODO other number types
		{
			Value: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Alice"),
					"breed":  cty.StringVal("Beagle"),
					"weight": cty.NumberIntVal(20),
					"toys":   cty.ListVal([]cty.Value{cty.StringVal("ball"), cty.StringVal("rope")}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Bobby"),
					"breed":  cty.StringVal("Golden"),
					"weight": cty.NumberIntVal(30),
					"toys":   cty.ListVal([]cty.Value{cty.StringVal("dummy"), cty.StringVal("frisbee")}),
				}),
			}),
			Want: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Alice"),
					"breed":  tftypes.NewValue(tftypes.String, "Beagle"),
					"weight": tftypes.NewValue(tftypes.Number, 20),
					"toys": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "ball"),
						tftypes.NewValue(tftypes.String, "rope"),
					}),
				}),
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Bobby"),
					"breed":  tftypes.NewValue(tftypes.String, "Golden"),
					"weight": tftypes.NewValue(tftypes.Number, 30),
					"toys": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "dummy"),
						tftypes.NewValue(tftypes.String, "frisbee"),
					}),
				}),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"enforcement": cty.NullVal(cty.String),
				}),
			}),
			Want: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"enforcement": tftypes.String,
					},
				},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"enforcement": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"enforcement": tftypes.NewValue(tftypes.String, nil),
				}),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got := ListTfValue(test.Value)

			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

// TODO more tests
