# Terraform Plugin SDK

This SDK enables building Terraform plugin which allows Terraform's users to manage existing and popular service providers as well as custom in-house solutions.

Terraform itself is a tool for building, changing, and versioning infrastructure safely and efficiently. You can find more about Terraform on its [website](https://www.terraform.io) and [its GitHub repository](https://github.com/hashicorp/terraform).

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

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.terraform.io/docs/extend/plugin-sdk.html) of the website.

## Versioning

The Terraform Plugin SDK is a [Go module](https://github.com/golang/go/wiki/Modules) versioned using [semantic versioning](https://semver.org/).

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-sdk/blob/master/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-sdk/blob/master/LICENSE)
