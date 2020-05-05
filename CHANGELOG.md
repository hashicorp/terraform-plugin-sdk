# 2.0.0-rc.1 (Unreleased)

BREAKING CHANGES

* The SDK no longer supports protocol 4 (Terraform 0.11 and below). Providers built on the SDK after v2 will need Terraform 0.12 to be used.
* The new, previously optional binary acceptance testing framework is now the default and only available mode for testing. Test code and provider code will no longer reside in the same process. Providers also will have their processes stopped and restarted multiple times during a test. This more accurately mirrors the behavior of providers in production.
* Updated type signatures for some functions to include context.Context support. These include helpers in the helper/customdiff package, the CustomizeDiffFunc type, and the StateUpgradeFunc type. [GH-276]
* The Partial and SetPartial methods on schema.ResourceData have been removed, as they were rarely necessary and poorly understood. [GH-318]
* The terraform.ResourceProvider interface has been removed. The *schema.Provider type should be used directly, instead. [GH-316]
* Deprecated helper/validation functions have been removed. [GH-333]
* PromoteSingle’s use is discouraged, and so it has been removed from helper/schema.Schema. [GH-337]
* schema.UnsafeSetFieldRaw’s use is discouraged, and so it has been removed. [GH-339]
* Calls to schema.ResourceData.Set that would return an error now panic by default. TF_SCHEMA_PANIC_ON_ERROR can be set to a falsey value to disable this behavior.
* schema.Resource.Refresh has been removed, as it is unused in protocol 5. [GH-370]
* The Removed field has been removed from helper/schema.Schema, which means providers can no longer specify error messages when a recently removed field is used. This functionality had a lot of bugs and corner cases that worked in unexpected ways, and so was removed. [GH-414]
* The helper/encryption package has been removed, following our [published guidance](https://www.terraform.io/docs/extend/best-practices/sensitive-state.html#don-39-t-encrypt-state). [GH-436]
* In scenarios where the Go testing package was used, the github.com/mitchellh/go-testing-interface package may be required instead. [GH-406]
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
* Many functions in the SDK now have support for context.Context, including CreateContextFunc, ReadContextFunc, UpdateContextFunc, and DeleteContextFunc, analogs to the existing CreateFunc, ReadFunc, UpdateFunc, and DeleteFuncs. This offers more accurate cancellation and timeouts. [GH-276]
* Many functions in the SDK now return a new diag.Diagnostics type, like the new CreateContextFunc, ReadContextFunc, UpdateContextFunc, DeleteContextFunc, and a new ValidateDiagFunc. When using these Diagnostics, Terraform will now indicate more precisely-scoped errors, and providers now have the ability to display warnings.
* A new feature, provider metadata, is shipping as part of Terraform 0.13. This feature allows module authors to give information to providers without the information being persisted to state, which is useful for indicating metadata about modules. This is experimental new functionality and its usage should be closely coordinated with the Terraform core team to ensure that limitations are understood. [See the PR in Terraform core for more information.](https://github.com/hashicorp/terraform/pull/22583) [GH-405]

DEPRECATIONS
* The ExistsFunc defined on a schema.Resource is now deprecated. This logic can be achieved in the ReadFunc for that schema.Resource instead, and often was duplicated unnecessarily.
* Functions that got context- or diagnostics-aware counterparts--like CreateFunc, ReadFunc, UpdateFunc, DeleteFunc, and ValidateFunc--are now deprecated in favor of their context- and/or diagnostics-aware counterparts.

ENHANCEMENTS
* A number of new map validators that take advantage of the Diagnostics support have been added. [GH-304]
* schema.Resource and schema.Schema now have optional Description fields, which will surface information for user-facing interfaces for the provider. These fields can hold plain text or markdown, depending on the global DescriptionKind setting. [GH-349]

BUG FIXES
* helper/acctest.RandIntRange will now correctly return an integer between min and max; previously, it would return an integer between 0 and max-min. [GH-300]
* NonRetryableError and RetryableError will now throw an explicit error if they’re given a nil error. Before unspecified and confusing behavior would arise. [GH-199]
* TypeSet hash values are no longer collapsed into a single value when they consist only of Computed attributes. [GH-197]
* Computed attributes now have stronger validation around what properties can be set on their schema.Schema. [GH-336]
* Using a schema.Resource as the Elem on a TypeMap now returns an error; previously, unspecified and confusing behavior was exhibited. [GH-338]
* Using TestCheckResourceAttrPair to compare the same attribute on the same resource will now throw an error. [GH-335]
* Test sweeping will now error if a dependency sweeper is specified but doesn’t exist. [GH-398]

---

For information on v1.x releases, please see [the v1 branch changelog](https://github.com/hashicorp/terraform-plugin-sdk/blob/v1-maint/CHANGELOG.md).
