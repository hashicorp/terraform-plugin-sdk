package customdiff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testProvider(s map[string]*schema.Schema, cd schema.CustomizeDiffFunc) terraform.ResourceProvider {
	return (&schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"test": {
				Schema: s,
				CustomizeDiffContext: func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
					return cd(d, meta)
				},
			},
		},
	})
}

func testDiff(provider terraform.ResourceProvider, old, new map[string]string) (*terraform.InstanceDiff, error) {
	newI := make(map[string]interface{}, len(new))
	for k, v := range new {
		newI[k] = v
	}

	return provider.Diff(
		&terraform.InstanceInfo{
			Id:         "test",
			Type:       "test",
			ModulePath: []string{},
		},
		&terraform.InstanceState{
			Attributes: old,
		},
		&terraform.ResourceConfig{
			Config: newI,
		},
	)
}
