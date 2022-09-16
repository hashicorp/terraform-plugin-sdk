package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestStepMergedConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		testCase TestCase
		testStep TestStep
		expected string
	}{
		"testcase-externalproviders-and-protov5providerfactories": {
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
			testStep: TestStep{
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"testcase-externalproviders-and-protov6providerfactories": {
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
			testStep: TestStep{
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"testcase-externalproviders-and-providerfactories": {
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
			testStep: TestStep{
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"testcase-externalproviders-missing-source-and-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {},
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
provider "test" {}

resource "test_test" "test" {}
`,
		},
		"testcase-externalproviders-source-and-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source:            "registry.terraform.io/hashicorp/test",
						VersionConstraint: "1.2.3",
					},
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"testcase-externalproviders-source": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source: "registry.terraform.io/hashicorp/test",
					},
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"testcase-externalproviders-versionconstraint": {
			testCase: TestCase{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						VersionConstraint: "1.2.3",
					},
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"testcase-protov5providerfactories": {
			testCase: TestCase{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil,
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
		"testcase-protov6providerfactories": {
			testCase: TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil,
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
		"testcase-providerfactories": {
			testCase: TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil,
				},
			},
			testStep: TestStep{
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
		"teststep-externalproviders-and-protov5providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"localtest": nil,
				},
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"teststep-externalproviders-and-protov6providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"localtest": nil,
				},
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"teststep-externalproviders-and-providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"externaltest": {
						Source:            "registry.terraform.io/hashicorp/externaltest",
						VersionConstraint: "1.2.3",
					},
				},
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"localtest": nil,
				},
				Config: `
resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
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


resource "externaltest_test" "test" {}

resource "localtest_test" "test" {}
`,
		},
		"teststep-externalproviders-missing-source-and-versionconstraint": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {},
				},
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
provider "test" {}

resource "test_test" "test" {}
`,
		},
		"teststep-externalproviders-source-and-versionconstraint": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source:            "registry.terraform.io/hashicorp/test",
						VersionConstraint: "1.2.3",
					},
				},
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"teststep-externalproviders-source": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						Source: "registry.terraform.io/hashicorp/test",
					},
				},
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"teststep-externalproviders-versionconstraint": {
			testCase: TestCase{},
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {
						VersionConstraint: "1.2.3",
					},
				},
				Config: `
resource "test_test" "test" {}
`,
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


resource "test_test" "test" {}
`,
		},
		"teststep-protov5providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil,
				},
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
		"teststep-protov6providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil,
				},
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
		"teststep-providerfactories": {
			testCase: TestCase{},
			testStep: TestStep{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil,
				},
				Config: `
resource "test_test" "test" {}
`,
			},
			expected: `
resource "test_test" "test" {}
`,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.testStep.mergedConfig(context.Background(), testCase.testCase)

			if diff := cmp.Diff(strings.TrimSpace(got), strings.TrimSpace(testCase.expected)); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestStepProviderConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		testStep TestStep
		expected string
	}{
		"externalproviders-and-protov5providerfactories": {
			testStep: TestStep{
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
			testStep: TestStep{
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
			testStep: TestStep{
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
			testStep: TestStep{
				ExternalProviders: map[string]ExternalProvider{
					"test": {},
				},
			},
			expected: `provider "test" {}`,
		},
		"externalproviders-source-and-versionconstraint": {
			testStep: TestStep{
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
			testStep: TestStep{
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
			testStep: TestStep{
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
			testStep: TestStep{
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": nil,
				},
			},
			expected: ``,
		},
		"protov6providerfactories": {
			testStep: TestStep{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": nil,
				},
			},
			expected: ``,
		},
		"providerfactories": {
			testStep: TestStep{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": nil,
				},
			},
			expected: ``,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.testStep.providerConfig(context.Background())

			if diff := cmp.Diff(strings.TrimSpace(got), strings.TrimSpace(testCase.expected)); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTest_TestStep_ExternalProviders(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: "# not empty",
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source: "registry.terraform.io/hashicorp/null",
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ExternalProviders_DifferentProviders(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source: "registry.terraform.io/hashicorp/null",
					},
				},
			},
			{
				Config: `resource "random_pet" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"random": {
						Source: "registry.terraform.io/hashicorp/random",
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ExternalProviders_DifferentVersions(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source:            "registry.terraform.io/hashicorp/null",
						VersionConstraint: "3.1.0",
					},
				},
			},
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source:            "registry.terraform.io/hashicorp/null",
						VersionConstraint: "3.1.1",
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ExternalProviders_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			Steps: []TestStep{
				{
					Config: "# not empty",
					ExternalProviders: map[string]ExternalProvider{
						"testnonexistent": {
							Source: "registry.terraform.io/hashicorp/testnonexistent",
						},
					},
				},
			},
		})
	})
}

