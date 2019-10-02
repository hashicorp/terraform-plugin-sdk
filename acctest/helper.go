package acctest

import (
	"os"

	tftest "github.com/apparentlymart/terraform-plugin-test"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func IfPluginServePlugin(providerFunc plugin.ProviderFunc) {
	if tftest.RunningAsPlugin() {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: providerFunc,
		})
		os.Exit(0)
	}
}
