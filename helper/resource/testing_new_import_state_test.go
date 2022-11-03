package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTest_TestStep_ImportStateCheck_SkipDataSourceState(t *testing.T) {
	t.Parallel()

	UnitTest(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"examplecloud": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					DataSourcesMap: map[string]*schema.Resource{
						"examplecloud_thing": {
							ReadContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("datasource-test")

								return nil
							},
							Schema: map[string]*schema.Schema{
								"id": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
					ResourcesMap: map[string]*schema.Resource{
						"examplecloud_thing": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("resource-test")

								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							Schema: map[string]*schema.Schema{
								"id": {
									Computed: true,
									Type:     schema.TypeString,
								},
							},
							Importer: &schema.ResourceImporter{
								StateContext: schema.ImportStatePassthroughContext,
							},
						},
					},
				}, nil
			},
		},
		Steps: []TestStep{
			{
				Config: `
					data "examplecloud_thing" "test" {}
					resource "examplecloud_thing" "test" {}
				`,
			},
			{
				ResourceName: "examplecloud_thing.test",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					if len(is) > 1 {
						return fmt.Errorf("expected 1 state, got: %d", len(is))
					}

					return nil
				},
			},
		},
	})
}
