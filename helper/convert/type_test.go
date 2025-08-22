package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestToTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Type
		Want  tftypes.Type
	}{
		{
			Value: cty.String,
			Want:  tftypes.String,
		},
		{
			Value: cty.Bool,
			Want:  tftypes.Bool,
		},
		{
			Value: cty.Number,
			Want:  tftypes.Number,
		},
		{
			Value: cty.List(cty.String),
			Want:  tftypes.List{ElementType: tftypes.String},
		},
		{
			Value: cty.List(cty.List(cty.Bool)),
			Want:  tftypes.List{ElementType: tftypes.List{ElementType: tftypes.Bool}},
		},
		{
			Value: cty.Set(cty.Number),
			Want:  tftypes.Set{ElementType: tftypes.Number},
		},
		{
			Value: cty.Tuple([]cty.Type{cty.String, cty.Number}),
			Want:  tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Number}},
		},
		{
			Value: cty.Set(cty.Object(map[string]cty.Type{
				"flavour": cty.String,
				"texture": cty.String,
			})),
			Want: tftypes.Set{ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"flavour": tftypes.String,
					"texture": tftypes.String,
				},
			}},
		},
		{
			Value: cty.Object(map[string]cty.Type{
				"chonk": cty.Bool,
				"blep":  cty.String,
				"mlem":  cty.Number,
				"noms":  cty.List(cty.String),
			}),
			Want: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"chonk": tftypes.Bool,
				"blep":  tftypes.String,
				"mlem":  tftypes.Number,
				"noms":  tftypes.List{ElementType: tftypes.String},
			}},
		},
		{
			Value: cty.Object(map[string]cty.Type{
				"chonk": cty.Object(map[string]cty.Type{
					"size":   cty.String,
					"weight": cty.Number,
				}),
				"blep": cty.List(cty.Object(map[string]cty.Type{
					"color": cty.String,
					"pattern": cty.Object(map[string]cty.Type{
						"type":  cty.Number,
						"style": cty.List(cty.String),
					}),
				})),
				"mlem": cty.Number,
				"noms": cty.List(cty.String),
			}),
			Want: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"chonk": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.String,
						"weight": tftypes.Number,
					},
				},
				"blep": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"color": tftypes.String,
					"pattern": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"type":  tftypes.Number,
						"style": tftypes.List{ElementType: tftypes.String},
					}},
				}}},
				"mlem": tftypes.Number,
				"noms": tftypes.List{ElementType: tftypes.String},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got := ToTfType(test.Value)

			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}
