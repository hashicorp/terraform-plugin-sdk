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

- [ ] **Go Modules**: We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage and version all our dependencies. Please make sure that you reflect dependency changes in your pull requests appropriately (e.g. `go get`, `go mod tidy` or other commands). Where possible it is better to raise a separate pull request with just dependency changes as it's easier to review such PR(s).

### Cosmetic changes, code formatting, and typos

In general we do not accept PRs containing only the following changes:

 - Correcting spelling or typos
 - Code formatting, including whitespace
 - Other cosmetic changes that do not affect functionality
 
While we appreciate the effort that goes into preparing PRs, there is always a tradeoff between benefit and cost. The costs involved in accepting such contributions include the time taken for thorough review, the noise created in the git history, and the increased number of GitHub notifications that maintainers must attend to.

In the case of `terraform-plugin-sdk`, the repo's close relationship to the `terraform` repo means that maintainers will sometimes port changes from `terraform` to `terraform-plugin-sdk`. Cosmetic changes to the SDK repo make this much more time-consuming as they will cause merge conflicts. This is the major hidden cost of cosmetic PRs, and the main reason we do not accept them at this time.

#### Exceptions

We belive that one should "leave the campsite cleaner than you found it", so you are welcome to clean up cosmetic issues in the neighbourhood when submitting a patch that makes functional changes or fixes.
