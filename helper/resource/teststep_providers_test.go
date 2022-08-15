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

func TestStepProviderConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		testStep TestStep
		expected string
	}{
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
