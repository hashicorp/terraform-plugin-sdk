# Q3 2020 Roadmap

Each quarter, the team will highlight areas of focus for our work and upcoming research.
Each release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. This upcoming calendar quarter (Jul-Sep 2020) we will be prioritizing the following areas of work:

## Currently In Progress
### v2.x Release
With the release of v2.0.0 we will continue to make improvements to the SDK to ease the upgrade path for provider developers who have not yet migrated and improve the functionality for those that have.
 
## Planned
### Protocol-native provider development packages
As part of a project to enhance the Terraform Provider developer experience, we will be releasing packages to interact with the Terraform protocol directly, with little to no abstraction. Future packages will build on these to offer abstractions and frameworks similar to the current SDK’s experience. Providers may wish to make use of these packages for more complex resources that need advanced behavior not supported natively in our current SDK.
 
### Terraform Provider Resource Mux (terraform-plugin-mux)
To pave the way for new opt-in behaviors and SDK improvements, we will be releasing a terraform-plugin-mux package that lets provider developers utilize multiple SDKs in the same provider on a per resource basis.
 
### terraform-exec Module

**Repository:** https://github.com/hashicorp/terraform-exec

Our goal is to introduce a new Go module that enables programmatic interaction with Terraform. This approach offers provider developers a more reliable and sustainable way to interact with Terraform and this library will support running Terraform CLI commands from Go binaries, obtaining state and schema output as terraform-json types. 
