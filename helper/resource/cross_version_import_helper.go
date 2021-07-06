package resource

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
	This helper is used to replace developing provider which contains a sdk upgrade and released provider.
    By deploying resources with released provider and importing resources with developing provider, to verify whether there's backend breaking change introduced.
*/

func isCrossVersionImportEnabled() bool {
	if enabled, err := strconv.ParseBool(os.Getenv("TF_ACC_CROSS_VERSION_IMPORT")); err == nil {
		return enabled
	}
	return false
}

func getStandardProviderName() string {
	if provider := os.Getenv("TF_ACC_PROVIDER"); provider != "" {
		return provider
	}
	return "azurerm"
}

func getStandardProviderNamespace() string {
	if namespace := os.Getenv("TF_ACC_PROVIDER_NAMESPACE"); namespace != "" {
		return namespace
	}
	return "hashicorp"
}

func getStandardProviderVersion() string {
	return os.Getenv("TF_ACC_PROVIDER_VERSION")
}

/*
replace develop provider with released provider
*/
func useExternalProvider(c TestCase) func() (*schema.Provider, error) {
	if isCrossVersionImportEnabled() {
		provider := getStandardProviderName()
		externalProvider := ExternalProvider{
			Source: fmt.Sprintf("registry.terraform.io/%s/%s", getStandardProviderNamespace(), provider),
		}
		if version := getStandardProviderVersion(); version != "" {
			externalProvider.VersionConstraint = "=" + version
		}
		c.ExternalProviders[provider] = externalProvider
		backupProviderFactory := c.ProviderFactories[provider]
		delete(c.ProviderFactories, provider)
		return backupProviderFactory
	}
	return nil
}

/*
replace released provider with develop provider
*/
func useDevelopProvider(c TestCase, backupProviderFactory func() (*schema.Provider, error)) {
	if isCrossVersionImportEnabled() {
		provider := getStandardProviderName()
		c.ProviderFactories[provider] = backupProviderFactory
		delete(c.ExternalProviders, provider)
	}
}
