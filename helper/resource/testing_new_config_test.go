// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"regexp"
	"testing"
)

func TestTest_TestStep_ExpectError_NewConfig(t *testing.T) {
	t.Parallel()

	Test(t, TestCase{
		ExternalProviders: map[string]ExternalProvider{
			"random": {
				Source:            "registry.terraform.io/hashicorp/random",
				VersionConstraint: "3.4.3",
			},
		},
		Steps: []TestStep{
			{
				Config: `resource "random_string" "one" {
					length = 2
					min_upper = 4
				}`,
				ExpectError: regexp.MustCompile(`Error: Invalid Attribute Value`),
			},
		},
	})
}
