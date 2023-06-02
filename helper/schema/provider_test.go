// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const invalidDurationErrMsg = "time: invalid duration \"invalid\""

func TestProviderGetSchema(t *testing.T) {
	// This functionality is already broadly tested in core_schema_test.go,
	// so this is just to ensure that the call passes through correctly.
	p := &Provider{
		Schema: map[string]*Schema{
			"bar": {
				Type:     TypeString,
				Required: true,
			},
		},
		ResourcesMap: map[string]*Resource{
			"foo": {
				Schema: map[string]*Schema{
					"bar": {
						Type:     TypeString,
						Required: true,
					},
				},
			},
		},
		DataSourcesMap: map[string]*Resource{
			"baz": {
				Schema: map[string]*Schema{
					"bur": {
						Type:     TypeString,
						Required: true,
					},
				},
			},
		},
	}

	want := &terraform.ProviderSchema{
		Provider: &configschema.Block{
			Attributes: map[string]*configschema.Attribute{
				"bar": {
					Type:     cty.String,
					Required: true,
				},
			},
			BlockTypes: map[string]*configschema.NestedBlock{},
		},
		ResourceTypes: map[string]*configschema.Block{
			"foo": testResource(&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"bar": {
						Type:     cty.String,
						Required: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{},
			}),
		},
		DataSources: map[string]*configschema.Block{
			"baz": testResource(&configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"bur": {
						Type:     cty.String,
						Required: true,
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{},
			}),
		},
	}
	got, err := p.GetSchema(&terraform.ProviderSchemaRequest{
		ResourceTypes: []string{"foo", "bar"},
		DataSources:   []string{"baz", "bar"},
	})
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if !cmp.Equal(got, want, equateEmpty, typeComparer) {
		t.Error("wrong result:\n", cmp.Diff(got, want, equateEmpty, typeComparer))
	}
}

func TestProviderConfigure(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		P             *Provider
		Config        map[string]interface{}
		ExpectedDiags diag.Diagnostics
	}{
		"nil": {
			P:      &Provider{},
			Config: nil,
		},

		"ConfigureFunc-no-diags": {
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureFunc: func(d *ResourceData) (interface{}, error) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, fmt.Errorf("nope")
				},
			},
			Config: map[string]interface{}{
				"foo": 42,
			},
		},

		"ConfigureContextFunc-no-diags": {
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureContextFunc: func(ctx context.Context, d *ResourceData) (interface{}, diag.Diagnostics) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, diag.Errorf("nope")
				},
			},
			Config: map[string]interface{}{
				"foo": 42,
			},
		},

		"ConfigureFunc-error": {
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureFunc: func(d *ResourceData) (interface{}, error) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, fmt.Errorf("nope")
				},
			},
			Config: map[string]interface{}{
				"foo": 52,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "nope",
					Detail:   "",
				},
			},
		},

		"ConfigureContextFunc-error": {
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureContextFunc: func(ctx context.Context, d *ResourceData) (interface{}, diag.Diagnostics) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, diag.Diagnostics{
						{
							Severity: diag.Error,
							Summary:  "Test Error Diagnostic",
							Detail:   "This is an error.",
						},
					}
				},
			},
			Config: map[string]interface{}{
				"foo": 52,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Test Error Diagnostic",
					Detail:   "This is an error.",
				},
			},
		},

		"ConfigureContextFunc-warning": {
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureContextFunc: func(ctx context.Context, d *ResourceData) (interface{}, diag.Diagnostics) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, diag.Diagnostics{
						{
							Severity: diag.Warning,
							Summary:  "Test Warning Diagnostic",
							Detail:   "This is a warning.",
						},
					}
				},
			},
			Config: map[string]interface{}{
				"foo": 52,
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Test Warning Diagnostic",
					Detail:   "This is a warning.",
				},
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := terraform.NewResourceConfigRaw(tc.Config)
			diags := tc.P.Configure(context.Background(), c)

			if diff := cmp.Diff(tc.ExpectedDiags, diags); diff != "" {
				t.Errorf("Unexpected diagnostics (-wanted +got): %s", diff)
			}
		})
	}
}

