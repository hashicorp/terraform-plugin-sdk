# Q1 2021 Roadmap

Each quarter, the team will highlight areas of focus for our work and upcoming
research.

Each release will include necessary tasks that lead to the completion of the
stated goals as well as community pull requests, enhancements, and features
that are not highlighted in the roadmap. This upcoming calendar quarter
(Jan-Mar 2021) we will be prioritizing the following areas of work:

## Currently In Progress

### v2.x Release

Following the release of v2.0.0 we will continue to make improvements to the
SDK to ease the upgrade path for provider developers who have not yet migrated
and improve the functionality for those that have.

### terraform-plugin-go / terraform-plugin-mux improvements

Following the release of the
[terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go) and
[terraform-plugin-mux](https://github.com/hashicorp/terraform-plugin-mux) modules, we
will work to improve and refine these modules to make them more robust, easier
to use, and more approachable for developers. This includes creating [new
abstractions](https://github.com/hashicorp/terraform-plugin-go-contrib) for
terraform-plugin-go that developers may find useful and will help inform our
future design efforts.

## Under Research

### An HCL2-native framework

With the release of terraform-plugin-go, we're researching, prototyping, and
designing a framework based on the same ideas and foundations, meant to mirror
terraform-plugin-sdk in approachability and ease of use while retaining
terraform-plugin-go's level of suitability for modern Terraform development.

### A modernized testing framework

Following the creation of our reattach-based binary test driver in v2.0.0 of
the SDK, we're investigating what a redesigned approach to testing Terraform
providers may look like and if we can aid developers in making and testing
clearer assertions about their providers.

### More observable provider development

We're currently researching and investigating ways in which we can surface more
information to provider developers about what is happening in their providers,
and how we can make the information we surface more accessible and easier to
digest.
