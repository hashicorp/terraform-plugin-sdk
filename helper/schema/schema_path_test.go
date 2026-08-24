package schema

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Regression test for https://github.com/hashicorp/terraform-plugin-sdk/issues/1175
//
// The validate* helpers built child paths with append(path, step) inside
// loops. Once an ancestor append had grown the backing array beyond its
// length, sibling iterations wrote their step into the same spare slot, so
// every previously stored diag.Diagnostic.AttributePath was rewritten to
// point at the last-visited sibling.
func TestSchemaMapValidate_DiagnosticPathsNotAliased(t *testing.T) {
	failValidation := func(v interface{}, p cty.Path) diag.Diagnostics {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       "always invalid",
			AttributePath: p,
		}}
	}

	m := schemaMap(map[string]*Schema{
		"block": {
			Type:     TypeList,
			Required: true,
			Elem: &Resource{
				Schema: map[string]*Schema{
					"inner": {
						Type:     TypeList,
						Optional: true,
						Elem: &Schema{
							Type:             TypeString,
							ValidateDiagFunc: failValidation,
						},
					},
				},
			},
		},
	})

	c := terraform.NewResourceConfigRaw(map[string]interface{}{
		"block": []interface{}{
			map[string]interface{}{
				"inner": []interface{}{"a", "b"},
			},
		},
	})

	diags := m.Validate(c)
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d: %v", len(diags), diags)
	}

	got := map[string]bool{}
	for _, d := range diags {
		if len(d.AttributePath) == 0 {
			t.Fatalf("diagnostic has empty AttributePath: %v", d)
		}
		last, ok := d.AttributePath[len(d.AttributePath)-1].(cty.IndexStep)
		if !ok {
			t.Fatalf("expected last step to be IndexStep, got %T", d.AttributePath[len(d.AttributePath)-1])
		}
		got[last.Key.AsBigFloat().String()] = true
	}

	if !got["0"] || !got["1"] {
		t.Fatalf("expected one diagnostic path ending in index 0 and one in index 1, got %v", got)
	}
}