func TestProviderResources(t *testing.T) {
	cases := []struct {
		P      *Provider
		Result []terraform.ResourceType
	}{
		{
			P:      &Provider{},
			Result: []terraform.ResourceType{},
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": nil,
				},
			},
			Result: []terraform.ResourceType{
				{Name: "bar", SchemaAvailable: true},
				{Name: "foo", SchemaAvailable: true},
			},
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": {Importer: &ResourceImporter{}},
					"baz": nil,
				},
			},
			Result: []terraform.ResourceType{
				{Name: "bar", Importable: true, SchemaAvailable: true},
				{Name: "baz", SchemaAvailable: true},
				{Name: "foo", SchemaAvailable: true},
			},
		},
	}

	for i, tc := range cases {
		actual := tc.P.Resources()
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d: %#v", i, actual)
		}
	}
}

func TestProviderDataSources(t *testing.T) {
	cases := []struct {
		P      *Provider
		Result []terraform.DataSource
	}{
		{
			P:      &Provider{},
			Result: []terraform.DataSource{},
		},

		{
			P: &Provider{
				DataSourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": nil,
				},
			},
			Result: []terraform.DataSource{
				{Name: "bar", SchemaAvailable: true},
				{Name: "foo", SchemaAvailable: true},
			},
		},
	}

	for i, tc := range cases {
		actual := tc.P.DataSources()
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d: got %#v; want %#v", i, actual, tc.Result)
		}
	}
}

func TestProviderValidate(t *testing.T) {
	cases := []struct {
		P      *Provider
		Config map[string]interface{}
		Err    bool
	}{
		{
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {},
				},
			},
			Config: nil,
			Err:    true,
		},
	}

	for i, tc := range cases {
		c := terraform.NewResourceConfigRaw(tc.Config)
		diags := tc.P.Validate(c)
		if diags.HasError() != tc.Err {
			t.Fatalf("%d: %#v", i, diags)
		}
	}
}

func TestProviderValidate_attributePath(t *testing.T) {
	cases := []struct {
		P             *Provider
		Config        map[string]interface{}
		ExpectedDiags diag.Diagnostics
	}{
		{ // legacy validate path automatically built, even across list
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeList,
						Required: true,
						Elem: &Resource{
							Schema: map[string]*Schema{
								"bar": {
									Type:     TypeString,
									Required: true,
									ValidateFunc: func(v interface{}, k string) ([]string, []error) {
										return []string{"warn"}, []error{fmt.Errorf("error")}
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Warning,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}, cty.IndexStep{Key: cty.NumberIntVal(0)}, cty.GetAttrStep{Name: "bar"}},
				},
				{
					Severity:      diag.Error,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}, cty.IndexStep{Key: cty.NumberIntVal(0)}, cty.GetAttrStep{Name: "bar"}},
				},
			},
		},
		{ // validate path automatically built, even across list
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeList,
						Required: true,
						Elem: &Resource{
							Schema: map[string]*Schema{
								"bar": {
									Type:     TypeString,
									Required: true,
									ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
										return diag.Diagnostics{{Severity: diag.Error}}
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}, cty.IndexStep{Key: cty.NumberIntVal(0)}, cty.GetAttrStep{Name: "bar"}},
				},
			},
		},
		{ // path is truncated at typeset
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeSet,
						Required: true,
						Elem: &Resource{
							Schema: map[string]*Schema{
								"bar": {
									Type:     TypeString,
									Required: true,
									ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
										return diag.Diagnostics{{Severity: diag.Error, AttributePath: cty.Path{cty.GetAttrStep{Name: "doesnotmatter"}}}}
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar": "baz",
					},
				},
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}},
				},
			},
		},
		{ // relative path is appended
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeList,
						Required: true,
						Elem: &Resource{
							Schema: map[string]*Schema{
								"bar": {
									Type:     TypeMap,
									Required: true,
									ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
										return diag.Diagnostics{{Severity: diag.Error, AttributePath: cty.Path{cty.IndexStep{Key: cty.StringVal("mapkey")}}}}
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar": map[string]interface{}{
							"mapkey": "val",
						},
					},
				},
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}, cty.IndexStep{Key: cty.NumberIntVal(0)}, cty.GetAttrStep{Name: "bar"}, cty.IndexStep{Key: cty.StringVal("mapkey")}},
				},
			},
		},
		{ // absolute path is not altered
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeList,
						Required: true,
						Elem: &Resource{
							Schema: map[string]*Schema{
								"bar": {
									Type:     TypeMap,
									Required: true,
									ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
										return diag.Diagnostics{{Severity: diag.Error, AttributePath: append(path, cty.IndexStep{Key: cty.StringVal("mapkey")})}}
									},
								},
							},
						},
					},
				},
			},
			Config: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar": map[string]interface{}{
							"mapkey": "val",
						},
					},
				},
			},
			ExpectedDiags: diag.Diagnostics{
				{
					Severity:      diag.Error,
					AttributePath: cty.Path{cty.GetAttrStep{Name: "foo"}, cty.IndexStep{Key: cty.NumberIntVal(0)}, cty.GetAttrStep{Name: "bar"}, cty.IndexStep{Key: cty.StringVal("mapkey")}},
				},
			},
		},
	}

	for i, tc := range cases {
		c := terraform.NewResourceConfigRaw(tc.Config)
		diags := tc.P.Validate(c)
		if len(diags) != len(tc.ExpectedDiags) {
			t.Fatalf("%d: wrong number of diags, expected %d, got %d", i, len(tc.ExpectedDiags), len(diags))
		}
		for j := range diags {
			if diags[j].Severity != tc.ExpectedDiags[j].Severity {
				t.Fatalf("%d: expected severity %v, got %v", i, tc.ExpectedDiags[j].Severity, diags[j].Severity)
			}
			if !diags[j].AttributePath.Equals(tc.ExpectedDiags[j].AttributePath) {
				t.Fatalf("%d: attribute paths do not match expected: %v, got %v", i, tc.ExpectedDiags[j].AttributePath, diags[j].AttributePath)
			}
		}
	}
}

