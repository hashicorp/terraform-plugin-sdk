# 2.10.2 (Unreleased)

NOTES:

* helper/schema: Started using terraform-plugin-log to write some SDK-level logs. Very few logs use this functionality now, but in the future, the environment variable `TF_LOG_SDK_HELPER_SCHEMA` will be able to set the log level for the SDK separately from the provider.

# 2.10.1 (December 17, 2021)

BUG FIXES:

* helper/schema: Fixed regression from version 2.9.0 in `(ResourceDiff).GetChangedKeysPrefix()` where passing an empty string (`""`) would no longer return all changed keys ([#829](https://github.com/hashicorp/terraform-plugin-sdk/issues/829))

# 2.10.0 (December 07, 2021)

NOTES:

* helper/resource: Previously, TF_ACC_LOG_PATH would not enable logging for the provider under test. This has been fixed, so logging from the Terraform binary, any external providers, and the provider under test will all be combined in a file at the specified path.

ENHANCEMENTS:

* Upgraded to terraform-plugin-go v0.5.0 ([#805](https://github.com/hashicorp/terraform-plugin-sdk/issues/805))
* Added support for terraform-plugin-log ([#805](https://github.com/hashicorp/terraform-plugin-sdk/issues/805))

# 2.9.0 (November 19, 2021)

NOTES:

* helper/schema: Added warning log for provider reconfiguration, which can occur with concurrent testing and cause unexpected testing results when there are differing provider configurations. To prevent this warning, testing should create separate provider instances for separate configurations. Providers can further implement [`sync.Once`](https://pkg.go.dev/sync#Once) to prevent reconfiguration effects or add an execution tracking variable in `Provider.ConfigureFunc` or `Provider.ConfigureContextFunc` implementations to raise errors, if desired. ([#636](https://github.com/hashicorp/terraform-plugin-sdk/issues/636))

ENHANCEMENTS:

* helper/resource: Added timing logging to sweepers ([#782](https://github.com/hashicorp/terraform-plugin-sdk/issues/782))
* helper/resource: Updated terraform-exec to work with Terraform 1.1 ([#822](https://github.com/hashicorp/terraform-plugin-sdk/issues/822))

BUG FIXES:

* helper/acctest: Prevent duplicate values from `RandInt()`, `RandIntRange()`, and `RandomWithPrefix()` invocations on platforms with less granular clocks ([#764](https://github.com/hashicorp/terraform-plugin-sdk/issues/764))
* helper/schema: Prevent potential panics with `(*ResourceData).HasChangeExcept()` and `(*ResourceData).HasChangesExcept()` ([#811](https://github.com/hashicorp/terraform-plugin-sdk/issues/811))
* helper/schema: Remove `TypeSet` truncation warning logs if none are truncated ([#767](https://github.com/hashicorp/terraform-plugin-sdk/issues/767))
* helper/schema: Ensure `(*ResourceDiff).SetNew()` and `(*ResourceDiff).SetNewComputed()` only remove planned differences from exact or nested attribute and block names instead of any name with the same prefix ([#716](https://github.com/hashicorp/terraform-plugin-sdk/issues/716))
* helper/schema: Fix deep equality checks with `(*ResourceData).HasChange()`, `(*ResourceData).HasChanges()`, `(*ResourceData).HasChangeExcept()`, and `(*ResourceData).HasChangesExcept()` ([#711](https://github.com/hashicorp/terraform-plugin-sdk/issues/711))
* helper/schema: Prevent potential panics since v2.8.0 with data sources that have optional attributes and no practitioner configuration ([#815](https://github.com/hashicorp/terraform-plugin-sdk/issues/815))

# 2.8.0 (September 24, 2021)

NOTES:

* Updated to [terraform-plugin-go v0.4.0](https://github.com/hashicorp/terraform-plugin-go/blob/main/CHANGELOG.md#040-september-24-2021). Users of terraform-plugin-mux will need to upgrade terraform-plugin-mux as well.

ENHANCEMENTS:

* Added experimental support for retrieving underlying raw protocol values from `helper/schema.ResourceData` and `helper/schema.ResourceDiff`, bypassing the shims. ([#802](https://github.com/hashicorp/terraform-plugin-sdk/issues/802))

# 2.7.1 (August 31, 2021)

BUG FIXES:

* helper/schema: Ensure `Provider.ConfigureContextFunc` warning-only diagnostics are returned ([#791](https://github.com/hashicorp/terraform-plugin-sdk/issues/791))

# 2.7.0 (June 25, 2021)

ENHANCEMENTS:

* Added `ProtoV6ProviderFactories` to `TestCase`, so protocol version 6 providers can be used in acceptance tests ([#761](https://github.com/hashicorp/terraform-plugin-sdk/issues/761))
* Made SDK-generated diagnostics clearer and more consistent ([#755](https://github.com/hashicorp/terraform-plugin-sdk/issues/755))
* Upgraded to use terraform-exec v0.14.0, which is required for acceptance test compatibility with Terraform v1.0.1 ([#775](https://github.com/hashicorp/terraform-plugin-sdk/issues/775))

# 2.6.1 (April 23, 2021)

BUG FIXES:

* Updated the GPG key used to verify Terraform installs in response to the [Terraform GPG key rotation](https://discuss.hashicorp.com/t/hcsec-2021-12-codecov-security-event-and-hashicorp-gpg-key-exposure/23512). ([#750](https://github.com/hashicorp/terraform-plugin-sdk/issues/750))

# 2.6.0 (April 21, 2021)

ENHANCEMENTS:

* Made TF_ACC_TERRAFORM_VERSION more permissive, accepting values in either vX.Y.Z or X.Y.Z formats. ([#731](https://github.com/hashicorp/terraform-plugin-sdk/issues/731))
* Upgraded to use terraform-plugin-go v0.3.0 ([#739](https://github.com/hashicorp/terraform-plugin-sdk/issues/739))

# 2.5.0 (March 24, 2021)

ENHANCEMENTS

* Added the ability to opt out of context timeouts in CRUD functions ([#723](https://github.com/hashicorp/terraform-plugin-sdk/issues/723))

# 2.4.4 (February 24, 2021)

NOTES

As per our Go version support policy, we now require Go 1.15 or higher to use the SDK.

BUG FIXES

* Resolved bug where Diagnostics wouldn't get associated with their configuration context in user output. ([#696](https://github.com/hashicorp/terraform-plugin-sdk/issues/696))

# 2.4.3 (February 10, 2021)

BUG FIXES

* Make acceptance testing framework compatible with Terraform 0.15 ([#694](https://github.com/hashicorp/terraform-plugin-sdk/issues/694))

# 2.4.2 (January 27, 2021)

BUG FIXES

* Don't panic in very specific circumstances involving CustomizeDiff and empty strings in the config ([#686](https://github.com/hashicorp/terraform-plugin-sdk/issues/686))

# 2.4.1 (January 20, 2021)

BUG FIXES

* Don't panic during assertions when testing sets with varying levels of nesting ([#648](https://github.com/hashicorp/terraform-plugin-sdk/issues/648))
* Prevent panics when sending Ctrl-C to Terraform ([#674](https://github.com/hashicorp/terraform-plugin-sdk/issues/674))
* Make the error message when a "required" block is missing clearer, identifying the block in question ([#672](https://github.com/hashicorp/terraform-plugin-sdk/issues/672))

# 2.4.0 (December 19, 2020)

ENHANCEMENTS

* Support `Unwrap` on SDK errors ([#647](https://github.com/hashicorp/terraform-plugin-sdk/issues/647))
* Allow for `nil` errors in `diag.FromErr` ([#623](https://github.com/hashicorp/terraform-plugin-sdk/issues/623))
* Added `validation.ToDiagFunc` helper to translate legacy validation functions into Diagnostics-aware validation functions. ([#611](https://github.com/hashicorp/terraform-plugin-sdk/issues/611))
* Disable Checkpoint network connections during acceptance testing unless a Terraform binary needs to be installed. ([#663](https://github.com/hashicorp/terraform-plugin-sdk/issues/663))

BUG FIXES

* Check for `nil` errors prior to invoking `ErrorCheck` ([#646](https://github.com/hashicorp/terraform-plugin-sdk/issues/646))
* More reliable handling of logging ([#639](https://github.com/hashicorp/terraform-plugin-sdk/issues/639))
* Modified error text to make golint and go vet happy when a non-empty plan is found in testing and an empty plan was expected ([#596](https://github.com/hashicorp/terraform-plugin-sdk/issues/596))
* Add `UseJSONNumber` to `helper/schema.Resource` to make it possible to represent large numbers precisely. Setting to `true` will make numbers appear as `json.Number` in `StateUpgrader`s instead of as `float64`. ([#662](https://github.com/hashicorp/terraform-plugin-sdk/issues/662))
* Fix logs sometimes appearing in test output when running acceptance tests. ([#665](https://github.com/hashicorp/terraform-plugin-sdk/issues/665))

NOTES

We have removed the deprecation of the non-diagnostic version of validation until the build-in validations are ported to the new format.

# 2.3.0 (November 20, 2020)

ENHANCEMENTS

* `helper/schema.ResourceData` now has `HasChangeExcept` and `HasChangesExcept` methods to check if the resource has changes _besides_ a given key or list of keys. ([#558](https://github.com/hashicorp/terraform-plugin-sdk/issues/558))
* `helper/resource.TestCase` now has an `ErrorCheck` property that can be set to a function, allowing the programmatic determination of whether to ignore an error or not. ([#592](https://github.com/hashicorp/terraform-plugin-sdk/issues/592))

# 2.2.0 (November 02, 2020)

FEATURES
* Updated to use the new [`terraform-plugin-go`](https://github.com/hashicorp/terraform-plugin-go) library as a foundation for the SDK, enabling it to be used with [`terraform-plugin-mux`](https://github.com/hashicorp/terraform-plugin-mux) ([#630](https://github.com/hashicorp/terraform-plugin-sdk/issues/630))
* Added the `TestCase.ProtoV5ProviderFactories` property to allow testing providers created with `terraform-plugin-go` with the `helper/resource` test framework. ([#630](https://github.com/hashicorp/terraform-plugin-sdk/issues/630))

# 2.1.0 (October 27, 2020)

FEATURES
* Relaxed validation in `InternalValidate` for explicitly set `id` attributes ([#613](https://github.com/hashicorp/terraform-plugin-sdk/issues/613))
* Ported TypeSet test check funcs essential for migrating to V2 of the SDK ([#614](https://github.com/hashicorp/terraform-plugin-sdk/issues/614))
* Improved debug output for how to manually invoke the Terraform CLI ([#615](https://github.com/hashicorp/terraform-plugin-sdk/issues/615))

# 2.0.4 (October 06, 2020)

BUG FIXES
* Fix a bug that would pass the post-destroy state to `helper/resource.TestCase.CheckDestroy` instead of the documented pre-destroy state ([#591](https://github.com/hashicorp/terraform-plugin-sdk/issues/591))
* Clean up the final remaining places where test numbers or dangling resources warnings could be omitted from errors ([#578](https://github.com/hashicorp/terraform-plugin-sdk/issues/578))
* Stop considering plans empty when they include data source changes ([#594](https://github.com/hashicorp/terraform-plugin-sdk/issues/594))

# 2.0.3 (September 15, 2020)

BUG FIXES

* Fixed a bug that would incorrectly mark tests using TestStep.ImportStateVerify as failed if they tested a resource with custom timeouts ([#576](https://github.com/hashicorp/terraform-plugin-sdk/issues/576))
* Fixed a bug where errors destroying infrastructure after tests wouldn't be reported ([#581](https://github.com/hashicorp/terraform-plugin-sdk/issues/581))
* Fixed a bug where test steps that expected a non-empty plan would fail because they had an empty plan, erroneously ([#580](https://github.com/hashicorp/terraform-plugin-sdk/issues/580))
* Fixed a bug where the plan output shown when an unexpected diff was encountered during testing would be shown in JSON instead of a human-readable format ([#584](https://github.com/hashicorp/terraform-plugin-sdk/issues/584))

# 2.0.2 (September 10, 2020)

BUG FIXES

* Fixed bug where state is read from the wrong workspace during import tests. ([#552](https://github.com/hashicorp/terraform-plugin-sdk/issues/552))
* Fixed bug where the resource could belong to another provider when finding the resource state to check during import tests ([#522](https://github.com/hashicorp/terraform-plugin-sdk/issues/522))
* Removed excessive logging when ExpectNonEmptyPlan was successfully matched ([#556](https://github.com/hashicorp/terraform-plugin-sdk/issues/556))
* Fixed bug where state from data sources, which can't be imported, would be surfaced during ImportStateVerify ([#555](https://github.com/hashicorp/terraform-plugin-sdk/issues/555))
* Fixed bug that ignored ExpectError when testing state imports ([#550](https://github.com/hashicorp/terraform-plugin-sdk/issues/550))
* Fixed bug that sometimes prevented TestStep numbers from appearing in error output ([#557](https://github.com/hashicorp/terraform-plugin-sdk/issues/557))
* Fixed bug that would ignore `TestStep.Destroy` when running tests. ([#563](https://github.com/hashicorp/terraform-plugin-sdk/issues/563))

# 2.0.1 (August 10, 2020)

BUG FIXES

* Restored reporting of failed test step number ([#524](https://github.com/hashicorp/terraform-plugin-sdk/issues/524))
* Restored output of a test that failed with unexpected diff to V1 style output ([#526](https://github.com/hashicorp/terraform-plugin-sdk/issues/526))

# 2.0.0 (July 30, 2020)

FEATURES

* Provide deprecated method for receiving a global context that receives stop cancellation. ([#502](https://github.com/hashicorp/terraform-plugin-sdk/issues/502))
* Support multiple providers in reattach mode ([#512](https://github.com/hashicorp/terraform-plugin-sdk/issues/512))
* Allow setting `ExternalProviders` in `resource.TestCase` to control what providers are downloaded with `terraform init` for a test. ([#516](https://github.com/hashicorp/terraform-plugin-sdk/issues/516))
* Restore `resource.TestEnvVar` ([#519](https://github.com/hashicorp/terraform-plugin-sdk/issues/519))

BUG FIXES

* Remove deprecation warnings which cause spam and crashes in provider acceptance tests. ([#503](https://github.com/hashicorp/terraform-plugin-sdk/issues/503))
* Fixed a bug in the test driver that caused errors for Windows users on versions of Terraform below 0.13.0-beta2. ([#499](https://github.com/hashicorp/terraform-plugin-sdk/issues/499))
* Fixed a bug in the test driver that caused timeouts when using the `IDRefreshName` on `resource.TestCase`s. ([#501](https://github.com/hashicorp/terraform-plugin-sdk/issues/501))
* Fixed a bug where data sources would not always reflect changes in their configs in the same `resource.TestStep` that the config changed. ([#515](https://github.com/hashicorp/terraform-plugin-sdk/issues/515))
* Fixed a bug that would prevent errors from being handled by ExpectError handlers during testing. ([#518](https://github.com/hashicorp/terraform-plugin-sdk/issues/518))

# 2.0.0-rc.2 (June 11, 2020)

FEATURES

* The test driver was reworked to allow for test coverage, race detection, and debugger support. ([#471](https://github.com/hashicorp/terraform-plugin-sdk/issues/471))
* A new `plugin.Debug` function allows starting the provider in a standalone mode that's compatible with debuggers, and outputs information on how to drive the standalone provider with Terraform. ([#471](https://github.com/hashicorp/terraform-plugin-sdk/issues/471))

BREAKING CHANGES

* Removed the `helper/mutexkv`, `helper/pathorcontents`, `httpclient`, and `helper/hashcode` packages. These packages can be easily replicated in plugin code if necessary or the v1 versions can be used side-by-side ([#438](https://github.com/hashicorp/terraform-plugin-sdk/issues/438))
* Removed `schema.PanicOnErr/TF_SCHEMA_PANIC_ON_ERR` environment variable. `d.Set()` errors are now logged in production and panic during acceptance tests (`TF_ACC=1`). ([#462](https://github.com/hashicorp/terraform-plugin-sdk/issues/462))
* Running provider tests now requires Terraform 0.12.26 or higher. ([#471](https://github.com/hashicorp/terraform-plugin-sdk/issues/471))
* Removed `acctest` package, as it is no longer needed. The calls to `acctest.UseBinaryDriver` can be deleted; they're no longer necessary. ([#471](https://github.com/hashicorp/terraform-plugin-sdk/issues/471))
* The `resource.TestCase.Providers` and `resource.TestCaseProviderFactories` maps must now have exactly one entry set between both of them, meaning one or the other should be used. Only the provider under test should be present in these maps. Providers that tests rely upon can be used by setting provider blocks in the test case, where `terraform init` will pick them up automatically. ([#471](https://github.com/hashicorp/terraform-plugin-sdk/issues/471))
* The `TF_LOG_PATH_MASK` used to filter provider logs by test name when running tests has been removed. ([#473](https://github.com/hashicorp/terraform-plugin-sdk/issues/473))

ENHANCEMENTS

* Added a `schema.Provider.UserAgent` method to generate a User-Agent string ([#474](https://github.com/hashicorp/terraform-plugin-sdk/issues/474))
* Convenience methods were added to the `diag` package to simplify common error cases ([#449](https://github.com/hashicorp/terraform-plugin-sdk/issues/449))

BUG FIXES

* Restored `d.Partial` and noted the edgecase it covers and odd Terraform behavior. ([#472](https://github.com/hashicorp/terraform-plugin-sdk/issues/472))
* Provider log output now respects the `TF_LOG` and `TF_LOG_PATH` environment variables when running tests. ([#473](https://github.com/hashicorp/terraform-plugin-sdk/issues/473))

# 2.0.0-rc.1 (May 05, 2020)

BREAKING CHANGES

* The SDK no longer supports protocol 4 (Terraform 0.11 and below). Providers built on the SDK after v2 will need Terraform 0.12 to be used.
* The new, previously optional binary acceptance testing framework is now the default and only available mode for testing. Test code and provider code will no longer reside in the same process. Providers also will have their processes stopped and restarted multiple times during a test. This more accurately mirrors the behavior of providers in production.
* Updated type signatures for some functions to include context.Context support. These include helpers in the helper/customdiff package, the CustomizeDiffFunc type, and the StateUpgradeFunc type. ([#276](https://github.com/hashicorp/terraform-plugin-sdk/issues/276))
* The Partial and SetPartial methods on schema.ResourceData have been removed, as they were rarely necessary and poorly understood. ([#318](https://github.com/hashicorp/terraform-plugin-sdk/issues/318))
* The terraform.ResourceProvider interface has been removed. The *schema.Provider type should be used directly, instead. ([#316](https://github.com/hashicorp/terraform-plugin-sdk/issues/316))
* Deprecated helper/validation functions have been removed. ([#333](https://github.com/hashicorp/terraform-plugin-sdk/issues/333))
* PromoteSingle’s use is discouraged, and so it has been removed from helper/schema.Schema. ([#337](https://github.com/hashicorp/terraform-plugin-sdk/issues/337))
* schema.UnsafeSetFieldRaw’s use is discouraged, and so it has been removed. ([#339](https://github.com/hashicorp/terraform-plugin-sdk/issues/339))
* Calls to schema.ResourceData.Set that would return an error now panic by default. TF_SCHEMA_PANIC_ON_ERROR can be set to a falsey value to disable this behavior.
* schema.Resource.Refresh has been removed, as it is unused in protocol 5. ([#370](https://github.com/hashicorp/terraform-plugin-sdk/issues/370))
* The Removed field has been removed from helper/schema.Schema, which means providers can no longer specify error messages when a recently removed field is used. This functionality had a lot of bugs and corner cases that worked in unexpected ways, and so was removed. ([#414](https://github.com/hashicorp/terraform-plugin-sdk/issues/414))
* The helper/encryption package has been removed, following our [published guidance](https://www.terraform.io/docs/extend/best-practices/sensitive-state.html#don-39-t-encrypt-state). ([#436](https://github.com/hashicorp/terraform-plugin-sdk/issues/436))
* In scenarios where the Go testing package was used, the github.com/mitchellh/go-testing-interface package may be required instead. ([#406](https://github.com/hashicorp/terraform-plugin-sdk/issues/406))
* <details><summary>A number of exported variables, functions, types, and interfaces that were not meant to be part of the SDK’s interface have been removed. Most plugins should not notice they are gone.</summary>
  
  The removals include:
  * helper/acctest.RemoteTestPrecheck
  * helper/acctest.SkipRemoteTestsEnvVar
  * helper/resource.EnvLogPathMask
  * helper/resource.GRPCTestProvider
  * helper/resource.LogOutput
  * helper/resource.Map
  * helper/resource.TestEnvVar
  * helper/resource.TestProvider
  * helper/schema.MultiMapReader
  * helper/schema.Provider.Input
  * plugin.Client
  * plugin.ClientConfig
  * plugin.DefaultProtocolVersion
  * plugin.GRPCProvider
  * plugin.GRPCProviderPlugin
  * plugin.GRPCProvisioner
  * plugin.GRPCProvisionerPlugin
  * plugin.HandShake.ProtocolVersion
  * plugin.ResourceProvider
  * plugin.ResourceProviderApplyArgs
  * plugin.ResourceProviderApplyResponse
  * plugin.ResourceProviderConfigureResponse
  * plugin.ResourceProviderDiffArgs
  * plugin.ResourceProviderDiffResponse
  * plugin.ResourceProviderGetSchemaArgs
  * plugin.ResourceProviderGetSchemaResponse
  * plugin.ResourceProviderImportStateArgs
  * plugin.ResourceProviderImportStateResponse
  * plugin.ResourceProviderInputArgs
  * plugin.ResourceProviderInputResponse
  * plugin.ResourceProviderPlugin
  * plugin.ResourceProviderReadDataApplyArgs
  * plugin.ResourceProviderReadDataApplyResponse
  * plugin.ResourceProviderReadDataDiffArgs
  * plugin.ResourceProviderReadDataDiffResponse
  * plugin.ResourceProviderRefreshArgs
  * plugin.ResourceProviderRefreshResponse
  * plugin.ResourceProviderServer
  * plugin.ResourceProviderStopResponse
  * plugin.ResourceProviderValidateArgs
  * plugin.ResourceProviderValidateResourceArgs
  * plugin.ResourceProviderValidateResourceResponse
  * plugin.ResourceProviderValidateResponse
  * plugin.UIInput
  * plugin.UIInputInputResponse
  * plugin.UIInputServer
  * plugin.UIOutput
  * plugin.UIOutputServer
  * plugin.VersionedPlugins no longer has a "provisioner" key
  * resource.RunNewTest
  * schema.Backend
  * schema.FromContextBackendConfig
  * schema.SetProto5
  * terraform.ApplyGraphBuilder
  * terraform.AttachResourceConfigTransformer
  * terraform.AttachSchemaTransformer
  * terraform.AttachStateTransformer
  * terraform.BackendState.Config
  * terraform.BackendState.Empty
  * terraform.BackendState.ForPlan
  * terraform.BackendState.SetConfig
  * terraform.BasicGraphBuilder
  * terraform.BuiltinEvalContext
  * terraform.CallbackUIOutput
  * terraform.CBDEdgeTransformer
  * terraform.CheckCoreVersionRequirements
  * terraform.CloseProviderEvalTree
  * terraform.CloseProviderTransformer
  * terraform.CloseProvisionerTransformer
  * terraform.ConcreteProviderNodeFunc
  * terraform.ConcreteResourceInstanceDeposedNodeFunc
  * terraform.ConcreteResourceInstanceNodeFunc
  * terraform.ConcreteResourceNodeFunc
  * terraform.ConfigTransformer
  * terraform.ConfigTreeDependencies
  * terraform.ConnectionBlockSupersetSchema
  * terraform.Context
  * terraform.ContextGraphOpts
  * terraform.ContextGraphWalker
  * terraform.ContextMeta
  * terraform.ContextOpts
  * terraform.CountBoundaryTransformer
  * terraform.DefaultVariableValues
  * terraform.DestroyEdge
  * terraform.DestroyEdgeTransformer
  * terraform.DestroyOutputTransformer
  * terraform.DestroyPlanGraphBuilder
  * terraform.DestroyValueReferenceTransformer
  * terraform.Diff (this was eventually cut)
  * terraform.Diff.ModuleByPath
  * terraform.Diff.RootModule
  * terraform.DiffAttrInput
  * terraform.DiffAttrOutput
  * terraform.DiffAttrType
  * terraform.DiffAttrUnknown
  * terraform.DiffChangeType
  * terraform.DiffCreate
  * terraform.DiffDestroy
  * terraform.DiffDestroyCreate
  * terraform.DiffInvalid
  * terraform.DiffNone
  * terraform.DiffRefresh
  * terraform.DiffTransformer
  * terraform.DiffUpdate
  * terraform.EphemeralState.DeepCopy
  * terraform.ErrNoState
  * terraform.Eval
  * terraform.EvalApply
  * terraform.EvalApplyPost
  * terraform.EvalApplyPre
  * terraform.EvalApplyProvisioners
  * terraform.EvalCheckModuleRemoved
  * terraform.EvalCheckPlannedChange
  * terraform.EvalCheckPreventDestroy
  * terraform.EvalCloseProvider
  * terraform.EvalCloseProvisioner
  * terraform.EvalConfigBlock
  * terraform.EvalConfigExpr
  * terraform.EvalConfigProvider
  * terraform.EvalContext
  * terraform.EvalCountFixZeroOneBoundaryGlobal
  * terraform.EvalDataForInstanceKey
  * terraform.EvalDataForNoInstanceKey
  * terraform.EvalDeleteLocal
  * terraform.EvalDeleteOutput
  * terraform.EvalDeposeState
  * terraform.EvalDiff
  * terraform.EvalDiffDestroy
  * terraform.EvalEarlyExitError
  * terraform.EvalFilter
  * terraform.EvalForgetResourceState
  * terraform.EvalGetProvider
  * terraform.EvalGetProvisioner
  * terraform.EvalGraphBuilder
  * terraform.EvalIf
  * terraform.EvalImportState
  * terraform.EvalImportStateVerify
  * terraform.EvalInitProvider
  * terraform.EvalInitProvisioner
  * terraform.EvalLocal
  * terraform.EvalMaybeRestoreDeposedObject
  * terraform.EvalMaybeTainted
  * terraform.EvalModuleCallArgument
  * terraform.EvalNode
  * terraform.EvalNodeFilterable
  * terraform.EvalNodeFilterFunc
  * terraform.EvalNodeFilterOp
  * terraform.EvalNodeOpFilterable
  * terraform.EvalNoop
  * terraform.EvalOpFilter
  * terraform.EvalRaw
  * terraform.EvalReadData
  * terraform.EvalReadDataApply
  * terraform.EvalReadDiff
  * terraform.EvalReadState
  * terraform.EvalReadStateDeposed
  * terraform.EvalReduceDiff
  * terraform.EvalRefresh
  * terraform.EvalRequireState
  * terraform.EvalReturnError
  * terraform.EvalSequence
  * terraform.EvalSetModuleCallArguments
  * terraform.Evaluator
  * terraform.EvalUpdateStateHook
  * terraform.EvalValidateCount
  * terraform.EvalValidateProvider
  * terraform.EvalValidateProvisioner
  * terraform.EvalValidateResource
  * terraform.EvalValidateSelfRef
  * terraform.EvalWriteDiff
  * terraform.EvalWriteOutput
  * terraform.EvalWriteResourceState
  * terraform.EvalWriteState
  * terraform.EvalWriteStateDeposed
  * terraform.ExpandTransform
  * terraform.ForcedCBDTransformer
  * terraform.Graph
  * terraform.GraphBuilder
  * terraform.GraphDot
  * terraform.GraphNodeAttachDestroyer
  * terraform.GraphNodeAttachProvider
  * terraform.GraphNodeAttachProviderConfigSchema
  * terraform.GraphNodeAttachProvisionerSchema
  * terraform.GraphNodeAttachResourceConfig
  * terraform.GraphNodeAttachResourceSchema
  * terraform.GraphNodeAttachResourceState
  * terraform.GraphNodeCloseProvider
  * terraform.GraphNodeCloseProvisioner
  * terraform.GraphNodeCreator
  * terraform.GraphNodeDeposedResourceInstanceObject
  * terraform.GraphNodeDeposer
  * terraform.GraphNodeDestroyer
  * terraform.GraphNodeDestroyerCBD
  * terraform.GraphNodeDynamicExpandable
  * terraform.GraphNodeEvalable
  * terraform.GraphNodeExpandable
  * terraform.GraphNodeProvider
  * terraform.GraphNodeProviderConsumer
  * terraform.GraphNodeProvisioner
  * terraform.GraphNodeProvisionerConsumer
  * terraform.GraphNodeReferenceable
  * terraform.GraphNodeReferenceOutside
  * terraform.GraphNodeReferencer
  * terraform.GraphNodeResource
  * terraform.GraphNodeResourceInstance
  * terraform.GraphNodeSubgraph
  * terraform.GraphNodeSubPath
  * terraform.GraphNodeTargetable
  * terraform.GraphNodeTargetDownstream
  * terraform.GraphTransformer
  * terraform.GraphTransformIf
  * terraform.GraphTransformMulti
  * terraform.GraphType
  * terraform.GraphTypeApply
  * terraform.GraphTypeEval
  * terraform.GraphTypeInvalid
  * terraform.GraphTypeLegacy
  * terraform.GraphTypeMap
  * terraform.GraphTypePlan
  * terraform.GraphTypePlanDestroy
  * terraform.GraphTypeRefresh
  * terraform.GraphTypeValidate
  * terraform.GraphVertexTransformer
  * terraform.GraphWalker
  * terraform.Hook
  * terraform.HookAction
  * terraform.HookActionContinue
  * terraform.HookActionHalt
  * terraform.ImportGraphBuilder
  * terraform.ImportOpts
  * terraform.ImportProviderValidateTransformer
  * terraform.ImportStateTransformer
  * terraform.ImportTarget
  * terraform.InputMode
  * terraform.InputModeProvider
  * terraform.InputModeStd
  * terraform.InputModeVar
  * terraform.InputModeVarUnset
  * terraform.InputOpts
  * terraform.InputValue
  * terraform.InputValues
  * terraform.InputValuesFromCaller
  * terraform.InstanceDiff.Copy
  * terraform.InstanceDiff.DelAttribute
  * terraform.InstanceDiff.GetAttributesLen
  * terraform.InstanceDiff.SetAttribute
  * terraform.InstanceDiff.SetDestroy
  * terraform.InstanceDiff.SetDestroyDeposed
  * terraform.InstanceDiff.SetTainted
  * terraform.InstanceInfo.ResourceAddress
  * terraform.InstanceKeyEvalData
  * terraform.InstanceType
  * terraform.LoadSchemas
  * terraform.LocalTransformer
  * terraform.MissingProviderTransformer
  * terraform.MissingProvisionerTransformer
  * terraform.MockEvalContext
  * terraform.MockHook
  * terraform.MockProvider
  * terraform.MockProvisioner
  * terraform.MockResourceProvider (this was removed)
  * terraform.MockResourceProvider.Input
  * terraform.MockResourceProvider.InputCalled
  * terraform.MockResourceProvider.InputConfig
  * terraform.MockResourceProvider.InputFn
  * terraform.MockResourceProvider.InputInput
  * terraform.MockResourceProvider.InputReturnConfig
  * terraform.MockResourceProvider.InputReturnError
  * terraform.MockResourceProvisioner
  * terraform.MockUIInput
  * terraform.MockUIOutput
  * terraform.ModuleDiff (this was eventually cut)
  * terraform.ModuleDiff.IsRoot
  * terraform.ModuleState.Empty
  * terraform.ModuleState.IsDescendent
  * terraform.ModuleState.IsRoot
  * terraform.ModuleState.Orphans
  * terraform.ModuleState.RemovedOutputs
  * terraform.ModuleState.View
  * terraform.ModuleVariableTransformer
  * terraform.MustShimLegacyState
  * terraform.NewContext
  * terraform.NewInstanceInfo
  * terraform.NewLegacyResourceAddress
  * terraform.NewLegacyResourceInstanceAddress
  * terraform.NewNodeAbstractResource
  * terraform.NewNodeAbstractResourceInstance
  * terraform.NewReferenceMap
  * terraform.NewResource
  * terraform.NewSemaphore
  * terraform.NilHook
  * terraform.NodeAbstractProvider
  * terraform.NodeAbstractResource
  * terraform.NodeAbstractResourceInstance
  * terraform.NodeApplyableModuleVariable
  * terraform.NodeApplyableOutput
  * terraform.NodeApplyableProvider
  * terraform.NodeApplyableResource
  * terraform.NodeApplyableResourceInstance
  * terraform.NodeCountBoundary
  * terraform.NodeDestroyableDataResourceInstance
  * terraform.NodeDestroyableOutput
  * terraform.NodeDestroyDeposedResourceInstanceObject
  * terraform.NodeDestroyResource
  * terraform.NodeDestroyResourceInstance
  * terraform.NodeDisabledProvider
  * terraform.NodeEvalableProvider
  * terraform.NodeLocal
  * terraform.NodeModuleRemoved
  * terraform.NodeOutputOrphan
  * terraform.NodePlanDeposedResourceInstanceObject
  * terraform.NodePlanDestroyableResourceInstance
  * terraform.NodePlannableResource
  * terraform.NodePlannableResourceInstance
  * terraform.NodePlannableResourceInstanceOrphan
  * terraform.NodeProvisioner
  * terraform.NodeRefreshableDataResource
  * terraform.NodeRefreshableDataResourceInstance
  * terraform.NodeRefreshableManagedResource
  * terraform.NodeRefreshableManagedResourceInstance
  * terraform.NodeRootVariable
  * terraform.NodeValidatableResource
  * terraform.NullGraphWalker
  * terraform.OrphanOutputTransformer
  * terraform.OrphanResourceCountTransformer
  * terraform.OrphanResourceInstanceTransformer
  * terraform.OrphanResourceTransformer
  * terraform.OutputTransformer
  * terraform.ParentProviderTransformer
  * terraform.ParseInstanceType
  * terraform.ParseResourceAddress
  * terraform.ParseResourceAddressForInstanceDiff
  * terraform.ParseResourceIndex
  * terraform.ParseResourcePath
  * terraform.ParseResourceStateKey
  * terraform.PathObjectCacheKey
  * terraform.Plan
  * terraform.PlanGraphBuilder
  * terraform.PrefixUIInput
  * terraform.ProviderConfigTransformer
  * terraform.ProviderEvalTree
  * terraform.ProviderHasDataSource
  * terraform.ProviderHasResource
  * terraform.ProviderSchema.SchemaForResourceAddr
  * terraform.ProviderSchema.SchemaForResourceType
  * terraform.ProviderTransformer
  * terraform.ProvisionerFactory
  * terraform.ProvisionerTransformer
  * terraform.ProvisionerUIOutput
  * terraform.PruneProviderTransformer
  * terraform.PruneUnusedValuesTransformer
  * terraform.ReadPlan
  * terraform.ReadState
  * terraform.ReadStateV1
  * terraform.ReadStateV2
  * terraform.ReadStateV3
  * terraform.ReferenceMap
  * terraform.ReferencesFromConfig
  * terraform.ReferenceTransformer
  * terraform.RefreshGraphBuilder
  * terraform.RemoteState.Equals
  * terraform.RemovableIfNotTargeted
  * terraform.RemovedModuleTransformer
  * terraform.Resource
  * terraform.ResourceAddress
  * terraform.ResourceAttrDiff.Empty
  * terraform.ResourceConfig.CheckSet
  * terraform.ResourceConfig.IsSet
  * terraform.ResourceCountTransformer
  * terraform.ResourceFlag
  * terraform.ResourceProviderCloser (this was removed)
  * terraform.ResourceProviderFactoryFixed (this was removed)
  * terraform.ResourceProviderResolver
  * terraform.ResourceProviderResolverFixed
  * terraform.ResourceProviderResolverFunc
  * terraform.ResourceProvisioner
  * terraform.ResourceProvisionerCloser
  * terraform.ResourceProvisionerFactory
  * terraform.RootTransformer
  * terraform.RootVariableTransformer
  * terraform.Schemas
  * terraform.Semaphore
  * terraform.ShimLegacyState
  * terraform.State.FromFutureTerraform
  * terraform.State.MarshalEqual
  * terraform.StateFilter
  * terraform.StateFilterResult
  * terraform.StateFilterResultSlice
  * terraform.StateTransformer
  * terraform.StateVersion
  * terraform.TargetsTransformer
  * terraform.TestStateFile
  * terraform.TransformProviders
  * terraform.TransitiveReductionTransformer
  * terraform.TypeDeposed
  * terraform.TypeInvalid
  * terraform.TypePrimary
  * terraform.TypeTainted
  * terraform.UIInput
  * terraform.UIOutput
  * terraform.UpgradeResourceState
  * terraform.ValidateGraphBuilder
  * terraform.ValueFromAutoFile
  * terraform.ValueFromCaller
  * terraform.ValueFromCLIArg
  * terraform.ValueFromConfig
  * terraform.ValueFromEnvVar
  * terraform.ValueFromInput
  * terraform.ValueFromNamedFile
  * terraform.ValueFromPlan
  * terraform.ValueFromUnknown
  * terraform.ValueSourceType
  * terraform.VertexTransformer
  * terraform.WritePlan
  * terraform.WriteState
  </details>

FEATURES
* Many functions in the SDK now have support for context.Context, including CreateContextFunc, ReadContextFunc, UpdateContextFunc, and DeleteContextFunc, analogs to the existing CreateFunc, ReadFunc, UpdateFunc, and DeleteFuncs. This offers more accurate cancellation and timeouts. ([#276](https://github.com/hashicorp/terraform-plugin-sdk/issues/276))
* Many functions in the SDK now return a new diag.Diagnostics type, like the new CreateContextFunc, ReadContextFunc, UpdateContextFunc, DeleteContextFunc, and a new ValidateDiagFunc. When using these Diagnostics, Terraform will now indicate more precisely-scoped errors, and providers now have the ability to display warnings.
* A new feature, provider metadata, is shipping as part of Terraform 0.13. This feature allows module authors to give information to providers without the information being persisted to state, which is useful for indicating metadata about modules. This is experimental new functionality and its usage should be closely coordinated with the Terraform core team to ensure that limitations are understood. [See the PR in Terraform core for more information.](https://github.com/hashicorp/terraform/pull/22583) ([#405](https://github.com/hashicorp/terraform-plugin-sdk/issues/405))

DEPRECATIONS
* The ExistsFunc defined on a schema.Resource is now deprecated. This logic can be achieved in the ReadFunc for that schema.Resource instead, and often was duplicated unnecessarily.
* Functions that got context- or diagnostics-aware counterparts--like CreateFunc, ReadFunc, UpdateFunc, DeleteFunc, and ValidateFunc--are now deprecated in favor of their context- and/or diagnostics-aware counterparts.

ENHANCEMENTS
* A number of new map validators that take advantage of the Diagnostics support have been added. ([#304](https://github.com/hashicorp/terraform-plugin-sdk/issues/304))
* schema.Resource and schema.Schema now have optional Description fields, which will surface information for user-facing interfaces for the provider. These fields can hold plain text or markdown, depending on the global DescriptionKind setting. ([#349](https://github.com/hashicorp/terraform-plugin-sdk/issues/349))

BUG FIXES
* helper/acctest.RandIntRange will now correctly return an integer between min and max; previously, it would return an integer between 0 and max-min. ([#300](https://github.com/hashicorp/terraform-plugin-sdk/issues/300))
* NonRetryableError and RetryableError will now throw an explicit error if they’re given a nil error. Before unspecified and confusing behavior would arise. ([#199](https://github.com/hashicorp/terraform-plugin-sdk/issues/199))
* TypeSet hash values are no longer collapsed into a single value when they consist only of Computed attributes. ([#197](https://github.com/hashicorp/terraform-plugin-sdk/issues/197))
* Computed attributes now have stronger validation around what properties can be set on their schema.Schema. ([#336](https://github.com/hashicorp/terraform-plugin-sdk/issues/336))
* Using a schema.Resource as the Elem on a TypeMap now returns an error; previously, unspecified and confusing behavior was exhibited. ([#338](https://github.com/hashicorp/terraform-plugin-sdk/issues/338))
* Using TestCheckResourceAttrPair to compare the same attribute on the same resource will now throw an error. ([#335](https://github.com/hashicorp/terraform-plugin-sdk/issues/335))
* Test sweeping will now error if a dependency sweeper is specified but doesn’t exist. ([#398](https://github.com/hashicorp/terraform-plugin-sdk/issues/398))

---

For information on v1.x releases, please see [the v1 branch changelog](https://github.com/hashicorp/terraform-plugin-sdk/blob/v1-maint/CHANGELOG.md).
