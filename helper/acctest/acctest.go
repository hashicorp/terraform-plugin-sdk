// Package acctest contains for Terraform Acceptance Tests
package acctest

import (
	"os"

	tftest "github.com/apparentlymart/terraform-plugin-test"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func InitProviderTesting(name string, providerFunc plugin.ProviderFunc) *tftest.Helper {
	if tftest.RunningAsPlugin() {
		// The test program is being re-launched as a provider plugin via our
		// stub program.
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: providerFunc,
		})
		os.Exit(0)
	}

	return tftest.AutoInitProviderHelper(name)
}