func TestProviderDiff_legacyTimeoutType(t *testing.T) {
	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"blah": {
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				Timeouts: &ResourceTimeout{
					Create: DefaultTimeout(10 * time.Minute),
				},
			},
		},
	}

	invalidCfg := map[string]interface{}{
		"foo": 42,
		"timeouts": []interface{}{
			map[string]interface{}{
				"create": "40m",
			},
		},
	}
	ic := terraform.NewResourceConfigRaw(invalidCfg)
	_, err := p.ResourcesMap["blah"].Diff(
		context.Background(),
		nil,
		ic,
		p.Meta(),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderDiff_timeoutInvalidValue(t *testing.T) {
	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"blah": {
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeInt,
						Optional: true,
					},
				},
				Timeouts: &ResourceTimeout{
					Create: DefaultTimeout(10 * time.Minute),
				},
			},
		},
	}

	invalidCfg := map[string]interface{}{
		"foo": 42,
		"timeouts": map[string]interface{}{
			"create": "invalid",
		},
	}
	ic := terraform.NewResourceConfigRaw(invalidCfg)
	_, err := p.ResourcesMap["blah"].Diff(
		context.Background(),
		nil,
		ic,
		p.Meta(),
	)
	if err == nil {
		t.Fatal("Expected provider.Diff to fail with invalid timeout value")
	}
	if !strings.Contains(err.Error(), invalidDurationErrMsg) {
		t.Fatalf("Unexpected error message: %q\nExpected message to contain %q",
			err.Error(),
			invalidDurationErrMsg)
	}
}

func TestProviderValidateResource(t *testing.T) {
	cases := []struct {
		P      *Provider
		Type   string
		Config map[string]interface{}
		Err    bool
	}{
		{
			P:      &Provider{},
			Type:   "foo",
			Config: nil,
			Err:    true,
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": {},
				},
			},
			Type:   "foo",
			Config: nil,
			Err:    false,
		},
	}

	for i, tc := range cases {
		c := terraform.NewResourceConfigRaw(tc.Config)
		diags := tc.P.ValidateResource(tc.Type, c)
		if diags.HasError() != tc.Err {
			t.Fatalf("%d: %#v", i, diags)
		}
	}
}

