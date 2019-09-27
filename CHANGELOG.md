# 1.1.0 (Unreleased)

FEATURES:

 * `schema.Provider.TerraformVersion` now defaults to "v0.11+compatible" to indicate when Terraform 0.10/0.11 CLI is communicating with the plugin. [GH-52]
 * `terraform plan` and `terraform apply` will now warn when the `-target` option is used, to draw attention to the fact that the result of applying the plan is likely to be incomplete, and to remind to re-run `terraform plan` with no targets afterwards to ensure that the configuration has converged. [GH-182]
 * config: New function `parseint` for parsing strings containing digits as integers in various bases. [GH-181]
 * config: New function `cidrsubnets`, which is a companion to the existing function `cidrsubnet` which can allocate multiple consecutive subnet prefixes (possibly of different prefix lengths) in a single call. [GH-187]
 
BUG FIXES:

 * Fix persistence of private data in acceptance tests. [GH-183]
 * command/import: fix error during import when implied provider was not used. [GH-184]
 * Fix evaluation errors when an indexed data source is evaluated during refresh. [GH-188]

# 1.0.0 (September 17, 2019)

This SDK is functionally equivalent to the "legacy" SDK in `hashicorp/terraform` [`v0.12.9`](https://github.com/hashicorp/terraform/blob/v0.12.9/CHANGELOG.md).

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/plugin-sdk.html) of the website.

FEATURES:

 * Add `meta` package which exposes the version of the SDK, replacing the `version` package which previously exposed the Terraform version ([#37](https://github.com/hashicorp/terraform-plugin-sdk/issues/37)] [[#24](https://github.com/hashicorp/terraform-plugin-sdk/issues/24))
