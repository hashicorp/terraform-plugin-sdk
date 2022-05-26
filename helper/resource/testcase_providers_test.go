package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestTestCaseProviderConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testCase TestCase
		expected string
	}{
		"externalproviders-missing-source-and-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {},
				},
			},
			expected: `provider "test" {}`,
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

			got := test.testCase.providerConfig(context.Background())

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