func TestProviderImportState(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		provider       *Provider
		info           *terraform.InstanceInfo
		id             string
		expectedStates []*terraform.InstanceState
		expectedErr    error
	}{
		"error-unknown-resource-type": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id:          "test-id",
			expectedErr: fmt.Errorf("unknown resource type: test_resource"),
		},
		"error-no-Importer": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": { /* no Importer */ },
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id:          "test-id",
			expectedErr: fmt.Errorf("resource test_resource doesn't support import"),
		},
		"error-missing-ResourceData": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": {
						Importer: &ResourceImporter{
							StateContext: func(_ context.Context, _ *ResourceData, _ interface{}) ([]*ResourceData, error) {
								return []*ResourceData{nil}, nil
							},
						},
					},
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id:          "test-id",
			expectedErr: fmt.Errorf("The provider returned a missing resource during ImportResourceState."),
		},
		"error-missing-ResourceData-Id": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": {
						Importer: &ResourceImporter{
							StateContext: func(_ context.Context, d *ResourceData, _ interface{}) ([]*ResourceData, error) {
								// Example from calling Read functionality,
								// but not checking for missing resource before return
								d.SetId("")
								return []*ResourceData{d}, nil
							},
						},
					},
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id:          "test-id",
			expectedErr: fmt.Errorf("The provider returned a resource missing an identifier during ImportResourceState."),
		},
		"Importer": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": {
						Importer: &ResourceImporter{},
					},
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id: "test-id",
			expectedStates: []*terraform.InstanceState{
				{
					Attributes: map[string]string{"id": "test-id"},
					Ephemeral:  terraform.EphemeralState{Type: "test_resource"},
					ID:         "test-id",
					Meta:       map[string]interface{}{"schema_version": "0"},
				},
			},
		},
		"Importer-State": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": {
						Importer: &ResourceImporter{
							State: func(d *ResourceData, _ interface{}) ([]*ResourceData, error) {
								if d.Id() != "test-id" {
									return nil, fmt.Errorf("expected d.Id() %q, got: %s", "test-id", d.Id())
								}

								if d.State().Ephemeral.Type != "test_resource" {
									return nil, fmt.Errorf("expected d.State().Ephemeral.Type %q, got: %s", "test_resource", d.State().Ephemeral.Type)
								}

								return []*ResourceData{d}, nil
							},
						},
					},
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id: "test-id",
			expectedStates: []*terraform.InstanceState{
				{
					Attributes: map[string]string{"id": "test-id"},
					Ephemeral:  terraform.EphemeralState{Type: "test_resource"},
					ID:         "test-id",
					Meta:       map[string]interface{}{"schema_version": "0"},
				},
			},
		},
		"Importer-StateContext": {
			provider: &Provider{
				ResourcesMap: map[string]*Resource{
					"test_resource": {
						Importer: &ResourceImporter{
							StateContext: func(_ context.Context, d *ResourceData, meta interface{}) ([]*ResourceData, error) {
								if d.Id() != "test-id" {
									return nil, fmt.Errorf("expected d.Id() %q, got: %s", "test-id", d.Id())
								}

								if d.State().Ephemeral.Type != "test_resource" {
									return nil, fmt.Errorf("expected d.State().Ephemeral.Type %q, got: %s", "test_resource", d.State().Ephemeral.Type)
								}

								return []*ResourceData{d}, nil
							},
						},
					},
				},
			},
			info: &terraform.InstanceInfo{
				Type: "test_resource",
			},
			id: "test-id",
			expectedStates: []*terraform.InstanceState{
				{
					Attributes: map[string]string{"id": "test-id"},
					Ephemeral:  terraform.EphemeralState{Type: "test_resource"},
					ID:         "test-id",
					Meta:       map[string]interface{}{"schema_version": "0"},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			states, err := testCase.provider.ImportState(context.Background(), testCase.info, testCase.id)

			if err != nil {
				if testCase.expectedErr == nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedErr.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedErr, err)
				}
			}

			if err == nil && testCase.expectedErr != nil {
				t.Fatalf("expected error %q, got none", testCase.expectedErr)
			}

			if diff := cmp.Diff(states, testCase.expectedStates); diff != "" {
				t.Fatalf("unexpected states difference: %s", diff)
			}
		})
	}
}

func TestProviderMeta(t *testing.T) {
	p := new(Provider)
	if v := p.Meta(); v != nil {
		t.Fatalf("bad: %#v", v)
	}

	expected := 42
	p.SetMeta(42)
	if v := p.Meta(); !reflect.DeepEqual(v, expected) {
		t.Fatalf("bad: %#v", v)
	}
}

