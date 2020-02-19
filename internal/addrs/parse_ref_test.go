package addrs

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		Input   string
		Want    *reference
		WantErr string
	}{

		// count
		{
			`count.index`,
			&reference{
				Subject: countAttr{
					Name: "index",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
				},
			},
			``,
		},
		{
			`count.index.blah`,
			&reference{
				Subject: countAttr{
					Name: "index",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 12, Byte: 11},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
					},
				},
			},
			``, // valid at this layer, but will fail during eval because "index" is a number
		},
		{
			`count`,
			nil,
			`The "count" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`count["hello"]`,
			nil,
			`The "count" object does not support this operation.`,
		},

		// each
		{
			`each.key`,
			&reference{
				Subject: forEachAttr{
					Name: "key",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 9, Byte: 8},
				},
			},
			``,
		},
		{
			`each.value.blah`,
			&reference{
				Subject: forEachAttr{
					Name: "value",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 11, Byte: 10},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:   hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
					},
				},
			},
			``,
		},
		{
			`each`,
			nil,
			`The "each" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`each["hello"]`,
			nil,
			`The "each" object does not support this operation.`,
		},
		// data
		{
			`data.external.foo`,
			&reference{
				Subject: resource{
					Mode: DataResourceMode,
					Type: "external",
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 18, Byte: 17},
				},
			},
			``,
		},
		{
			`data.external.foo.bar`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: DataResourceMode,
						Type: "external",
						Name: "foo",
					},
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 18, Byte: 17},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "bar",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
							End:   hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
					},
				},
			},
			``,
		},
		{
			`data.external.foo["baz"].bar`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: DataResourceMode,
						Type: "external",
						Name: "foo",
					},
					Key: stringKey("baz"),
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 25, Byte: 24},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "bar",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 25, Byte: 24},
							End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
						},
					},
				},
			},
			``,
		},
		{
			`data.external.foo["baz"]`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: DataResourceMode,
						Type: "external",
						Name: "foo",
					},
					Key: stringKey("baz"),
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 25, Byte: 24},
				},
			},
			``,
		},
		{
			`data`,
			nil,
			`The "data" object must be followed by two attribute names: the data source type and the resource name.`,
		},
		{
			`data.external`,
			nil,
			`The "data" object must be followed by two attribute names: the data source type and the resource name.`,
		},

		// local
		{
			`local.foo`,
			&reference{
				Subject: localValue{
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 10, Byte: 9},
				},
			},
			``,
		},
		{
			`local.foo.blah`,
			&reference{
				Subject: localValue{
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 10, Byte: 9},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
			},
			``,
		},
		{
			`local.foo["blah"]`,
			&reference{
				Subject: localValue{
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 10, Byte: 9},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseIndex{
						Key: cty.StringVal("blah"),
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:   hcl.Pos{Line: 1, Column: 18, Byte: 17},
						},
					},
				},
			},
			``,
		},
		{
			`local`,
			nil,
			`The "local" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`local["foo"]`,
			nil,
			`The "local" object does not support this operation.`,
		},

		// module
		{
			`module.foo`,
			&reference{
				Subject: moduleCallInstance{
					Call: moduleCall{
						Name: "foo",
					},
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 11, Byte: 10},
				},
			},
			``,
		},
		{
			`module.foo.bar`,
			&reference{
				Subject: moduleCallOutput{
					Call: moduleCallInstance{
						Call: moduleCall{
							Name: "foo",
						},
					},
					Name: "bar",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 15, Byte: 14},
				},
			},
			``,
		},
		{
			`module.foo.bar.baz`,
			&reference{
				Subject: moduleCallOutput{
					Call: moduleCallInstance{
						Call: moduleCall{
							Name: "foo",
						},
					},
					Name: "bar",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 15, Byte: 14},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "baz",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:   hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
					},
				},
			},
			``,
		},
		{
			`module.foo["baz"]`,
			&reference{
				Subject: moduleCallInstance{
					Call: moduleCall{
						Name: "foo",
					},
					Key: stringKey("baz"),
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 18, Byte: 17},
				},
			},
			``,
		},
		{
			`module.foo["baz"].bar`,
			&reference{
				Subject: moduleCallOutput{
					Call: moduleCallInstance{
						Call: moduleCall{
							Name: "foo",
						},
						Key: stringKey("baz"),
					},
					Name: "bar",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 22, Byte: 21},
				},
			},
			``,
		},
		{
			`module.foo["baz"].bar.boop`,
			&reference{
				Subject: moduleCallOutput{
					Call: moduleCallInstance{
						Call: moduleCall{
							Name: "foo",
						},
						Key: stringKey("baz"),
					},
					Name: "bar",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 22, Byte: 21},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "boop",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:   hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			},
			``,
		},
		{
			`module`,
			nil,
			`The "module" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`module["foo"]`,
			nil,
			`The "module" object does not support this operation.`,
		},

		// path
		{
			`path.module`,
			&reference{
				Subject: pathAttr{
					Name: "module",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
				},
			},
			``,
		},
		{
			`path.module.blah`,
			&reference{
				Subject: pathAttr{
					Name: "module",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 12, Byte: 11},
							End:   hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
					},
				},
			},
			``, // valid at this layer, but will fail during eval because "module" is a string
		},
		{
			`path`,
			nil,
			`The "path" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`path["module"]`,
			nil,
			`The "path" object does not support this operation.`,
		},

		// self
		{
			`self`,
			&reference{
				Subject: self,
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 5, Byte: 4},
				},
			},
			``,
		},
		{
			`self.blah`,
			&reference{
				Subject: self,
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 5, Byte: 4},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			},
			``,
		},

		// terraform
		{
			`terraform.workspace`,
			&reference{
				Subject: terraformAttr{
					Name: "workspace",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 20, Byte: 19},
				},
			},
			``,
		},
		{
			`terraform.workspace.blah`,
			&reference{
				Subject: terraformAttr{
					Name: "workspace",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 20, Byte: 19},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 20, Byte: 19},
							End:   hcl.Pos{Line: 1, Column: 25, Byte: 24},
						},
					},
				},
			},
			``, // valid at this layer, but will fail during eval because "workspace" is a string
		},
		{
			`terraform`,
			nil,
			`The "terraform" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`terraform["workspace"]`,
			nil,
			`The "terraform" object does not support this operation.`,
		},

		// var
		{
			`var.foo`,
			&reference{
				Subject: inputVariable{
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 8, Byte: 7},
				},
			},
			``,
		},
		{
			`var.foo.blah`,
			&reference{
				Subject: inputVariable{
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 8, Byte: 7},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "blah",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
			},
			``, // valid at this layer, but will fail during eval because "module" is a string
		},
		{
			`var`,
			nil,
			`The "var" object cannot be accessed directly. Instead, access one of its attributes.`,
		},
		{
			`var["foo"]`,
			nil,
			`The "var" object does not support this operation.`,
		},

		// anything else, interpreted as a managed resource reference
		{
			`boop_instance.foo`,
			&reference{
				Subject: resource{
					Mode: ManagedResourceMode,
					Type: "boop_instance",
					Name: "foo",
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 18, Byte: 17},
				},
			},
			``,
		},
		{
			`boop_instance.foo.bar`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: ManagedResourceMode,
						Type: "boop_instance",
						Name: "foo",
					},
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 18, Byte: 17},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "bar",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
							End:   hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
					},
				},
			},
			``,
		},
		{
			`boop_instance.foo["baz"].bar`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: ManagedResourceMode,
						Type: "boop_instance",
						Name: "foo",
					},
					Key: stringKey("baz"),
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 25, Byte: 24},
				},
				Remaining: hcl.Traversal{
					hcl.TraverseAttr{
						Name: "bar",
						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 25, Byte: 24},
							End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
						},
					},
				},
			},
			``,
		},
		{
			`boop_instance.foo["baz"]`,
			&reference{
				Subject: resourceInstance{
					Resource: resource{
						Mode: ManagedResourceMode,
						Type: "boop_instance",
						Name: "foo",
					},
					Key: stringKey("baz"),
				},
				SourceRange: tfdiags.SourceRange{
					Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
					End:   tfdiags.SourcePos{Line: 1, Column: 25, Byte: 24},
				},
			},
			``,
		},
		{
			`boop_instance`,
			nil,
			`A reference to a resource type must be followed by at least one attribute access, specifying the resource name.`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			traversal, travDiags := hclsyntax.ParseTraversalAbs([]byte(test.Input), "", hcl.Pos{Line: 1, Column: 1})
			if travDiags.HasErrors() {
				t.Fatal(travDiags.Error())
			}

			got, diags := parseRefPub(traversal)

			switch len(diags) {
			case 0:
				if test.WantErr != "" {
					t.Fatalf("succeeded; want error: %s", test.WantErr)
				}
			case 1:
				if test.WantErr == "" {
					t.Fatalf("unexpected diagnostics: %s", diags.Err())
				}
				if got, want := diags[0].Description().Detail, test.WantErr; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
			default:
				t.Fatalf("too many diagnostics: %s", diags.Err())
			}

			if diags.HasErrors() {
				return
			}

			for _, problem := range deep.Equal(got, test.Want) {
				t.Errorf(problem)
			}
		})
	}
}
