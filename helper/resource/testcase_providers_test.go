// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestTestCaseProviderConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testCase          TestCase
		skipProviderBlock bool
		expected          string
	}{
		"externalproviders-and-protov5providerfactories": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"localtest": nil,
				},
			},
			expected: `
terraform {
  required_providers {
    externaltest = {
      source = "registry.terraform.io/hashicorp/externaltest"
      version = "1.2.3"
    }
  }
}

provider "externaltest" {}
`,
		},
		"externalproviders-and-protov6providerfactories": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"localtest": nil,
				},
			},
			expected: `
terraform {
  required_providers {
    externaltest = {
      source = "registry.terraform.io/hashicorp/externaltest"
      version = "1.2.3"
    }
  }
}

provider "externaltest" {}
`,
		},
		"externalproviders-and-providerfactories": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"localtest": nil,
				},
			},
			expected: `
terraform {
  required_providers {
    externaltest = {
      source = "registry.terraform.io/hashicorp/externaltest"
      version = "1.2.3"
    }
  }
}

provider "externaltest" {}
`,
		},
		"externalproviders-missing-source-and-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {},
				},
			},
			expected: `provider "test" {}`,
		},
		"externalproviders-skip-provider-block": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source:            "registry.terraform.io/hashicorp/test",
						VersionConstraint: "1.2.3",
					},
				},
			},
			skipProviderBlock: true,
			expected: `
terraform {
  required_providers {
    test = {
      source = "registry.terraform.io/hashicorp/test"
      version = "1.2.3"
    }
  }
}
`,
		},
		"externalproviders-source-and-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source:            "registry.terraform.io/hashicorp/test",
						VersionConstraint: "1.2.3",
					},
				},
			},
			expected: `
terraform {
  required_providers {
    test = {
      source = "registry.terraform.io/hashicorp/test"
      version = "1.2.3"
    }
  }
}

provider "test" {}
`,
		},
		"externalproviders-source": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source: "registry.terraform.io/hashicorp/test",
					},
				},
			},
			expected: `
terraform {
  required_providers {
    test = {
      source = "registry.terraform.io/hashicorp/test"
    }
  }
}

provider "test" {}
`,
		},
		"externalproviders-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						VersionConstraint: "1.2.3",
					},
				},
			},
			expected: `
terraform {
  required_providers {
    test = {
      version = "1.2.3"
    }
  }
}

provider "test" {}
`,
		},
		"protov5providerfactories": {
			testCase: TestCase{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil,
				},
			},
			expected: ``,
		},
		"protov6providerfactories": {
			testCase: TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil,
				},
			},
			expected: ``,
		},
		"providerfactories": {
			testCase: TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil,
				},
			},
			expected: ``,
		},
		"providers": {
			testCase: TestCase{
				Providers: map[string]*schema.Provider{
					"test": {},
				},
			},
			expected: `provider "test" {}`,
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.testCase.providerConfig(context.Background(), test.skipProviderBlock)

			if diff := cmp.Diff(strings.TrimSpace(got), strings.TrimSpace(test.expected)); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTest_TestCase_ExternalProviders(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"null": {
				Source: "registry.terraform.io/hashicorp/null",
			},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})
}

func TestTest_TestCase_ExternalProviders_NonHashiCorpNamespace(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			// This can be set to any provider outside the hashicorp namespace.
			// bflad/scaffoldingtest happens to be a published version of
			// terraform-provider-scaffolding-framework.
			"scaffoldingtest": {
				Source:            "registry.terraform.io/bflad/scaffoldingtest",
				VersionConstraint: "0.1.0",
			},
		},
		Steps: []TestStep{
			{
				Config: `resource "scaffoldingtest_example" "test" {}`,
			},
		},
	})
}

func TestTest_TestCase_ExternalProvidersAndProviderFactories_NonHashiCorpNamespace(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			// This can be set to any provider outside the hashicorp namespace.
			// bflad/scaffoldingtest happens to be a published version of
			// terraform-provider-scaffolding-framework.
			"scaffoldingtest": {
				Source:            "registry.terraform.io/bflad/scaffoldingtest",
				VersionConstraint: "0.1.0",
			},
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"null": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"null_resource": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("test")
								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							Schema: map[string]*schema.Schema{
								"triggers": {
									Elem:     &schema.Schema{Type: schema.TypeString},
									ForceNew: true,
									Optional: true,
									Type:     schema.TypeMap,
								},
							},
						},
					},
				}, nil
			},
		},
		Steps: []TestStep{
			{
				Config: `
					resource "null_resource" "test" {}
					resource "scaffoldingtest_example" "test" {}
				`,
			},
		},
	})
}

func TestTest_TestCase_ExternalProviders_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			ExternalProviders: map[string]ExternalProvider{
				"testnonexistent": {
					Source: "registry.terraform.io/hashicorp/testnonexistent",
				},
			},
			Steps: []TestStep{
				{
					Config: "# not empty",
				},
			},
		})
	})
}

func TestTest_TestCase_ProtoV5ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"test": func() (tfprotov5.ProviderServer, error) { //nolint:unparam // required signature
				return nil, nil
			},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})
}

func TestTest_TestCase_ProtoV5ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
				"test": func() (tfprotov5.ProviderServer, error) { //nolint:unparam // required signature
					return nil, fmt.Errorf("test")
				},
			},
			Steps: []TestStep{
				{
					Config: "# not empty",
				},
			},
		})
	})
}

func TestTest_TestCase_ProtoV6ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": func() (tfprotov6.ProviderServer, error) { //nolint:unparam // required signature
				return nil, nil
			},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})
}

func TestTest_TestCase_ProtoV6ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"test": func() (tfprotov6.ProviderServer, error) { //nolint:unparam // required signature
					return nil, fmt.Errorf("test")
				},
			},
			Steps: []TestStep{
				{
					Config: "# not empty",
				},
			},
		})
	})
}

func TestTest_TestCase_ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return nil, nil
			},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})
}

func TestTest_TestCase_ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
					return nil, fmt.Errorf("test")
				},
			},
			Steps: []TestStep{
				{
					Config: "# not empty",
				},
			},
		})
	})
}

func TestTest_TestCase_Providers(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		Providers: map[string]*schema.Provider{
			"test": {},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})
}
