---
name: Feature request
about: Suggest a new feature or other enhancement.
labels: enhancement
---

### SDK version
<!--
Inspect your go.mod as below to find the version, and paste the result between the ``` marks below.

go list -m github.com/hashicorp/terraform-plugin-sdk/...

If you are not running the latest version of the SDK, please try upgrading
because your feature may have already been implemented.

If the command above doesn't yield any results, it means you may either be using v1 of the SDK or
have not have migrated to the standalone SDK yet. See https://www.terraform.io/docs/extend/plugin-sdk.html
for more.
-->
```
...
```

### Use-cases
<!---
In order to properly evaluate a feature request, it is necessary to understand the use-cases for it.
Please describe below the _end goal_ you are trying to achieve that has led you to request this feature.
Please keep this section focused on the problem and not on the suggested solution. We'll get to that in a moment, below!
-->

### Attempted Solutions
<!---
If you've already tried to solve the problem within SDK's existing features and found a limitation that prevented you from succeeding, please describe it below in as much detail as possible.

Ideally, this would include real HCL configuration that you tried, real Terraform command lines you ran, relevant snippet of code from your provider codebase and what results you got in each case.

Please remove any sensitive information such as passwords before sharing configuration snippets and command lines.
--->

### Proposal
<!---
If you have an idea for a way to address the problem via a change to SDK features, please describe it below.

In this section, it's helpful to include specific examples of how what you are suggesting might look in configuration files, or on the command line, since that allows us to understand the full picture of what you are proposing.

If you're not sure of some details, don't worry! When we evaluate the feature request we may suggest modifications as necessary to work within the design constraints of the SDK and Terraform Core.
-->

### References
<!--
Are there any other GitHub issues, whether open or closed, that are related to the problem you've described above or to the suggested solution? If so, please create a list below that mentions each of them. For example:

- #6017
-->
