# 1.16.1 (March 24, 2021)

BUG FIXES:

* Backported #591, making sure the pre-destroy state was passed to CheckDestroy, instead of the post-destroy state ([#728](https://github.com/hashicorp/terraform-plugin-sdk/issues/728))
* Updated import testing when using binary testing to work with Terraform 0.13 and above, with registry addresses in state. ([#702](https://github.com/hashicorp/terraform-plugin-sdk/issues/702))

# 1.16.0 (September 24, 2020)

FEATURES:

* Backported reattach mode for binary acceptance testing. Set `TF_ACCTEST_REATTACH` to `1` when using binary testing to enable reattach mode, which will allow debugging the provider under test and accurate test coverage results. ([#527](https://github.com/hashicorp/terraform-plugin-sdk/issues/527))

ENHANCEMENTS:

* Improved plan output for unexpected diffs when using binary testing ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))

BUG FIXES:

* Fixed a bug with binary testing that would use the wrong state when verifying import state ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Restored TestStep numbers in various outputs for binary testing ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Made resource detection when verifying import state more robust ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Removed excessive logging when using binary acceptance testing ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Fixed a bug that would sometimes bypass ExpectNonEmptyError during binary testing ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Fixed binary testing to respect `TestStep.Destroy` and more accurately mirror the legacy testing behavior ([#553](https://github.com/hashicorp/terraform-plugin-sdk/issues/553))
* Fixed a bug with ExpectNonEmptyPlan tests when using binary testing ([#590](https://github.com/hashicorp/terraform-plugin-sdk/issues/590))
* Surfaced errors when running destroy after tests when using binary testing ([#590](https://github.com/hashicorp/terraform-plugin-sdk/issues/590))

# 1.15.0 (July 08, 2020)

FEATURES:

* The binary test driver will now automatically install and verify the signature of a `terraform` binary if needed ([#491](https://github.com/hashicorp/terraform-plugin-sdk/issues/491))

# 1.14.0 (June 17, 2020)

FEATURES:

* Bump hashicorp/go-plugin to v1.2.0 which should enable grpc reflection ([#468](https://github.com/hashicorp/terraform-plugin-sdk/issues/468))

# 1.13.1 (June 04, 2020)

BUG FIXES:

* Remove deprecation for `d.Partial` ([#463](https://github.com/hashicorp/terraform-plugin-sdk/issues/463))
* Fix bug when serializing bool in TypeMap ([#465](https://github.com/hashicorp/terraform-plugin-sdk/issues/465))

# 1.13.0 (May 20, 2020)

DEPRECATIONS:

* Deprecate `DisableBinaryDriver` ([#450](https://github.com/hashicorp/terraform-plugin-sdk/issues/450))
* Deprecate the `helper/mutexkv`, `helper/pathorcontents`, `httpclient`, and `helper/hashcode` packages ([#453](https://github.com/hashicorp/terraform-plugin-sdk/issues/453))

# 1.12.0 (May 06, 2020)

FEATURES:

* Allow disabling binary testing via `TF_DISABLE_BINARY_TESTING` environment variable. ([#441](https://github.com/hashicorp/terraform-plugin-sdk/issues/441))

BUG FIXES:

* More accurate results for `schema.ResourceData.HasChange` when dealing with a Set inside another Set. ([#362](https://github.com/hashicorp/terraform-plugin-sdk/issues/362))

DEPRECATED:

* helper/encryption: In line with sensitive state best practices, the `helper/encryption` package is deprecated. ([#437](https://github.com/hashicorp/terraform-plugin-sdk/issues/437))

# 1.11.0 (April 30, 2020)

ENHANCEMENTS:

* Better error messaging when indexing into TypeSet for test checks, while the binary driver is enabled (currently not supported) ([#417](https://github.com/hashicorp/terraform-plugin-sdk/issues/417))
* Prevent ConflictsWith from self referencing and prevent referencing multi item Lists or Sets ([#416](https://github.com/hashicorp/terraform-plugin-sdk/issues/416)] [[#423](https://github.com/hashicorp/terraform-plugin-sdk/issues/423)] [[#426](https://github.com/hashicorp/terraform-plugin-sdk/issues/426))

# 1.10.0 (April 23, 2020)

FEATURES:

* Added validation helper `RequiredWith` ([#342](https://github.com/hashicorp/terraform-plugin-sdk/issues/342))

BUG FIXES:

* Binary acceptance test driver: omit test cleanup when state is empty ([#356](https://github.com/hashicorp/terraform-plugin-sdk/issues/356))
* Make mockT.Fatal halt execution ([#396](https://github.com/hashicorp/terraform-plugin-sdk/issues/396))

DEPENDENCIES:

* `github.com/hashicorp/terraform-plugin-test@v1.2.0` -> `v1.3.0` [[#400](https://github.com/hashicorp/terraform-plugin-sdk/issues/400)] 

# 1.9.1 (April 09, 2020)

BUG FIXES:

* Binary acceptance test driver: fix cleanup of temporary directories ([#378](https://github.com/hashicorp/terraform-plugin-sdk/issues/378))

# 1.9.0 (March 26, 2020)

DEPRECATED:

* helper/schema: `ResourceData.GetOkExists` will not be removed in the next major version unless a suitable replacement or alternative can be prescribed ([#350](https://github.com/hashicorp/terraform-plugin-sdk/issues/350))

FEATURES:

* Added support for additional protocol 5.2 fields (`Description`, `DescriptionKind`, `Deprecated`) ([#353](https://github.com/hashicorp/terraform-plugin-sdk/issues/353))

BUG FIXES:

* Binary acceptance test driver: auto-configure providers ([#355](https://github.com/hashicorp/terraform-plugin-sdk/issues/355))

# 1.8.0 (March 11, 2020)

FEATURES:

* helper/validation: `StringNotInSlice` ([#341](https://github.com/hashicorp/terraform-plugin-sdk/issues/341))

# 1.7.0 (February 12, 2020)

FEATURES:

* Binary acceptance test driver ([#262](https://github.com/hashicorp/terraform-plugin-sdk/issues/262))

DEPRECATED:

* helper/schema: `ResourceData.Partial` ([#317](https://github.com/hashicorp/terraform-plugin-sdk/issues/317))
* helper/schema: `ResourceData.SetPartial` ([#317](https://github.com/hashicorp/terraform-plugin-sdk/issues/317))

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
