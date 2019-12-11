package acctest

import (
	"os"

	tftest "github.com/apparentlymart/terraform-plugin-test"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

var TestHelper *tftest.Helper

func UseNewFramework(name string, providerFunc plugin.ProviderFunc) {
	sourceDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if tftest.RunningAsPlugin() {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: providerFunc,
		})
		os.Exit(0)
	} else {
		TestHelper = tftest.AutoInitProviderHelper(name, sourceDir)
	}
}
