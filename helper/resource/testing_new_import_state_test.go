// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"fmt"
	"regexp"
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

func TestTest_TestStep_ImportStateVerify(t *testing.T) {
	t.Parallel()

	UnitTest(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"examplecloud": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"examplecloud_thing": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("resource-test")

								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								_ = d.Set("other", "testvalue")

								return nil
							},
							Schema: map[string]*schema.Schema{
								"other": {
									Computed: true,
									Type:     schema.TypeString,
								},
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
				Config: `resource "examplecloud_thing" "test" {}`,
			},
			{
				ResourceName:      "examplecloud_thing.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestTest_TestStep_ImportStateVerifyIgnore(t *testing.T) {
	t.Parallel()

	UnitTest(t, TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"examplecloud": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"examplecloud_thing": {
							CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								d.SetId("resource-test")

								_ = d.Set("create_only", "testvalue")

								return nil
							},
							DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
								return nil
							},
							ReadContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
								_ = d.Set("read_only", "testvalue")

								return nil
							},
							Schema: map[string]*schema.Schema{
								"create_only": {
									Computed: true,
									Type:     schema.TypeString,
								},
								"read_only": {
									Computed: true,
									Type:     schema.TypeString,
								},
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
				Config: `resource "examplecloud_thing" "test" {}`,
			},
			{
				ResourceName:            "examplecloud_thing.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_only"},
			},
		},
	})
}

func TestTest_TestStep_ExpectError_ImportState(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"random": {
				Source:            "registry.terraform.io/hashicorp/time",
				VersionConstraint: "0.9.1",
			},
		},
		Steps: []TestStep{
			{
				Config:        `resource "time_static" "one" {}`,
				ImportStateId: "invalid time string",
				ResourceName:  "time_static.one",
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Error: Import time static error`),
			},
		},
	})
}
