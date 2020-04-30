// Package acctest provides the ability to use the binary test driver. The binary
// test driver allows you to run your acceptance tests with a binary of Terraform.
// This is currently the only mechanism for driving tests. It provides a realistic testing
// experience and matrix testing against multiple versions of Terraform CLI,
// as long as they are >= 0.12.0
//
// The driver must be enabled by initialising the test helper in your TestMain
// function in all provider packages that run acceptance tests. Most providers have only
// one package.
//
// After importing this package, you must define a TestMain and have the following:
//
//   func TestMain(m *testing.M) {
//     acctest.UseBinaryDriver("provider_name", Provider)
//     resource.TestMain(m)
//   }
//
// Where `Provider` is the function that returns the instance of a configured `*schema.Provider`
// Some providers already have a TestMain defined, usually for the purpose of enabling test
// sweepers. These additional occurrences should be removed.
//
// It is no longer necessary to import other Terraform providers as Go modules: these
// imports should be removed.
package acctest
