// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource_test

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ExampleTestCheckTypeSetElemAttr() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_set_attribute = ["value1", "value2", "value3"]
	//     }
	//
	// The following TestCheckTypeSetElemAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckTypeSetElemAttr("example_thing.test", "example_set_attribute.*", "value1")
	resource.TestCheckTypeSetElemAttr("example_thing.test", "example_set_attribute.*", "value2")
	resource.TestCheckTypeSetElemAttr("example_thing.test", "example_set_attribute.*", "value3")
}

func ExampleTestCheckTypeSetElemNestedAttrs() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_set_block {
	//         key1 = "value1a"
	//         key2 = "value2a"
	//         key3 = "value3a"
	//       }
	//
	//       example_set_block {
	//         key1 = "value1b"
	//         key2 = "value2b"
	//         key3 = "value3b"
	//       }
	//     }
	//
	// The following TestCheckTypeSetElemNestedAttrs can be written to assert
	// against the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckTypeSetElemNestedAttrs(
		"example_thing.test",
		"example_set_block.*",
		map[string]string{
			"key1": "value1a",
			"key2": "value2a",
			"key3": "value3a",
		},
	)
	resource.TestCheckTypeSetElemNestedAttrs(
		"example_thing.test",
		"example_set_block.*",
		map[string]string{
			"key1": "value1b",
			"key2": "value2b",
			"key3": "value3b",
		},
	)
}

func ExampleTestCheckTypeSetElemAttrPair() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     data "example_lookup" "test" {
	//       example_string_attribute = "test-value"
	//     }
	//
	//     resource "example_thing" "test" {
	//       example_set_attribute = [
	//         data.example_lookup.test.example_string_attribute,
	//         "another-test-value",
	//       ]
	//     }
	//
	// The following TestCheckTypeSetElemAttrPair can be written to assert
	// against the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckTypeSetElemAttrPair(
		"example_thing.test",
		"example_set_attribute.*",
		"data.example_lookup.test",
		"example_string_attribute",
	)
}

func ExampleTestMatchTypeSetElemNestedAttrs() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_set_block {
	//         key1 = "value1a"
	//         key2 = "value2a"
	//         key3 = "value3a"
	//       }
	//
	//       example_set_block {
	//         key1 = "value1b"
	//         key2 = "value2b"
	//         key3 = "value3b"
	//       }
	//     }
	//
	// The following TestMatchTypeSetElemNestedAttrs can be written to assert
	// against the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestMatchTypeSetElemNestedAttrs(
		"example_thing.test",
		"example_set_block.*",
		map[string]*regexp.Regexp{
			"key1": regexp.MustCompile(`1a$`),
			"key2": regexp.MustCompile(`2a$`),
			"key3": regexp.MustCompile(`3a$`),
		},
	)
	resource.TestMatchTypeSetElemNestedAttrs(
		"example_thing.test",
		"example_set_block.*",
		map[string]*regexp.Regexp{
			"key1": regexp.MustCompile(`1b$`),
			"key2": regexp.MustCompile(`2b$`),
			"key3": regexp.MustCompile(`3b$`),
		},
	)
}
