[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-sdk/v2)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2)

# Terraform Plugin SDK

This SDK enables building Terraform plugin which allows Terraform's users to manage existing and popular service providers as well as custom in-house solutions. The SDK is stable and broadly used across the provider ecosystem.

For new provider development it is recommended to investigate [`terraform-plugin-framework`](https://github.com/hashicorp/terraform-plugin-framework), which is a reimagined provider SDK that supports additional capabilities. Refer to the [Which SDK Should I Use?](https://terraform.io/docs/plugin/which-sdk.html) documentation for more information about differences between SDKs.

Terraform itself is a tool for building, changing, and versioning infrastructure safely and efficiently. You can find more about Terraform on its [website](https://www.terraform.io) and [its GitHub repository](https://github.com/hashicorp/terraform).

## Terraform CLI Compatibility

Terraform 0.12.0 or later is needed for version 2.0.0 and later of the Plugin SDK.

When running provider tests, Terraform 0.12.26 or later is needed for version 2.0.0 and later of the Plugin SDK. Users can still use any version after 0.12.0.

## Go Compatibility

This project follows the [support policy](https://golang.org/doc/devel/release.html#policy) of Go as its support policy. The two latest major releases of Go are supported by the project.

Currently, that means Go **1.17** or later must be used when including this project as a dependency.

## Getting Started

See the [Call APIs with Terraform Providers](https://learn.hashicorp.com/collections/terraform/providers) guide on [learn.hashicorp.com](https://learn.hashicorp.com) for a guided tour of provider development.

## Documentation

See [Extending Terraform](https://www.terraform.io/docs/extend/index.html) section on the website.

## Scope (Providers VS Core)

### Terraform Core

 - acts as gRPC _client_
 - interacts with the user
 - parses (HCL/JSON) configuration
 - manages state as whole, asks **Provider(s)** to mutate provider-specific parts of state
 - handles backends & provisioners
 - handles inputs, outputs, modules, and functions
 - discovers **Provider(s)** and their versions per configuration
 - manages **Provider(s)** lifecycle (i.e. spins up & tears down provider process)
 - passes relevant parts of parsed (valid JSON/HCL) and interpolated configuration to **Provider(s)**
 - decides ordering of (Create, Read, Update, Delete) operations on resources and data sources
 - ...

### Terraform Provider (via this SDK)

 - acts as gRPC _server_
 - executes any domain-specific logic based on received parsed configuration
   - (Create, Read, Update, Delete, Import, Validate) a Resource
   - Read a Data Source
 - tests domain-specific logic via provided acceptance test framework
 - provides **Core** updated state of a resource or data source and/or appropriate feedback in the form of validation or other errors

## Migrating to SDK v1 from built-in SDK

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/guides/v1-upgrade-guide.html) of the website.

## Migrating to SDK v2 from SDK v1

Migrating to the v2 release of the SDK is covered in the [v2 Upgrade Guide](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html) of the website.

## Versioning

The Terraform Plugin SDK is a [Go module](https://github.com/golang/go/wiki/Modules) versioned using [semantic versioning](https://semver.org/). See [SUPPORT.md](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/SUPPORT.md) for information on our support policies.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/LICENSE)
