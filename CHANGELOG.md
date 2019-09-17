# 1.0.0 (Unreleased)

This SDK is functionally equivalent to the "legacy" SDK in `hashicorp/terraform` [`v0.12.9`](https://github.com/hashicorp/terraform/blob/v0.12.9/CHANGELOG.md).

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/plugin-sdk.html) of the website.

FEATURES:

 * Add `meta` package which exposes the version of the SDK, replacing the `version` package which previously exposed the Terraform version [GH-37] [GH-24]
