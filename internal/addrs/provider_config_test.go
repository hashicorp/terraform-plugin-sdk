package addrs

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestParseProviderConfigCompact(t *testing.T) {
	tests := []struct {
		Input    string
		Want     ProviderConfig
		WantDiag string
	}{
		{
			`aws`,
			ProviderConfig{
				Type: "aws",
			},
			``,
		},
		{
			`aws.foo`,
			ProviderConfig{
				Type:  "aws",
				Alias: "foo",
			},
			``,
		},
		{
			`aws["foo"]`,
			ProviderConfig{},
			`The provider type name must either stand alone or be followed by an alias name separated with a dot.`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			traversal, parseDiags := hclsyntax.ParseTraversalAbs([]byte(test.Input), "", hcl.Pos{})
			if len(parseDiags) != 0 {
				t.Errorf("unexpected diagnostics during parse")
				for _, diag := range parseDiags {
					t.Logf("- %s", diag)
				}
				return
			}

			got, diags := parseProviderConfigCompact(traversal)

			if test.WantDiag != "" {
				if len(diags) != 1 {
					t.Fatalf("got %d diagnostics; want 1", len(diags))
				}
				gotDetail := diags[0].Description().Detail
				if gotDetail != test.WantDiag {
					t.Fatalf("wrong diagnostic detail\ngot:  %s\nwant: %s", gotDetail, test.WantDiag)
				}
				return
			} else {
				if len(diags) != 0 {
					t.Fatalf("got %d diagnostics; want 0", len(diags))
				}
			}

			for _, problem := range deep.Equal(got, test.Want) {
				t.Error(problem)
			}
		})
	}
}
func TestParseAbsProviderConfig(t *testing.T) {
	tests := []struct {
		Input    string
		Want     absProviderConfig
		WantDiag string
	}{
		{
			`provider.aws`,
			absProviderConfig{
				Module: RootModuleInstance,
				ProviderConfig: ProviderConfig{
					Type: "aws",
				},
			},
			``,
		},
		{
			`provider.aws.foo`,
			absProviderConfig{
				Module: RootModuleInstance,
				ProviderConfig: ProviderConfig{
					Type:  "aws",
					Alias: "foo",
				},
			},
			``,
		},
		{
			`module.baz.provider.aws`,
			absProviderConfig{
				Module: ModuleInstance{
					{
						Name: "baz",
					},
				},
				ProviderConfig: ProviderConfig{
					Type: "aws",
				},
			},
			``,
		},
		{
			`module.baz.provider.aws.foo`,
			absProviderConfig{
				Module: ModuleInstance{
					{
						Name: "baz",
					},
				},
				ProviderConfig: ProviderConfig{
					Type:  "aws",
					Alias: "foo",
				},
			},
			``,
		},
		{
			`module.baz["foo"].provider.aws`,
			absProviderConfig{
				Module: ModuleInstance{
					{
						Name:        "baz",
						InstanceKey: stringKey("foo"),
					},
				},
				ProviderConfig: ProviderConfig{
					Type: "aws",
				},
			},
			``,
		},
		{
			`module.baz[1].provider.aws`,
			absProviderConfig{
				Module: ModuleInstance{
					{
						Name:        "baz",
						InstanceKey: intKey(1),
					},
				},
				ProviderConfig: ProviderConfig{
					Type: "aws",
				},
			},
			``,
		},
		{
			`module.baz[1].module.bar.provider.aws`,
			absProviderConfig{
				Module: ModuleInstance{
					{
						Name:        "baz",
						InstanceKey: intKey(1),
					},
					{
						Name: "bar",
					},
				},
				ProviderConfig: ProviderConfig{
					Type: "aws",
				},
			},
			``,
		},
		{
			`aws`,
			absProviderConfig{},
			`Provider address must begin with "provider.", followed by a provider type name.`,
		},
		{
			`aws.foo`,
			absProviderConfig{},
			`Provider address must begin with "provider.", followed by a provider type name.`,
		},
		{
			`provider`,
			absProviderConfig{},
			`Provider address must begin with "provider.", followed by a provider type name.`,
		},
		{
			`provider.aws.foo.bar`,
			absProviderConfig{},
			`Extraneous operators after provider configuration alias.`,
		},
		{
			`provider["aws"]`,
			absProviderConfig{},
			`The prefix "provider." must be followed by a provider type name.`,
		},
		{
			`provider.aws["foo"]`,
			absProviderConfig{},
			`Provider type name must be followed by a configuration alias name.`,
		},
		{
			`module.foo`,
			absProviderConfig{},
			`Provider address must begin with "provider.", followed by a provider type name.`,
		},
		{
			`module.foo["provider"]`,
			absProviderConfig{},
			`Provider address must begin with "provider.", followed by a provider type name.`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			traversal, parseDiags := hclsyntax.ParseTraversalAbs([]byte(test.Input), "", hcl.Pos{})
			if len(parseDiags) != 0 {
				t.Errorf("unexpected diagnostics during parse")
				for _, diag := range parseDiags {
					t.Logf("- %s", diag)
				}
				return
			}

			got, diags := parseAbsProviderConfig(traversal)

			if test.WantDiag != "" {
				if len(diags) != 1 {
					t.Fatalf("got %d diagnostics; want 1", len(diags))
				}
				gotDetail := diags[0].Description().Detail
				if gotDetail != test.WantDiag {
					t.Fatalf("wrong diagnostic detail\ngot:  %s\nwant: %s", gotDetail, test.WantDiag)
				}
				return
			} else {
				if len(diags) != 0 {
					t.Fatalf("got %d diagnostics; want 0", len(diags))
				}
			}

			for _, problem := range deep.Equal(got, test.Want) {
				t.Error(problem)
			}
		})
	}
}
