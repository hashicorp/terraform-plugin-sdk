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

func testDiff(provider *schema.Provider, old, new map[string]string) (*terraform.InstanceDiff, error) {
	newI := make(map[string]interface{}, len(new))
	for k, v := range new {
		newI[k] = v
	}

	return provider.ResourcesMap["test"].Diff(
		context.Background(),
		&terraform.InstanceState{
			Attributes: old,
		},
		&terraform.ResourceConfig{
			Config: newI,
		},
		provider.Meta(),
	)
}
