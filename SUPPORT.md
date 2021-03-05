# Support Policy

Version 1 of the Terraform Plugin SDK is considered **deprecated**, which means
the following support policy is in place:

- Critical bug fixes will be accepted and merged until July 31, 2021, at which
point this version is considered End Of Life and no longer supported. New
features and non-critical bug fixes will not be accepted.
- We will not break the public interface exposed by the SDK in a minor or patch
release unless a critical security issue demands it, and only then as a last
resort.
- We do not proactively test version 1 of the SDK against new releases of
Terraform to ensure compatibility, but will address critical compatibility
issues that are reported until July 31, 2021.
- We do not backport bug fixes to prior minor releases. Only the latest minor
release receives new patch versions.

Please see the [upgrade
guide](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html) for
information on how to upgrade to version 2.
