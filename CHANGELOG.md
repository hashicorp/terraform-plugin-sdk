# 1.2.0 (October 25, 2019)

FEATURES:

* helper/resource: Introduce sweeper flag `-sweep-allow-failures` to continue other sweepers after failures ([#198](https://github.com/hashicorp/terraform-plugin-sdk/issues/198))

# 1.1.1 (October 03, 2019)

BUG FIXES:

 * `SDKVersion` in v1.1.0 was incorrectly set to "1.0.0" due to a bug in the release script. Fix for versions beginning at v1.1.1. ([#191](https://github.com/hashicorp/terraform-plugin-sdk/issues/191))

# 1.1.0 (September 27, 2019)

FEATURES:

 * `schema.Provider.TerraformVersion` now defaults to "0.11+compatible" to indicate when Terraform 0.10/0.11 CLI is communicating with the plugin. ([#52](https://github.com/hashicorp/terraform-plugin-sdk/issues/52))
 * `terraform plan` and `terraform apply` will now warn when the `-target` option is used, to draw attention to the fact that the result of applying the plan is likely to be incomplete, and to remind to re-run `terraform plan` with no targets afterwards to ensure that the configuration has converged. ([#182](https://github.com/hashicorp/terraform-plugin-sdk/issues/182))
 * config: New function `parseint` for parsing strings containing digits as integers in various bases. ([#181](https://github.com/hashicorp/terraform-plugin-sdk/issues/181))
 * config: New function `cidrsubnets`, which is a companion to the existing function `cidrsubnet` which can allocate multiple consecutive subnet prefixes (possibly of different prefix lengths) in a single call. ([#187](https://github.com/hashicorp/terraform-plugin-sdk/issues/187))
 
BUG FIXES:

 * Fix persistence of private data in acceptance tests. ([#183](https://github.com/hashicorp/terraform-plugin-sdk/issues/183))
 * command/import: fix error during import when implied provider was not used. ([#184](https://github.com/hashicorp/terraform-plugin-sdk/issues/184))
 * Fix evaluation errors when an indexed data source is evaluated during refresh. ([#188](https://github.com/hashicorp/terraform-plugin-sdk/issues/188))

# 1.0.0 (September 17, 2019)

This SDK is functionally equivalent to the "legacy" SDK in `hashicorp/terraform` [`v0.12.9`](https://github.com/hashicorp/terraform/blob/v0.12.9/CHANGELOG.md).

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/plugin-sdk.html) of the website.

FEATURES:

 * Add `meta` package which exposes the version of the SDK, replacing the `version` package which previously exposed the Terraform version ([#37](https://github.com/hashicorp/terraform-plugin-sdk/issues/37)] [[#24](https://github.com/hashicorp/terraform-plugin-sdk/issues/24))
