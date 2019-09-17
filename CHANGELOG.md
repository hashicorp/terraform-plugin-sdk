# 1.0.0 (September 17, 2019)

This SDK is functionally equivalent to the "legacy" SDK in `hashicorp/terraform` [`v0.12.9`](https://github.com/hashicorp/terraform/blob/v0.12.9/CHANGELOG.md).

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/plugin-sdk.html) of the website.

FEATURES:

 * Add `meta` package which exposes the version of the SDK, replacing the `version` package which previously exposed the Terraform version ([#37](https://github.com/hashicorp/terraform-plugin-sdk/issues/37)] [[#24](https://github.com/hashicorp/terraform-plugin-sdk/issues/24))
