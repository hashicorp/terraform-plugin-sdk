package acctest

import (
	"os"

	tftest "github.com/apparentlymart/terraform-plugin-test"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

var TestHelper *tftest.Helper
var ProviderName string

func UseNewFramework(name string, providerFunc plugin.ProviderFunc) {
	if tftest.RunningAsPlugin() {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: providerFunc,
		})
		os.Exit(0)
	} else {
		ProviderName = name
		TestHelper = tftest.AutoInitProviderHelper(name)
	}
}