func TestTest_TestStep_ExternalProviders_NonHashiCorpNamespace(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				ExternalProviders: map[string]ExternalProvider{
					// This can be set to any provider outside the hashicorp namespace.
					// bflad/scaffoldingtest happens to be a published version of
					// terraform-provider-scaffolding-framework.
					"scaffoldingtest": {
						Source:            "registry.terraform.io/bflad/scaffoldingtest",
						VersionConstraint: "0.1.0",
					},
				},
				Config: `resource "scaffoldingtest_example" "test" {}`,
			},
		},
	})
}

func TestTest_TestStep_ExternalProvidersAndProviderFactories_NonHashiCorpNamespace(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
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
				Config: `
					resource "null_resource" "test" {}
					resource "scaffoldingtest_example" "test" {}
				`,
			},
		},
	})
}

func TestTest_TestStep_ExternalProviders_To_ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source:            "registry.terraform.io/hashicorp/null",
						VersionConstraint: "3.1.1",
					},
				},
			},
			{
				Config: `resource "null_resource" "test" {}`,
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
			},
		},
	})
}

func TestTest_TestStep_ExternalProviders_To_ProviderFactories_StateUpgraders(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source:            "registry.terraform.io/hashicorp/null",
						VersionConstraint: "3.1.1",
					},
				},
			},
			{
				Check:  TestCheckResourceAttr("null_resource.test", "id", "test-schema-version-1"),
				Config: `resource "null_resource" "test" {}`,
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
									SchemaVersion: 1, // null 3.1.3 is version 0
									StateUpgraders: []schema.StateUpgrader{
										{
											Type: cty.Object(map[string]cty.Type{
												"id":       cty.String,
												"triggers": cty.Map(cty.String),
											}),
											Upgrade: func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
												// null 3.1.3 sets the id attribute to a stringified random integer.
												// Double check that our resource wasn't created by this TestStep.
												id, ok := rawState["id"].(string)

												if !ok || id == "test" {
													return rawState, fmt.Errorf("unexpected rawState: %v", rawState)
												}

												rawState["id"] = "test-schema-version-1"

												return rawState, nil
											},
											Version: 0,
										},
									},
								},
							},
						}, nil
					},
				},
			},
		},
	})
}

func TestTest_TestStep_Taint(t *testing.T) {
	t.Parallel()

	var idOne, idTwo string

	Test(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"random": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"random_id": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId(time.Now().String())
								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							Schema: map[string]*schema.Schema{},
						},
					},
				}, nil
			},
		},
		Steps: []TestStep{
			{
				Config: `resource "random_id" "test" {}`,
				Check: ComposeAggregateTestCheckFunc(
					extractResourceAttr("random_id.test", "id", &idOne),
				),
			},
			{
				Taint:  []string{"random_id.test"},
				Config: `resource "random_id" "test" {}`,
				Check: ComposeAggregateTestCheckFunc(
					extractResourceAttr("random_id.test", "id", &idTwo),
				),
			},
		},
	})

	if idOne == idTwo {
		t.Errorf("taint is not causing destroy-create cycle, idOne == idTwo: %s == %s", idOne, idTwo)
	}
}

func extractResourceAttr(resourceName string, attributeName string, attributeValue *string) TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource name %s not found in state", resourceName)
		}

		attrValue, ok := rs.Primary.Attributes[attributeName]

		if !ok {
			return fmt.Errorf("attribute %s not found in resource %s state", attributeName, resourceName)
		}

		*attributeValue = attrValue

		return nil
	}
}

