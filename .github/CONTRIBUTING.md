# Contributing to Terraform Plugin SDK

**First:** if you're unsure or afraid of _anything_, just ask
or submit the issue describing the problem you're aiming to solve.

Any bug fix and feature has to be considered in the context
of many (100+) providers and the wider Terraform ecosystem.
This is great as your contribution can have a big positive impact,
but we have to assess potential negative impact too (e.g. breaking
existing providers which may not use a new feature).

To provide some safety to the wider provider ecosystem, we strictly follow
[semantic versioning](https://semver.org/) and any changes that could be
considered as breaking will only be released as part of major release.

## Table of Contents

 - [I just have a question](#i-just-have-a-question)
 - [I want to report a vulnerability](#i-want-to-report-a-vulnerability)
 - [Scope (Core vs SDK vs Providers)](#scope-core-vs-sdk-vs-providers)
 - [New Issue](#new-issue)
 - [New Pull Request](#new-pull-request)

## I just have a question

> **Note:** We use GitHub for tracking bugs and feature requests related to Plugin SDK.

For questions, please see relevant channels at https://www.terraform.io/community.html

## I want to report a vulnerability

Please disclose security vulnerabilities responsibly by following the procedure
described at https://www.hashicorp.com/security#vulnerability-reporting

## Scope (Core vs SDK vs Providers)

While Terraform acts as a single program from the user's perspective
it is made up of a few parts, each of which have different role and repository.

This section describes the scope of notable repositories, which may help you
ensure you're in the right place when reporting bugs and feature requests,
or submitting a patch.

 - `hashicorp/terraform` - Terraform **Core** which implements all the low-level functionality which isn't domain specific (that's covered by providers). Read more about [distinction between core & providers in the Readme](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/README.md#scope-providers-vs-core).
 - `terraform-providers/*` - This organization contains all official Terraform **Providers** built on top of the Plugin SDK
 - `hashicorp/terraform-plugin-sdk` - Terraform **Plugin SDK** used to build Providers
 - `hashicorp/terraform-website` - Source code of **documentation** published on [terraform.io](https://www.terraform.io), including [Extend section](https://www.terraform.io/docs/extend/index.html) which has source in [the `extend` folder](https://github.com/hashicorp/terraform-website/tree/main/content/source/docs/extend).
 - `hashicorp/hcl2` - **HCL** (HashiCorp Config Language) is the language used by users of Terraform (Core) to describe infrastructure. The parser and other features concerning the language (such as builtin functions) are found here.
 - `zclconf/go-cty` - **cty**, the type system used by both Terraform (Core) and SDK (therefore providers too) to represent data in state before and after gRPC encoding/decoding

## New Issue

We welcome issues of all kinds including feature requests, bug reports or documentation suggestions. Below are guidelines for well-formed issues of each type.

### Bug Reports

 - [ ] **Test against latest release**: Make sure you test against the latest avaiable version of both Terraform and SDK.
It is possible we already fixed the bug you're experiencing.

 - [ ] **Search for duplicates**: It's helpful to keep bug reports consolidated to one thread, so do a quick search on existing bug reports to check if anybody else has reported the same thing. You can scope searches by the label `bug` to help narrow things down.

 - [ ] **Include steps to reproduce**: Provide steps to reproduce the issue, along with code examples (both HCL and Go, where applicable) and/or real code, so we can try to reproduce it. Without this, it makes it much harder (sometimes impossible) to fix the issue.

### Feature Requests

 - [ ] **Search for possible duplicate requests**: It's helpful to keep requests consolidated to one thread, so do a quick search on existing requests to check if anybody else has reported the same thing. You can scope searches by the label `enhancement` to help narrow things down.

 - [ ] **Include a use case description**: In addition to describing the behavior of the feature you'd like to see added, it's helpful to also lay out the reason why the feature would be important and how it would benefit the wider Terraform ecosystem. Use case in context of 1 provider is good, wider context of more providers is better.

## New Pull Request

Thank you for contributing!

We are happy to review pull requests without associated issues,
but we highly recommend starting by describing and discussing
your problem or feature and attaching use cases to an issue first
before raising a pull request.

- [ ] **Early validation of idea and implementation plan**: Terraform's SDK is complicated enough that there are often several ways to implement something, each of which has different implications and tradeoffs. Working through a plan of attack with the team before you dive into implementation will help ensure that you're working in the right direction.

- [ ] **Unit Tests**: It may go without saying, but every new patch should be covered by tests wherever possible.

- [ ] **Provider testing**: The SDK's Test Framework is still undergoing active development and may not catch all corner cases when patches are tested outside of real provider code. It is therefore extremely valuable if you can run acceptance tests of at least one provider which takes advantage of your bug fix or uses new feature and demonstrate that your patch doesn't break other provider(s) relying on the existing SDK.

- [ ] **Go Modules**: We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage and version all our dependencies. Please make sure that you reflect dependency changes in your pull requests appropriately (e.g. `go get`, `go mod tidy` or other commands). Refer to the [dependency updates](#dependency-updates) section for more information about how this project maintains existing dependencies.

- [ ] **Changelog**: Refer to the [changelog](#changelog) section for more information about how to create changelog entries.

### Cosmetic changes, code formatting, and typos

In general we do not accept PRs containing only the following changes:

 - Correcting spelling or typos
 - Code formatting, including whitespace
 - Other cosmetic changes that do not affect functionality
 
While we appreciate the effort that goes into preparing PRs, there is always a tradeoff between benefit and cost. The costs involved in accepting such contributions include the time taken for thorough review, the noise created in the git history, and the increased number of GitHub notifications that maintainers must attend to.

In the case of `terraform-plugin-sdk`, the repo's close relationship to the `terraform` repo means that maintainers will sometimes port changes from `terraform` to `terraform-plugin-sdk`. Cosmetic changes to the SDK repo make this much more time-consuming as they will cause merge conflicts. This is the major hidden cost of cosmetic PRs, and the main reason we do not accept them at this time.

#### Exceptions

We belive that one should "leave the campsite cleaner than you found it", so you are welcome to clean up cosmetic issues in the neighbourhood when submitting a patch that makes functional changes or fixes.

### Dependency Updates

Dependency management is performed by [dependabot](https://docs.github.com/en/code-security/supply-chain-security/keeping-your-dependencies-updated-automatically/about-dependabot-version-updates). Where possible, dependency updates should occur through that system to ensure all Go module files are appropriately updated and to prevent duplicated effort of concurrent update submissions. Once available, updates are expected to be verified and merged to prevent latent technical debt.

### Changelog

HashiCorp’s open-source projects have always maintained user-friendly, readable CHANGELOGs that allow practitioners and developers to tell at a glance whether a release should have any effect on them, and to gauge the risk of an upgrade. We use [go-changelog](https://github.com/hashicorp/go-changelog) to generate and update the changelog from files created in the `.changelog/` directory.

#### Changelog Format

The changelog format requires an entry in the following format, where HEADER corresponds to the changelog category, and the entry is the changelog entry itself. The entry should be included in a file in the `.changelog` directory with the naming convention `{PR-NUMBER}.txt`. For example, to create a changelog entry for pull request 1234, there should be a file named `.changelog/1234.txt`.

``````markdown
```release-note:{HEADER}
{ENTRY}
```
``````

If a pull request should contain multiple changelog entries, then multiple blocks can be added to the same changelog file. For example:

``````markdown
```release-note:note
tfsdk: The `Old()` function has been deprecated. Any code using `Old()` should be updated to use the new `New()` function instead.
```

```release-note:enhancement
tfsdk: Added `New()` function, which does new and exciting things
```
``````

#### Pull Request Types to CHANGELOG

The CHANGELOG is intended to show developer-impacting changes to the codebase for a particular version. If every change or commit to the code resulted in an entry, the CHANGELOG would become less useful for developers. The lists below are general guidelines and examples for when a decision needs to be made to decide whether a change should have an entry.

##### Changes that should not have a CHANGELOG entry

- Documentation updates
- Testing updates
- Code refactoring

##### Changes that may have a CHANGELOG entry

- Dependency updates: If the update contains relevant bug fixes or enhancements that affect developers, those should be called out.

##### Changes that should have a CHANGELOG entry

###### Major Features

A major feature entry should use the `release-note:feature` header.

``````markdown
```release-note:feature
Added `great` package, which solves all the problems
```
``````

###### Bug Fixes

A new bug entry should use the `release-note:bug` header and have a prefix indicating the sub-package it corresponds to, a colon, then followed by a brief summary. Use an `all` prefix should the fix apply to all sub-packages.

``````markdown
```release-note:bug
tfsdk: Prevented potential panic in `Example()` function
```
``````

###### Enhancements

A new enhancement entry should use the `release-note:enhancement` header and have a prefix indicating the sub-package it corresponds to, a colon, then followed by a brief summary. Use an `all` prefix for enchancements that apply to all sub-packages.

``````markdown
```release-note:enhancement
attr: Added `Great` interface for doing great things
```
``````

###### Deprecations

A deprecation entry should use the `release-note:note` header and have a prefix indicating the sub-package it corresponds to, a colon, then followed by a brief summary. Use an `all` prefix for changes that apply to all sub-packages.

``````markdown
```release-note:note
diag: The `Old()` function is being deprecated in favor of the `New()` function
```
``````

###### Breaking Changes and Removals

A breaking-change entry should use the `release-note:breaking-change` header and have a prefix indicating the sub-package it corresponds to, a colon, then followed by a brief summary. Use an `all` prefix for changes that apply to all sub-packages.

``````markdown
```release-note:breaking-change
tfsdk: The `Example` type `Old` field has been removed since it is not necessary
```
``````

## Linting

GitHub Actions workflow bug and style checking is performed via [`actionlint`](https://github.com/rhysd/actionlint).

To run the GitHub Actions linters locally, install the `actionlint` tool, and run:

```shell
actionlint
```

Go code bug and style checking is performed via [`golangci-lint`](https://golangci-lint.run/).

To run the Go linters locally, install the `golangci-lint` tool, and run:

```shell
golangci-lint run ./...
```

## Testing

Code contributions should be supported by both unit and integration tests wherever possible. 

### GitHub Actions Tests

GitHub Actions workflow testing is performed via [`act`](https://github.com/nektos/act).

To run the GitHub Actions testing locally (setting appropriate event):

```shell
act --artifact-server-path /tmp --env ACTIONS_RUNTIME_TOKEN=test -P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest pull_request
```

The command options can be added to a `~/.actrc` file:

```text
--artifact-server-path /tmp
--env ACTIONS_RUNTIME_TOKEN=test
-P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest
```

So they do not need to be specified every invocation:

```shell
act pull_request
```

To test the `ci-go/terraform-provider-corner` job, a valid GitHub Personal Access Token (PAT) with public read permissions is required. It can be passed in via the `-s GITHUB_TOKEN=...` command option.

### Go Unit Tests

Go code unit testing is perfomed via Go's built-in testing functionality.

To run the Go unit testing locally:

```shell
go test ./...
```

This codebase follows Go conventions for unit testing. Some guidelines include:

- **File Naming**: Test files should be named `*_test.go` and usually reside in the same package as the code being tested.
- **Test Naming**: Test functions must include the `Test` prefix and should be named after the function or method under test. An `Example()` function test should be named `TestExample`. A `Data` type `Example()` method test should be named `TestDataExample`.
- **Concurrency**: Where possible, unit tests should be able to run concurrently and include a call to [`t.Parallel()`](https://pkg.go.dev/testing#T.Parallel). Usage of mutable shared data, such as environment variables or global variables that are used with reads and writes, is strongly discouraged.
- **Table Driven**: Where possible, unit tests should be written using the [table driven testing](https://github.com/golang/go/wiki/TableDrivenTests) style.
- **go-cmp**: Where possible, comparison testing should be done via [`go-cmp`](https://pkg.go.dev/github.com/google/go-cmp). In particular, the [`cmp.Diff()`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Diff) and [`cmp.Equal()`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Equal) functions are helpful.

A common template for implementing unit tests is:

```go
func TestExample(t *testing.T) {
    t.Parallel()

    testCases := map[string]struct{
        // fields to store inputs and expectations
    }{
        "test-description": {
            // fields from above
        },
    }

    for name, testCase := range testCases {
        // Do not omit this next line
        name, testCase := name, testCase

        t.Run(name, func(t *testing.T) {
            t.Parallel()

            // Implement test referencing testCase fields
        })
    }
}
```

### Integration tests

We use a special "corner case" Terraform provider for integration testing of terraform-plugin-sdk, called [terraform-provider-corner](https://github.com/hashicorp/terraform-provider-corner).

Integration testing for terraform-plugin-sdk involves compiling this provider against the version of the framework to be tested, and running the provider's acceptance tests. The `ci-go/terraform-provider-corner` CI job does this automatically for each PR commit and each commit to `main`. This ensures that changes to terraform-plugin-sdk do not cause regressions.

#### Creating a test case in terraform-provider-corner

The terraform-provider-corner repo contains several provider servers (which are combined in order to test [terraform-plugin-mux](https://github.com/hashicorp/terraform-plugin-mux)) to test different versions of the Terraform Plugin SDK and Framework.

To add a test case for terraform-plugin-sdk, add or modify resource code as appropriate in the [`sdkv2provider`](https://github.com/hashicorp/terraform-provider-corner/tree/main/internal/sdkv2provider). Then, create an acceptance test for the desired behaviour.

Creating a test case in terraform-provider-corner is a very helpful way to illustrate your bug report or feature request with easily reproducible provider code. We welcome PRs to terraform-provider-corner that demonstrate bugs and edge cases.

#### Adding integration tests to your terraform-plugin-sdk PR

When fixing a bug or adding a new feature to the framework, it is helpful to create a test case in real provider code. Since the test will fail until your change is included in a terraform-plugin-sdk release used by terraform-provider-corner, we recommend doing the following:

0. Fork and clone the terraform-plugin-sdk and terraform-provider-corner repositories to your local machine. Identify the bug you want to fix or the feature you want to add to terraform-plugin-sdk.
1. On your local fork of terraform-provider-corner, create a failing acceptance test demonstrating this behaviour. The test should be named `testAcc*`.
2. Add a `replace` directive to the `go.mod` file in your local terraform-provider-corner, pointing to your local fork of terraform-plugin-sdk.
3. Make the desired code change on your local fork of terraform-plugin-sdk. Don't forget unit tests as well!
4. Verify that the acceptance test now passes.
5. Make a PR to terraform-plugin-sdk proposing the code change.
6. Make a PR to terraform-provider-corner adding the new acceptance test and noting that it depends on the PR to terraform-plugin-sdk.

Maintainers will ensure that the acceptance test is merged into terraform-provider-corner once the terraform-plugin-sdk change is merged and released.

## Maintainers Guide

This section is dedicated to the maintainers of this project.

### Releases

Before running a release:

- **`meta/meta.go`**: The versions must be appropriately updated.
- **Changelog**: The changelog must be constructed from unreleased entries in the `.changelog` directory.

Install the latest version of the [`changelog-build`](https://pkg.go.dev/github.com/hashicorp/go-changelog/cmd/changelog-build) command, if it not already available:

```shell
go install github.com/hashicorp/go-changelog/cmd/changelog-build
```

Run the [`changelog-build`](https://pkg.go.dev/github.com/hashicorp/go-changelog/cmd/changelog-build) command from the root directory of the repository:

```shell
changelog-build -changelog-template .changelog.tmpl -entries-dir .changelog -last-release $(git describe --tags --abbrev=0) -note-template .changelog-note.tmpl -this-release HEAD
```

This will generate a section of Markdown text for the next release. Open the `CHANGELOG.md` file, add a `# X.Y.Z` header as the first line, then add the output from the `changelog-build` command.

Commit, push, create a release Git tag, and push the tag:

```shell
git add CHANGELOG.md meta/meta.go
git commit -m "Update CHANGELOG and versions for v1.2.3"
git push
git tag v1.2.3
git push --tags
```

GitHub Actions will pick up the new release tag and kick off the release workflow.
