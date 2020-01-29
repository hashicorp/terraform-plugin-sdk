# 1.6.0 (January 29, 2020)

DEPRECATED:

* helper/validation: `ValidateListUniqueStrings` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `SingleIP` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `IPRange` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `CIDRNetwork` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `ValidateJsonString` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `ValidateRegexp` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `ValidateRFC3339TimeString` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))

FEATURES:

* helper/validation: `IntDivisibleBy` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IntNotInSlice` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsIPv6Address` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsIPv4Address` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsCIDR` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsMACAddress` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsPortNumber` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsPortNumberOrZero` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsDayOfTheWeek` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsMonth` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsRFC3339Time` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsURLWithHTTPS` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsURLWithHTTPorHTTPS` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `IsURLWithScheme` ([#296](https://github.com/hashicorp/terraform-plugin-sdk/issues/296))
* helper/validation: `ListOfUniqueStrings` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `IsIPAddress` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `IsIPv4Range` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `IsCIDRNetwork` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `StringIsJSON` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))
* helper/validation: `StringIsValidRegExp` ([#301](https://github.com/hashicorp/terraform-plugin-sdk/issues/301))

# 1.5.0 (January 16, 2020)

FEATURES: 

* helper/validation: `StringIsEmpty` ([#294](https://github.com/hashicorp/terraform-plugin-sdk/issues/294))
* helper/validation: `StringIsNotEmpty` ([#294](https://github.com/hashicorp/terraform-plugin-sdk/issues/294))
* helper/validation: `StringIsWhiteSpace` ([#294](https://github.com/hashicorp/terraform-plugin-sdk/issues/294))
* helper/validation: `StringIsNotWhiteSpace` ([#294](https://github.com/hashicorp/terraform-plugin-sdk/issues/294))
* helper/validation: `IsUUID` ([#294](https://github.com/hashicorp/terraform-plugin-sdk/issues/294)) ([#297](https://github.com/hashicorp/terraform-plugin-sdk/issues/297))

BUG FIXES:

* schema/ExactlyOneOf: Fix handling of unknowns in complex types ([#287](https://github.com/hashicorp/terraform-plugin-sdk/issues/287))

# 1.4.1 (December 18, 2019)

BUG FIXES:

* helper/resource: Don't crash when dependent test sweeper is missing ([#279](https://github.com/hashicorp/terraform-plugin-sdk/issues/279))

# 1.4.0 (November 20, 2019)

NOTES:

* pruned dead code from internal pkg ([#251](https://github.com/hashicorp/terraform-plugin-sdk/issues/251))
* bumped dependency of `terraform-config-inspect` to remove transitive dependency ([#252](https://github.com/hashicorp/terraform-plugin-sdk/issues/252))

FEATURES: 

* helper/validation: Add `FloatAtLeast` and `FloatAtMost` validation functions ([#239](https://github.com/hashicorp/terraform-plugin-sdk/issues/239))
* helper/validation: Add `StringDoesNotMatch` validation function ([#240](https://github.com/hashicorp/terraform-plugin-sdk/issues/240))
* ResourceData: Add `HasChanges` variadic method ([#241](https://github.com/hashicorp/terraform-plugin-sdk/issues/241))

# 1.3.0 (November 06, 2019)

NOTES:

* The internalized version of Terraform that exists for the acceptance test framework has received several cherry picks in an effort to keep it in sync with how Terraform behaves. This process is performed on a best effort basis.

FEATURES: 

* helper/validation: Add `StringDoesNotContainAny` validation function ([#212](https://github.com/hashicorp/terraform-plugin-sdk/issues/212))
* helper/schema: Introduce `ExactlyOneOf` and `AtLeastOneOf` validation checks against schema attributes ([#225](https://github.com/hashicorp/terraform-plugin-sdk/issues/225))

BUG FIXES:

* helper/resource: Ensure dependent sweepers are all added. ([#213](https://github.com/hashicorp/terraform-plugin-sdk/issues/213))

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
