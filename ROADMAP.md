# Q3 2021 Roadmap

Each quarter, the team will highlight areas of focus for our work and upcoming
research.

Each release will include necessary tasks that lead to the completion of the
stated goals as well as community pull requests, enhancements, and features
that are not highlighted in the roadmap. This upcoming calendar quarter
(Aug-Oct 2021) we will be prioritizing the following areas of work:

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

### An HCL2-native framework

Following the release of
[terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework)
we will work to add more features to the framework, aiming for feature parity with
SDKv2, and work to make the framework more approachable, easier to use, and better
documented.

### More observable provider development

With the release of
[terraform-plugin-log](https://github.com/hashicorp/terraform-plugin-log) we will work
to integrate this tooling into terraform-plugin-go, terraform-plugin-framework, and v2
of the SDK to help enable more powerful debugging of Terraform providers.

## Under Research

### A modernized testing framework

Following the creation of our reattach-based binary test driver in v2.0.0 of
the SDK, we're investigating what a redesigned approach to testing Terraform
providers may look like and if we can aid developers in making and testing
clearer assertions about their providers.