func TestProvider_InternalValidate(t *testing.T) {
	cases := []struct {
		P           *Provider
		ExpectedErr error
	}{
		{
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeBool,
						Optional: true,
					},
				},
			},
			ExpectedErr: nil,
		},
		{ // Reserved resource fields should be allowed in provider block
			P: &Provider{
				Schema: map[string]*Schema{
					"provisioner": {
						Type:     TypeString,
						Optional: true,
					},
					"count": {
						Type:     TypeInt,
						Optional: true,
					},
				},
			},
			ExpectedErr: nil,
		},
		{ // Reserved provider fields should not be allowed
			P: &Provider{
				Schema: map[string]*Schema{
					"alias": {
						Type:     TypeString,
						Optional: true,
					},
				},
			},
			ExpectedErr: fmt.Errorf("%s is a reserved field name for a provider", "alias"),
		},
		{ // ConfigureFunc and ConfigureContext cannot both be set
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": {
						Type:     TypeString,
						Optional: true,
					},
				},
				ConfigureFunc: func(d *ResourceData) (interface{}, error) {
					return nil, nil
				},
				ConfigureContextFunc: func(ctx context.Context, d *ResourceData) (interface{}, diag.Diagnostics) {
					return nil, nil
				},
			},
			ExpectedErr: fmt.Errorf("ConfigureFunc and ConfigureContextFunc must not both be set"),
		},
	}

	for i, tc := range cases {
		err := tc.P.InternalValidate()
		if tc.ExpectedErr == nil {
			if err != nil {
				t.Fatalf("%d: Error returned (expected no error): %s", i, err)
			}
			continue
		}
		if tc.ExpectedErr != nil && err == nil {
			t.Fatalf("%d: Expected error (%s), but no error returned", i, tc.ExpectedErr)
		}
		if err.Error() != tc.ExpectedErr.Error() {
			t.Fatalf("%d: Errors don't match. Expected: %#v Given: %#v", i, tc.ExpectedErr, err)
		}
	}
}

func TestProviderUserAgentAppendViaEnvVar(t *testing.T) {
	if oldenv, isSet := os.LookupEnv(uaEnvVar); isSet {
		defer os.Setenv(uaEnvVar, oldenv)
	} else {
		defer os.Unsetenv(uaEnvVar)
	}

	expectedBase := "Terraform/4.5.6 (+https://www.terraform.io) Terraform-Plugin-SDK/" + meta.SDKVersionString()

	testCases := []struct {
		providerName    string
		providerVersion string
		envVarValue     string
		expected        string
	}{
		{"", "", "", expectedBase},
		{"", "", " ", expectedBase},
		{"", "", " \n", expectedBase},
		{"", "", "test/1", expectedBase + " test/1"},
		{"", "", "test/1 (comment)", expectedBase + " test/1 (comment)"},
		{"My-Provider", "", "", expectedBase + " My-Provider"},
		{"My-Provider", "", " ", expectedBase + " My-Provider"},
		{"My-Provider", "", " \n", expectedBase + " My-Provider"},
		{"My-Provider", "", "test/1", expectedBase + " My-Provider test/1"},
		{"My-Provider", "", "test/1 (comment)", expectedBase + " My-Provider test/1 (comment)"},
		{"My-Provider", "1.2.3", "", expectedBase + " My-Provider/1.2.3"},
		{"My-Provider", "1.2.3", " ", expectedBase + " My-Provider/1.2.3"},
		{"My-Provider", "1.2.3", " \n", expectedBase + " My-Provider/1.2.3"},
		{"My-Provider", "1.2.3", "test/1", expectedBase + " My-Provider/1.2.3 test/1"},
		{"My-Provider", "1.2.3", "test/1 (comment)", expectedBase + " My-Provider/1.2.3 test/1 (comment)"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Setenv(uaEnvVar, tc.envVarValue)
			p := &Provider{TerraformVersion: "4.5.6"}
			givenUA := p.UserAgent(tc.providerName, tc.providerVersion)
			if givenUA != tc.expected {
				t.Fatalf("Expected User-Agent '%s' does not match '%s'", tc.expected, givenUA)
			}
		})
	}
}
