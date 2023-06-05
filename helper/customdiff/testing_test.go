// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customdiff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testProvider(s map[string]*schema.Schema, cd schema.CustomizeDiffFunc) *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"test": {
				Schema:        s,
				CustomizeDiff: cd,
			},
		},
	}
}

func testDiff(provider *schema.Provider, oldValue, newValue map[string]string) (*terraform.InstanceDiff, error) {
	newI := make(map[string]interface{}, len(newValue))
	for k, v := range newValue {
		newI[k] = v
	}

	return provider.ResourcesMap["test"].Diff(
		context.Background(),
		&terraform.InstanceState{
			Attributes: oldValue,
		},
		&terraform.ResourceConfig{
			Config: newI,
		},
		provider.Meta(),
	)
}
