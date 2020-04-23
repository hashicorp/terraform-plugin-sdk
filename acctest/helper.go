package acctest

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

var TestHelper *tftest.Helper

var TestProviderFunc plugin.ProviderFunc

func UseBinaryDriver(name string, providerFunc plugin.ProviderFunc) {
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
		TestProviderFunc = providerFunc
		binDir := filepath.Join(sourceDir, ".terraform", "plugins", "registry.terraform.io", "hashicorp", name, "v999.999.999", runtime.GOOS+"_"+runtime.GOARCH)
		err = os.MkdirAll(binDir, 0777)
		if err != nil {
			panic(err)
		}
		f, err := os.OpenFile(filepath.Join(binDir, "terraform-plugin-"+name), os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
}
