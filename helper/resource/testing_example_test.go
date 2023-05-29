// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource_test

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ExampleComposeAggregateTestCheckFunc() {
	// This function is typically implemented in a TestStep type Check field.
	// Any TestCheckFunc and number of TestCheckFunc may be used within the
	// function parameters. Any errors are combined and displayed together.
	resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("example_thing.test", "example_attribute1", "one"),
		resource.TestCheckResourceAttr("example_thing.test", "example_attribute2", "two"),
	)
}

func ExampleTestCheckNoResourceAttr() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_string_attribute = "test-value"
	//     }
	//
	// The following TestCheckNoResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckNoResourceAttr("example_thing.test", "non_existent_attribute")
}

func ExampleTestCheckResourceAttr_typeBool() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_bool_attribute = true
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttr("example_thing.test", "example_bool_attribute", "true")
}

func ExampleTestCheckResourceAttr_typeFloat() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_float_attribute = 1.2
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttr("example_thing.test", "example_float_attribute", "1.2")
}

func ExampleTestCheckResourceAttr_typeInt() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_int_attribute = 123
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttr("example_thing.test", "example_int_attribute", "123")
}

func ExampleTestCheckResourceAttr_typeListAttribute() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_list_attribute = ["value1", "value2", "value3"]
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.

	// Verify the list attribute contains 3 and only 3 elements
	resource.TestCheckResourceAttr("example_thing.test", "example_list_attribute.#", "3")

	// Verify each list attribute element value
	resource.TestCheckResourceAttr("example_thing.test", "example_list_attribute.0", "value1")
	resource.TestCheckResourceAttr("example_thing.test", "example_list_attribute.1", "value2")
	resource.TestCheckResourceAttr("example_thing.test", "example_list_attribute.2", "value3")
}

func ExampleTestCheckResourceAttr_typeListBlock() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_list_block {
	//         example_string_attribute = "test-nested-value"
	//       }
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.

	// Verify the list block contains 1 and only 1 definition
	resource.TestCheckResourceAttr("example_thing.test", "example_list_block.#", "1")

	// Verify a first list block attribute value
	resource.TestCheckResourceAttr("example_thing.test", "example_list_block.0.example_string_attribute", "test-nested-value")
}

func ExampleTestCheckResourceAttr_typeMap() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_map_attribute = {
	//         key1 = "value1"
	//         key2 = "value2"
	//         key3 = "value3"
	//       }
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.

	// Verify the map attribute contains 3 and only 3 elements
	resource.TestCheckResourceAttr("example_thing.test", "example_map_attribute.%", "3")

	// Verify each map attribute element value
	resource.TestCheckResourceAttr("example_thing.test", "example_map_attribute.key1", "value1")
	resource.TestCheckResourceAttr("example_thing.test", "example_map_attribute.key2", "value2")
	resource.TestCheckResourceAttr("example_thing.test", "example_map_attribute.key3", "value3")
}

func ExampleTestCheckResourceAttr_typeString() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_string_attribute = "test-value"
	//     }
	//
	// The following TestCheckResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttr("example_thing.test", "example_string_attribute", "test-value")
}

func ExampleTestCheckResourceAttrWith_typeString() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_string_attribute = "Very long string..."
	//     }
	//
	// The following TestCheckResourceAttrWith can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.

	// Verify the attribute value string length is above 1000
	resource.TestCheckResourceAttrWith("example_thing.test", "example_string_attribute", func(value string) error {
		if len(value) <= 1000 {
			return fmt.Errorf("should be longer than 1000 characters")
		}
		return nil
	})
}

func ExampleTestCheckResourceAttrWith_typeInt() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_int_attribute = 10
	//     }
	//
	// The following TestCheckResourceAttrWith can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.

	// Verify the attribute value is an integer, and it's between 5 (included) and 20 (excluded)
	resource.TestCheckResourceAttrWith("example_thing.test", "example_string_attribute", func(value string) error {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}

		if valueInt < 5 && valueInt >= 20 {
			return fmt.Errorf("should be between 5 and 20")
		}
		return nil
	})
}

func ExampleTestCheckResourceAttrPair() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test1" {
	//       example_string_attribute = "test-value"
	//     }
	//
	//     resource "example_thing" "test2" {
	//       example_string_attribute = example_thing.test1.example_string_attribute
	//     }
	//
	// The following TestCheckResourceAttrPair can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttrPair(
		"example_thing.test1",
		"example_string_attribute",
		"example_thing.test2",
		"example_string_attribute",
	)
}

func ExampleTestCheckResourceAttrSet() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_string_attribute = "test-value"
	//     }
	//
	// The following TestCheckResourceAttrSet can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestCheckResourceAttrSet("example_thing.test", "example_string_attribute")
}

func ExampleTestMatchResourceAttr() {
	// This function is typically implemented in a TestStep type Check field,
	// wrapped with ComposeAggregateTestCheckFunc to combine results from
	// multiple checks.
	//
	// Given the following example configuration:
	//
	//     resource "example_thing" "test" {
	//       example_string_attribute = "test-value"
	//     }
	//
	// The following TestMatchResourceAttr can be written to assert against
	// the expected state values.
	//
	// NOTE: State value checking is only necessary for Computed attributes,
	//       as the testing framework will automatically return test failures
	//       for configured attributes that mismatch the saved state, however
	//       this configuration and test is shown for illustrative purposes.
	resource.TestMatchResourceAttr("example_thing.test", "example_string_attribute", regexp.MustCompile(`^test-`))
}