func TestTest_TestStep_ProtoV5ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		Steps: []TestStep{
			{
				Config: "# not empty",
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"test": func() (tfprotov5.ProviderServer, error) { //nolint:unparam // required signature
						return nil, nil
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ProtoV5ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			Steps: []TestStep{
				{
					Config: "# not empty",
					ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
						"test": func() (tfprotov5.ProviderServer, error) { //nolint:unparam // required signature
							return nil, fmt.Errorf("test")
						},
					},
				},
			},
		})
	})
}

func TestTest_TestStep_ProtoV6ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		Steps: []TestStep{
			{
				Config: "# not empty",
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"test": func() (tfprotov6.ProviderServer, error) { //nolint:unparam // required signature
						return nil, nil
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ProtoV6ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			Steps: []TestStep{
				{
					Config: "# not empty",
					ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
						"test": func() (tfprotov6.ProviderServer, error) { //nolint:unparam // required signature
							return nil, fmt.Errorf("test")
						},
					},
				},
			},
		})
	})
}

func TestTest_TestStep_ProviderFactories(t *testing.T) {
	t.Parallel()

	Test(&mockT{}, TestCase{
		Steps: []TestStep{
			{
				Config: "# not empty",
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
						return nil, nil
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Error(t *testing.T) {
	t.Parallel()

	testExpectTFatal(t, func() {
		Test(&mockT{}, TestCase{
			Steps: []TestStep{
				{
					Config: "# not empty",
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
							return nil, fmt.Errorf("test")
						},
					},
				},
			},
		})
	})
}

func TestTest_TestStep_ProviderFactories_To_ExternalProviders(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "null_resource" "test" {}`,
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
			},
			{
				Config: `resource "null_resource" "test" {}`,
				ExternalProviders: map[string]ExternalProvider{
					"null": {
						Source: "registry.terraform.io/hashicorp/null",
					},
				},
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_Inline(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		Steps: []TestStep{
			{
				Config: `resource "random_password" "test" { length = 12 }`,
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"random": func() (*schema.Provider, error) { //nolint:unparam // required signature
						return &schema.Provider{
							ResourcesMap: map[string]*schema.Resource{
								"random_password": {
									DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
										return nil
									},
									ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
										return nil
									},
									Schema: map[string]*schema.Schema{
										"length": {
											Required: true,
											ForceNew: true,
											Type:     schema.TypeInt,
										},
										"result": {
											Type:      schema.TypeString,
											Computed:  true,
											Sensitive: true,
										},

										"id": {
											Computed: true,
											Type:     schema.TypeString,
										},
									},
									Importer: &schema.ResourceImporter{
										StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
											val := d.Id()

											d.SetId("none")

											err := d.Set("result", val)
											if err != nil {
												panic(err)
											}

											err = d.Set("length", len(val))
											if err != nil {
												panic(err)
											}

											return []*schema.ResourceData{d}, nil
										},
									},
								},
							},
						}, nil
					},
				},
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testCheckResourceAttrInstanceState("id", "none"),
					testCheckResourceAttrInstanceState("result", "Z=:cbrJE?Ltg"),
					testCheckResourceAttrInstanceState("length", "12"),
				),
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_Inline_WithPersistMatch(t *testing.T) {
	var result1, result2 string

	t.Parallel()

	Test(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"random": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"random_password": {
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							Schema: map[string]*schema.Schema{
								"length": {
									Required: true,
									ForceNew: true,
									Type:     schema.TypeInt,
								},
								"result": {
									Type:      schema.TypeString,
									Computed:  true,
									Sensitive: true,
								},

								"id": {
									Computed: true,
									Type:     schema.TypeString,
								},
							},
							Importer: &schema.ResourceImporter{
								StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
									val := d.Id()

									d.SetId("none")

									err := d.Set("result", val)
									if err != nil {
										panic(err)
									}

									err = d.Set("length", len(val))
									if err != nil {
										panic(err)
									}

									return []*schema.ResourceData{d}, nil
								},
							},
						},
					},
				}, nil
			},
		},
		Steps: []TestStep{
			{
				Config:             `resource "random_password" "test" { length = 12 }`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				Config: `resource "random_password" "test" { length = 12 }`,
				Check: ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
				),
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_Inline_WithoutPersist(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"random": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"random_password": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("none")
								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							Schema: map[string]*schema.Schema{
								"length": {
									Required: true,
									ForceNew: true,
									Type:     schema.TypeInt,
								},
								"result": {
									Type:      schema.TypeString,
									Computed:  true,
									Sensitive: true,
								},

								"id": {
									Computed: true,
									Type:     schema.TypeString,
								},
							},
							Importer: &schema.ResourceImporter{
								StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
									val := d.Id()

									d.SetId("none")

									err := d.Set("result", val)
									if err != nil {
										panic(err)
									}

									err = d.Set("length", len(val))
									if err != nil {
										panic(err)
									}

									return []*schema.ResourceData{d}, nil
								},
							},
						},
					},
				}, nil
			},
		},
		Steps: []TestStep{
			{
				Config:             `resource "random_password" "test" { length = 12 }`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: false,
			},
			{
				Config: `resource "random_password" "test" { length = 12 }`,
				Check: ComposeTestCheckFunc(
					TestCheckNoResourceAttr("random_password.test", "result"),
				),
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_External(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"random": {
				Source: "registry.terraform.io/hashicorp/random",
			},
		},
		Steps: []TestStep{
			{
				Config:             `resource "random_password" "test" { length = 12 }`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testCheckResourceAttrInstanceState("id", "none"),
					testCheckResourceAttrInstanceState("result", "Z=:cbrJE?Ltg"),
					testCheckResourceAttrInstanceState("length", "12"),
				),
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_External_WithPersistMatch(t *testing.T) {
	var result1, result2 string

	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"random": {
				Source: "registry.terraform.io/hashicorp/random",
			},
		},
		Steps: []TestStep{
			{
				Config:             `resource "random_password" "test" { length = 12 }`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				Config: `resource "random_password" "test" { length = 12 }`,
				Check: ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
				),
			},
		},
	})
}

func TestTest_TestStep_ProviderFactories_Import_External_WithoutPersistNonMatch(t *testing.T) {
	var result1, result2 string

	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"random": {
				Source: "registry.terraform.io/hashicorp/random",
			},
		},
		Steps: []TestStep{
			{
				Config:             `resource "random_password" "test" { length = 12 }`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: false,
				ImportStateCheck: composeImportStateCheck(
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				Config: `resource "random_password" "test" { length = 12 }`,
				Check: ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
				),
			},
		},
	})
}

func composeImportStateCheck(fs ...ImportStateCheckFunc) ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

func testExtractResourceAttrInstanceState(attributeName string, attributeValue *string) ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("unexpected number of instance states: %d", len(is))
		}

		s := is[0]

		attrValue, ok := s.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found in instance state", attributeName)
		}

		*attributeValue = attrValue

		return nil
	}
}

func testCheckResourceAttrInstanceState(attributeName, attributeValue string) ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("unexpected number of instance states: %d", len(is))
		}

		s := is[0]

		attrVal, ok := s.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s found in instance state", attributeName)
		}

		if attrVal != attributeValue {
			return fmt.Errorf("expected: %s got: %s", attributeValue, attrVal)
		}

		return nil
	}
}

func testExtractResourceAttr(resourceName string, attributeName string, attributeValue *string) TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource name %s not found in state", resourceName)
		}

		attrValue, ok := rs.Primary.Attributes[attributeName]

		if !ok {
			return fmt.Errorf("attribute %s not found in resource %s state", attributeName, resourceName)
		}

		*attributeValue = attrValue

		return nil
	}
}

func testCheckAttributeValuesEqual(i *string, j *string) TestCheckFunc {
	return func(s *terraform.State) error {
		if testStringValue(i) != testStringValue(j) {
			return fmt.Errorf("attribute values are different, got %s and %s", testStringValue(i), testStringValue(j))
		}

		return nil
	}
}

func testCheckAttributeValuesDiffer(i *string, j *string) TestCheckFunc {
	return func(s *terraform.State) error {
		if testStringValue(i) == testStringValue(j) {
			return fmt.Errorf("attribute values are the same")
		}

		return nil
	}
}

func testStringValue(sPtr *string) string {
	if sPtr == nil {
		return ""
	}

	return *sPtr
}
